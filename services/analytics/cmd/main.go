package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	analyticsv1 "rtb-platform/pb/analytics/v1"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/services/analytics/internal/adapters/eventstore"
	"rtb-platform/services/analytics/internal/domain"
	"rtb-platform/services/analytics/internal/server"

	"google.golang.org/grpc"
)

type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Log     LogConfig     `yaml:"log"`
	Metrics MetricsConfig `yaml:"metrics"`
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

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "analytics"))
	appLogger.Info("starting analytics service")

	if err := metrics.Init(context.Background(), "analytics", cfg.Metrics.UseOTLP); err != nil {
		appLogger.Error("metrics init failed", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	store := eventstore.NewMemoryStore()
	// Можно заполнить тестовыми событиями для проверки GetReport (опционально)

	reportSvc := domain.NewReportService(store)
	forecastSvc := domain.NewForecastService()
	factorSvc := domain.NewFactorService()

	grpcServer := grpc.NewServer()
	analyticsSrv := server.NewAnalyticsServer(reportSvc, forecastSvc, factorSvc, appLogger)
	analyticsv1.RegisterAnalyticsServiceServer(grpcServer, analyticsSrv)

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
