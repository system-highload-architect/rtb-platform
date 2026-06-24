package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	accountingv1 "rtb-platform/pb/accounting/v1"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/services/accounting/internal/domain"
	"rtb-platform/services/accounting/internal/server"

	"google.golang.org/grpc"
)

type AppConfig struct {
	Server      ServerConfig      `yaml:"server"`
	Log         LogConfig         `yaml:"log"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	Idempotency IdempotencyConfig `yaml:"idempotency"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
type MetricsConfig struct {
	UseOTLP bool `yaml:"use_otlp"`
}
type IdempotencyConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "accounting"))
	appLogger.Info("starting accounting service")

	if err := metrics.Init(context.Background(), "accounting", cfg.Metrics.UseOTLP); err != nil {
		appLogger.Error("metrics init failed", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	// In-memory хранилище
	store := domain.NewInmemStore()

	// Добавим тестовые кампании
	store.Set("campaign-1", fixedpoint.Money(10000)) // 100.00 руб
	store.Set("campaign-2", fixedpoint.Money(5000))  //  50.00 руб

	idempStore := idempotent.NewStore(cfg.Idempotency.TTL)

	svc := domain.NewService(store, idempStore)

	grpcServer := grpc.NewServer()
	accountingSrv := server.NewAccountingServer(svc, appLogger)
	accountingv1.RegisterAccountingServiceServer(grpcServer, accountingSrv)

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
	shutdownMgr.Wait()
}
