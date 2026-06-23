package kafka

import (
	"context"
	"rtb-platform/services/auction/internal/domain"
	"rtb-platform/services/auction/internal/ports"
)

type dummyPublisher struct{}

// NewDummyPublisher создаёт заглушку публикатора событий.
func NewDummyPublisher() ports.EventPublisher {
	return &dummyPublisher{}
}

func (d *dummyPublisher) PublishBidEvent(ctx context.Context, event domain.BidEvent) error {
	return nil
}
