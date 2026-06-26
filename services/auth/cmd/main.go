package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	authv1 "rtb-platform/pb/auth/v1"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/services/auth/internal/domain"
	"rtb-platform/services/auth/internal/server"

	"google.golang.org/grpc"
)

type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	JWT     JWTConfig     `yaml:"jwt"`
	Log     LogConfig     `yaml:"log"`
	Metrics MetricsConfig `yaml:"metrics"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	AccessTTL  time.Duration `yaml:"access_ttl"`
	RefreshTTL time.Duration `yaml:"refresh_ttl"`
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
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml"), config.WithEnvPrefix("")); err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Log.Level, cfg.Log.Format, slog.String("service", "auth"))
	appLogger.Info("starting auth service")

	if err := metrics.Init(context.Background(), "auth", cfg.Metrics.UseOTLP); err != nil {
		appLogger.Error("metrics init failed", "error", err)
		os.Exit(1)
	}
	defer metrics.Shutdown(context.Background())

	store := domain.NewInmemStore()
	// Опционально – создать администратора по умолчанию
	store.Create(domain.User{
		ID:           "admin-1",
		Email:        "admin@rtb-platform.local",
		PasswordHash: fmt.Sprintf("%x", sha256.Sum256([]byte("Admin123!"))),
		Role:         "admin",
		CreatedAt:    time.Now(),
	})

	grpcServer := grpc.NewServer()
	authSrv := server.NewAuthServer(store, cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, appLogger)
	authv1.RegisterAuthServiceServer(grpcServer, authSrv)

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
