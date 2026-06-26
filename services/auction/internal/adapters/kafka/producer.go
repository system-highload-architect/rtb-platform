package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"rtb-platform/services/auction/internal/domain"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

func NewProducer(brokers []string, topic string, logger *slog.Logger) *Producer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: w, logger: logger}
}

func (p *Producer) PublishBidEvent(ctx context.Context, event domain.BidEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.BidID),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
