package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	accountingv1 "rtb-platform/pb/accounting/v1"
	analyticsv1 "rtb-platform/pb/analytics/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
	authv1 "rtb-platform/pb/auth/v1"

	"rtb-platform/pkg/config"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/ratelimit"
	"rtb-platform/pkg/shutdown"

	"rtb-platform/services/gateway/internal/adapters/grpcclient"
	"rtb-platform/services/gateway/internal/domain"
	"rtb-platform/services/gateway/internal/handler"
	"rtb-platform/services/gateway/internal/middleware"
	"rtb-platform/services/gateway/internal/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AppConfig — полная конфигурация сервиса gateway.
type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Security SecurityConfig `yaml:"security"`
	Log      LogConfig      `yaml:"log"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	Port         int           `yaml:"port" env:"SERVER_PORT"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type GRPCConfig struct {
	Auction    string `yaml:"auction" env:"GRPC_AUCTION"`
	Accounting string `yaml:"accounting" env:"GRPC_ACCOUNTING"`
	Analytics  string `yaml:"analytics" env:"GRPC_ANALYTICS"`
	Auth       string `yaml:"auth" env:"GRPC_AUTH"`
}

type SecurityConfig struct {
	RateLimit      float64       `yaml:"rate_limit"`
	RateBurst      float64       `yaml:"rate_burst"`
	IdempotencyTTL time.Duration `yaml:"idempotency_ttl"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type MetricsConfig struct {
	UseOTLP bool `yaml:"use_otlp"`
}

func main() {
	// 1. Загрузка конфигурации
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml"), config.WithEnvPrefix("")); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	// 2. Инициализация логгера
	logger := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "gateway"))
	logger.Info("starting gateway")

	// 3. Инициализация метрик
	if err := metrics.Init(context.Background(), "gateway", cfg.Metrics.UseOTLP); err != nil {
		logger.Error("failed to init metrics", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	// 4. Создание gRPC клиентов
	auctionConn, err := grpc.NewClient(cfg.GRPC.Auction, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("cannot connect to auction", "error", err)
		os.Exit(1)
	}
	defer auctionConn.Close()
	auctionClient := auctionv1.NewAuctionServiceClient(auctionConn)

	accountingConn, err := grpc.NewClient(cfg.GRPC.Accounting, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("cannot connect to accounting", "error", err)
		os.Exit(1)
	}
	defer accountingConn.Close()
	accountingClient := accountingv1.NewAccountingServiceClient(accountingConn)

	analyticsConn, err := grpc.NewClient(cfg.GRPC.Analytics, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("cannot connect to analytics", "error", err)
		os.Exit(1)
	}
	defer analyticsConn.Close()
	analyticsClient := analyticsv1.NewAnalyticsServiceClient(analyticsConn)

	// 5. Создание портов (интерфейсов)
	auctionPort := grpcclient.NewAuctionPort(auctionClient, logger)
	accountingPort := grpcclient.NewAccountingPort(accountingClient, logger)
	analyticsPort := grpcclient.NewAnalyticsPort(analyticsClient, logger)

	// После создания analyticsClient
	authConn, err := grpc.NewClient(cfg.GRPC.Auth, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("cannot connect to auth", "error", err)
		os.Exit(1)
	}
	defer authConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)
	authPort := grpcclient.NewAuthPort(authClient, logger)

	// Создаём middleware аутентификации
	authMiddleware := middleware.NewAuthMiddleware(authPort)

	// 6. Доменная логика (обработчики JSON-RPC)
	jsonRPCService := domain.NewJSONRPCService(auctionPort, accountingPort, analyticsPort, authPort, logger)

	// 7. Инфраструктурные компоненты
	limiter := ratelimit.NewLimiter(cfg.Security.RateLimit, cfg.Security.RateBurst)
	idempotentStore := idempotent.NewStore(cfg.Security.IdempotencyTTL)

	// Создаём обработчики аналитики
	analyticsHandler := handler.NewAnalyticsHandler(analyticsPort)

	rateLimitMW := middleware.NewRateLimitMiddleware(limiter)
	idempotentMW := middleware.NewIdempotentMiddleware(idempotentStore)
	appsecMW := middleware.NewAppsecMiddleware([]string{"localhost", "127.0.0.1", "localhost:8080"})

	// Сервер
	srv := server.NewHTTPServer(
		server.WithPort(cfg.Server.Port),
		server.WithLogger(logger),
		server.WithRateLimitMiddleware(rateLimitMW),
		server.WithIdempotentMiddleware(idempotentMW),
		server.WithAppsecMiddleware(appsecMW),
		server.WithJSONRPCService(jsonRPCService),
		server.WithAnalyticsHandler(analyticsPort),        // для Excel
		server.WithAnalyticsRESTHandler(analyticsHandler), // для REST API
		server.WithAuthMiddleware(authMiddleware),
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithIdleTimeout(cfg.Server.IdleTimeout),
		server.WithStaticDir("web/dist"),
	)

	srv.Handle("/metrics", metrics.Handler())

	// 9. Graceful shutdown
	shutdownMgr := shutdown.NewManager(30 * time.Second)
	shutdownMgr.Add("http_server", 0, func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	}, 5*time.Second)
	shutdownMgr.Add("idempotent_cache", 1, func(ctx context.Context) error {
		idempotentStore.Stop()
		return nil
	}, 1*time.Second)
	shutdownMgr.Add("rate_limiter", 1, func(ctx context.Context) error {
		limiter.Stop()
		return nil
	}, 1*time.Second)

	go func() {
		logger.Info("gateway listening", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	shutdownMgr.Wait()
}
