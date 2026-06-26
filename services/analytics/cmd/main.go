package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	analyticsv1 "rtb-platform/pb/analytics/v1"
	"rtb-platform/pkg/config"
	"rtb-platform/pkg/logger"
	"rtb-platform/pkg/metrics"
	"rtb-platform/pkg/shutdown"
	"rtb-platform/services/analytics/internal/adapters/clickhouse"
	"rtb-platform/services/analytics/internal/adapters/eventstore"
	"rtb-platform/services/analytics/internal/domain"
	"rtb-platform/services/analytics/internal/server"

	"google.golang.org/grpc"
)

type AppConfig struct {
	Server     ServerConfig     `yaml:"server"`
	ClickHouse ClickHouseConfig `yaml:"clickhouse"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Log        LogConfig        `yaml:"log"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

type KafkaConfig struct {
	Brokers string `yaml:"brokers" env:"KAFKA_BROKERS"`
	Topic   string `yaml:"topic" env:"KAFKA_TOPIC"`
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

type ClickHouseConfig struct {
	DSN string `yaml:"use_otlp"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("configs/dev.yaml"), config.WithEnvPrefix("")); err != nil {
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

	var store domain.EventStore
	if cfg.ClickHouse.DSN != "" {
		ctx := context.Background()
		chStore, err := clickhouse.NewCHStore(ctx, cfg.ClickHouse.DSN)
		if err != nil {
			appLogger.Error("failed to connect to ClickHouse, falling back to memory", "error", err)
			store = eventstore.NewMemoryStore()
		} else {
			store = chStore
			defer chStore.Close()
			appLogger.Info("using ClickHouse event store")
		}
	} else {
		store = eventstore.NewMemoryStore()
		appLogger.Info("using in-memory event store")
	}

	// Генерация тестовых событий для ClickHouse (если используется ClickHouse)
	// Генерация тестовых событий (для любого хранилища)
	// Генерация тестовых событий
	now := time.Now()
	for daysAgo := 7; daysAgo >= 0; daysAgo-- {
		eventDate := now.AddDate(0, 0, -daysAgo)
		for hour := 0; hour < 24; hour += 6 {
			for _, camp := range []uint32{1001, 1002} {
				price := int64(150)
				if camp == 1002 {
					price = 200
				}
				store.Add(domain.Event{
					BidID:         fmt.Sprintf("bid-%d-%d-%d", camp, daysAgo, hour),
					CampaignID:    camp,
					DeviceID:      fmt.Sprintf("device-%d", (daysAgo*hour)%5),
					PriceCents:    price,
					Win:           true,
					LtvScore:      0.5 + float64(hour%3)*0.2,
					GeoFactor:     0.8 + float64(daysAgo%3)*0.1,
					ImpressionVal: float64(price) * 0.8,
					Timestamp:     eventDate.Add(time.Duration(hour) * time.Hour),
				})
			}
		}
	}
	fmt.Println("Inserted test events into store")

	// Kafka consumer (если используется ClickHouse, то подключаем Kafka для наполнения)
	if cfg.Kafka.Brokers != "" {
		kafkaStore, ok := store.(*clickhouse.CHStore) // если у нас ClickHouse, то можем пробросить Kafka напрямую
		if ok {
			consumer := eventstore.NewKafkaConsumer(
				strings.Split(cfg.Kafka.Brokers, ","),
				cfg.Kafka.Topic,
				"analytics-consumer-group",
				kafkaStore,
				appLogger,
			)
			consumer.Start(context.Background())
			defer consumer.Close()
			appLogger.Info("Kafka consumer started")
		} else {
			// fallback: используем MemoryStore или другой store, тоже можно слушать Kafka
			consumer := eventstore.NewKafkaConsumer(
				strings.Split(cfg.Kafka.Brokers, ","),
				cfg.Kafka.Topic,
				"analytics-consumer-group",
				store,
				appLogger,
			)
			consumer.Start(context.Background())
			defer consumer.Close()
			appLogger.Info("Kafka consumer started (non-ClickHouse store)")
		}
	}

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
