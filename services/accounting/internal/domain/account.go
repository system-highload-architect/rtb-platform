package domain

import (
	"errors"
	"sync"

	"rtb-platform/pkg/fixedpoint"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrCampaignNotFound  = errors.New("campaign not found")
)

type BalanceStore interface {
	Get(campaignID string) (fixedpoint.Money, bool)
	Debit(campaignID string, amount fixedpoint.Money) error
	Set(campaignID string, balance fixedpoint.Money)
}

type inmemStore struct {
	mu     sync.RWMutex
	ledger map[string]fixedpoint.Money
}

func NewInmemStore() BalanceStore {
	return &inmemStore{ledger: make(map[string]fixedpoint.Money)}
}

func (s *inmemStore) Get(campaignID string) (fixedpoint.Money, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bal, ok := s.ledger[campaignID]
	return bal, ok
}

func (s *inmemStore) Debit(campaignID string, amount fixedpoint.Money) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bal, ok := s.ledger[campaignID]
	if !ok {
		return ErrCampaignNotFound
	}
	newBal, err := bal.Sub(amount)
	if err != nil {
		return err
	}
	if newBal.Sign() < 0 {
		return ErrInsufficientFunds
	}
	s.ledger[campaignID] = newBal
	return nil
}

func (s *inmemStore) Set(campaignID string, balance fixedpoint.Money) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ledger[campaignID] = balance
}
