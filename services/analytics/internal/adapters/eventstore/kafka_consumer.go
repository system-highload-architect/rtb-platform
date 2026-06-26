package eventstore

import (
	"context"
	"encoding/json"
	"log/slog"

	"rtb-platform/services/analytics/internal/domain"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer читает события из Kafka и сохраняет их в EventStore.
type KafkaConsumer struct {
	reader *kafka.Reader
	store  domain.EventStore
	logger *slog.Logger
}

// NewKafkaConsumer создаёт consumer group.
func NewKafkaConsumer(brokers []string, topic, groupID string, store domain.EventStore, logger *slog.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	return &KafkaConsumer{
		reader: reader,
		store:  store,
		logger: logger,
	}
}

// Start запускает бесконечный цикл чтения.
func (kc *KafkaConsumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := kc.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				kc.logger.Error("kafka fetch error", "error", err)
				continue
			}

			var event domain.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				kc.logger.Error("kafka unmarshal error", "error", err)
				// всё равно коммитим, чтобы не застревать
				kc.reader.CommitMessages(ctx, msg)
				continue
			}

			// Добавляем событие в хранилище (ClickHouse)
			kc.store.Add(event)

			// Коммитим оффсет
			if err := kc.reader.CommitMessages(ctx, msg); err != nil {
				kc.logger.Error("kafka commit error", "error", err)
			}
		}
	}()
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
