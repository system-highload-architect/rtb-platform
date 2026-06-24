package domain

import (
	"context"

	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/pkg/idempotent"
)

type Service struct {
	store      BalanceStore
	idempotent *idempotent.Store
}

func NewService(store BalanceStore, idemp *idempotent.Store) *Service {
	return &Service{store: store, idempotent: idemp}
}

func (s *Service) Debit(ctx context.Context, campaignID string, amount fixedpoint.Money, bidID string) error {
	if bidID != "" {
		if !s.idempotent.Check(bidID) {
			return nil
		}
	}
	return s.store.Debit(campaignID, amount)
}

func (s *Service) GetBalance(ctx context.Context, campaignID string) (fixedpoint.Money, error) {
	bal, ok := s.store.Get(campaignID)
	if !ok {
		return 0, ErrCampaignNotFound
	}
	return bal, nil
}
