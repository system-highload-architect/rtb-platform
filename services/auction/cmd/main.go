package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	accountingv1 "rtb-platform/pb/accounting/v1"
	auctionv1 "rtb-platform/pb/auction/v1"

	"rtb-platform/pkg/breaker"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/experiment"
	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/pkg/geoip"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/sampler"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/pkg/timedcache"

	"rtb-platform/services/auction/internal/adapters/aerospike"
	"rtb-platform/services/auction/internal/adapters/fraud"
	"rtb-platform/services/auction/internal/adapters/geodata"
	"rtb-platform/services/auction/internal/adapters/grpcclient"
	"rtb-platform/services/auction/internal/adapters/kafka"
	"rtb-platform/services/auction/internal/adapters/mongodb"
	"rtb-platform/services/auction/internal/domain"
	"rtb-platform/services/auction/internal/domain/scoring"
	"rtb-platform/services/auction/internal/ports"
	"rtb-platform/services/auction/internal/server"

	as "github.com/aerospike/aerospike-client-go/v7"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AppConfig struct {
	Server      ServerConfig      `yaml:"server"`
	Aerospike   AerospikeConfig   `yaml:"aerospike"`
	Mongo       MongoConfig       `yaml:"mongo"`
	Scoring     ScoringConfig     `yaml:"scoring"`
	Fraud       FraudConfig       `yaml:"fraud"`
	Log         LogConfig         `yaml:"log"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	GeoIP       GeoIPConfig       `yaml:"geoip"`
	Sampler     SamplerConfig     `yaml:"sampler"`
	Experiments ExperimentsConfig `yaml:"experiments"`
	Breaker     BreakerConfig     `yaml:"breaker"`
	GRPC        GRPCConfig        `yaml:"grpc"`
}

type GRPCConfig struct {
	Accounting string `yaml:"accounting" env:"GRPC_ACCOUNTING"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}
