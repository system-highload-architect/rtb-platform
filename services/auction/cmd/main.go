package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	auctionv1 "rtb-platform/pb/auction/v1"
	"rtb-platform/pkg/breaker"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/experiment"
	"rtb-platform/pkg/geoip"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/sampler"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/services/auction/internal/adapters/aerospike"
	"rtb-platform/services/auction/internal/adapters/fraud"
	"rtb-platform/services/auction/internal/adapters/geodata"
	"rtb-platform/services/auction/internal/adapters/kafka"
	"rtb-platform/services/auction/internal/adapters/mongodb"
	"rtb-platform/services/auction/internal/domain/scoring"
	"rtb-platform/services/auction/internal/server"

	as "github.com/aerospike/aerospike-client-go/v7"
	"google.golang.org/grpc"
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
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
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

	// GeoIP
	geoipDB, err := geoip.New(cfg.GeoIP.DatabasePath)
	if err != nil {
		appLogger.Error("failed to load geoip db", "error", err)
		os.Exit(1)
	}
	defer geoipDB.Close()

	// Sampler
	samplerInstance := sampler.NewSampler(cfg.Sampler.Rate)

	// Experiments
	experiments := experiment.New(cfg.Experiments.Flags)

	// Circuit Breakers
	aeroBreaker := breaker.New("aerospike", cfg.Breaker.AerospikeThreshold, cfg.Breaker.AerospikeTimeout)
	mongoBreaker := breaker.New("mongo", cfg.Breaker.MongoThreshold, cfg.Breaker.MongoTimeout)

	// Aerospike client (пока nil)
	var aeroClient *as.Client = nil
	userRepo := aerospike.NewUserRepo(aeroClient, cfg.Aerospike.Namespace, cfg.Aerospike.Set, aeroBreaker)

	// Campaign cache
	campaignTTL := 5 * time.Minute
	campaignCache := mongodb.NewCachedCampaignRepo(campaignTTL, mongoBreaker)

	// Fraud detector
	fraudDetector := fraud.NewInmemDetector(cfg.Fraud.Blacklist)

	// Geo resolver
	geoResolver := geodata.NewInmemGeoResolver(nil)

	// Publisher
	publisher := kafka.NewDummyPublisher()

	// Основной скорер
	scorer := scoring.NewPredictiveScorer(
		cfg.Scoring.LTVCoeffs,
		cfg.Scoring.CTR,
		cfg.Scoring.CVR,
		cfg.Scoring.BaseConversionValue,
		cfg.Scoring.GeoDecay,
	)

	// Альтернативный скорер (для A/B) – можно другие коэффициенты
	altScorer := scoring.NewPredictiveScorer(
		cfg.Scoring.LTVCoeffs,
		cfg.Scoring.CTR*1.1,
		cfg.Scoring.CVR*0.9,
		cfg.Scoring.BaseConversionValue,
		cfg.Scoring.GeoDecay,
	)

	idempStore := idempotent.NewStore(5 * time.Minute)

	grpcServer := grpc.NewServer()
	auctionSrv := server.NewAuctionServer(
		userRepo,
		campaignCache,
		fraudDetector,
		geoResolver,
		scorer,
		altScorer,
		publisher,
		idempStore,
		appLogger,
		geoipDB,
		samplerInstance,
		experiments,
	)
	auctionv1.RegisterAuctionServiceServer(grpcServer, auctionSrv)

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