type AerospikeConfig struct {
	Hosts     []string `yaml:"hosts"`
	Namespace string   `yaml:"namespace"`
	Set       string   `yaml:"set"`
}
type MongoConfig struct {
	URI string `yaml:"uri"`
	DB  string `yaml:"db"`
}
type ScoringConfig struct {
	LTVCoeffs           []float64 `yaml:"ltv_coeffs"`
	CTR                 float64   `yaml:"ctr"`
	CVR                 float64   `yaml:"cvr"`
	BaseConversionValue float64   `yaml:"base_conversion_value"`
	GeoDecay            float64   `yaml:"geo_decay"`
}
type FraudConfig struct {
	Blacklist []string `yaml:"blacklist"`
}
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
type MetricsConfig struct {
	UseOTLP bool `yaml:"use_otlp"`
}
type GeoIPConfig struct {
	DatabasePath string `yaml:"database_path"`
}
type SamplerConfig struct {
	Rate float64 `yaml:"rate"`
}
type ExperimentsConfig struct {
	Flags map[string]float64 `yaml:"flags"`
}
type BreakerConfig struct {
	AerospikeThreshold int           `yaml:"aerospike_threshold"`
	AerospikeTimeout   time.Duration `yaml:"aerospike_timeout"`
	MongoThreshold     int           `yaml:"mongo_threshold"`
	MongoTimeout       time.Duration `yaml:"mongo_timeout"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml"), config.WithEnvPrefix("")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "auction"))
	appLogger.Info("starting auction service")

	if err := metrics.Init(context.Background(), "auction", cfg.Metrics.UseOTLP); err != nil {
		appLogger.Error("metrics init failed", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	// GeoIP (опционально, если файл существует)
	var geoipDB *geoip.GeoDB
	if cfg.GeoIP.DatabasePath != "" {
		var err error
		geoipDB, err = geoip.New(cfg.GeoIP.DatabasePath)
		if err != nil {
			appLogger.Error("failed to load geoip db, continuing without GeoIP", "error", err)
			geoipDB = nil
		} else {
			defer geoipDB.Close()
		}
	}

	// Sampler
	samplerInstance := sampler.NewSampler(cfg.Sampler.Rate)

	// Experiments
	experiments := experiment.New(cfg.Experiments.Flags)

	// Circuit Breakers
	aeroBreaker := breaker.New("aerospike", cfg.Breaker.AerospikeThreshold, cfg.Breaker.AerospikeTimeout)
	mongoBreaker := breaker.New("mongo", cfg.Breaker.MongoThreshold, cfg.Breaker.MongoTimeout)

	// Инициализация клиента Aerospike (пока можно использовать nil, если Aerospike не запущен)
	var aeroClient *as.Client = nil
	// В реальном коде:
	// clientPolicy := as.NewClientPolicy()
	// aeroClient, err := as.NewClientWithPolicyAndHost(clientPolicy, as.NewHosts(cfg.Aerospike.Hosts...)...)
	// if err != nil { ... }

	userRepo := aerospike.NewUserRepo(aeroClient, cfg.Aerospike.Namespace, cfg.Aerospike.Set, aeroBreaker)

	// Campaign cache (in-memory) campaignCache
	campaignTTL := 24 * time.Hour // чтобы не истекали во время тестов
	var campaignCache ports.CampaignRepo
	if cfg.Mongo.URI != "" {
		ctx := context.Background()
		mongoRepo, err := mongodb.NewMongoRepo(ctx, cfg.Mongo.URI, cfg.Mongo.DB, mongoBreaker)
		if err != nil {
			appLogger.Error("failed to connect to MongoDB, falling back to in-memory cache", "error", err)
			campaignCache = mongodb.NewCachedCampaignRepo(campaignTTL, mongoBreaker)
		} else {
			campaignCache = mongoRepo
			defer mongoRepo.Close()
			mongoRepo.StartAutoRefresh(ctx, 5*time.Minute)
			appLogger.Info("using MongoDB campaign repository")
		}
	} else {
		campaignCache = mongodb.NewCachedCampaignRepo(campaignTTL, mongoBreaker)
		appLogger.Info("using in-memory campaign cache")
	}

	// Временная загрузка тестовых кампаний (уберем после реализации MongoDB)
	testCampaigns := []domain.Campaign{
		{
			ID:           1001,
			BidCents:     150,
			DailyBudget:  fixedpoint.NewFromInt64(10000),
			CreativeURL:  "https://cdn.example.com/creatives/1001.jpg",
			BillboardLat: 55.7600,
			BillboardLng: 37.6200,
		},
		{
			ID:           1002,
			BidCents:     200,
			DailyBudget:  fixedpoint.NewFromInt64(5000),
			CreativeURL:  "https://cdn.example.com/creatives/1002.jpg",
			BillboardLat: 55.7500,
			BillboardLng: 37.6100,
		},
	}
	campaignCache.Update(testCampaigns)

	// Fraud detector
	fraudDetector := fraud.NewInmemDetector(cfg.Fraud.Blacklist)

	// Geo resolver (пока заглушка с пустыми координатами)
	geoResolver := geodata.NewInmemGeoResolver(nil)

	// Kafka producer
	kafkaProducer := kafka.NewProducer([]string{"kafka:29092"}, "bid_events", appLogger)
	defer kafkaProducer.Close()

	// Основной скорер
	scorer := scoring.NewPredictiveScorer(
		cfg.Scoring.LTVCoeffs,
		cfg.Scoring.CTR,
		cfg.Scoring.CVR,
		cfg.Scoring.BaseConversionValue,
		cfg.Scoring.GeoDecay,
	)

	// Альтернативный скорер для A/B-теста (немного другие параметры)
	altScorer := scoring.NewPredictiveScorer(
		cfg.Scoring.LTVCoeffs,
		cfg.Scoring.CTR*1.1,
		cfg.Scoring.CVR*0.9,
		cfg.Scoring.BaseConversionValue,
		cfg.Scoring.GeoDecay,
	)

	idempStore := idempotent.NewStore(5 * time.Minute)

	// gRPC-сервер
	grpcServer := grpc.NewServer()
	// Accounting gRPC client
	accountingConn, err := grpc.NewClient(cfg.GRPC.Accounting, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLogger.Error("cannot connect to accounting", "error", err)
		os.Exit(1)
	}
	defer accountingConn.Close()
	accountingClient := accountingv1.NewAccountingServiceClient(accountingConn)
	accountingPort := grpcclient.NewAccountingPort(accountingClient, appLogger)

	auctionSrv := server.NewAuctionServer(
		userRepo,
		campaignCache,
		fraudDetector,
		geoResolver,
		scorer,
		altScorer,
		kafkaProducer,
		idempStore,
		appLogger,
		geoipDB,
		samplerInstance,
		experiments,
		accountingPort,
	)
	auctionv1.RegisterAuctionServiceServer(grpcServer, auctionSrv)

	// Асинхронный кэш аукционов
	auctionCache := timedcache.New[string, []*auctionv1.BidRequest](
		100*time.Millisecond,
		timedcache.WithFinalizer[string, []*auctionv1.BidRequest](func(key string, requests []*auctionv1.BidRequest) {
			auctionSrv.ProcessBids(context.Background(), key, requests)
		}),
	)
	defer auctionCache.Stop()
	auctionSrv.SetAuctionCache(auctionCache)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		appLogger.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		appLogger.Info("gRPC server listening", "port", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			appLogger.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	shutdownMgr := shutdown.NewManager(30 * time.Second)
	shutdownMgr.Add("grpc", 0, func(ctx context.Context) error {
		grpcServer.GracefulStop()
		return nil
	}, 5*time.Second)
	shutdownMgr.Add("campaign_cache", 1, func(ctx context.Context) error {
		campaignCache.Stop()
		return nil
	}, 3*time.Second)
	shutdownMgr.Wait()
}
