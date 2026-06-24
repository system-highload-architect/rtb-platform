package eventstore

import (
	"context"
	"sync"
	"time"

	"rtb-platform/services/analytics/internal/domain"
)

// MemoryStore – простейшее in-memory хранилище событий (временное).
type MemoryStore struct {
	mu     sync.RWMutex
	events []domain.Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Add(event domain.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *MemoryStore) Query(ctx context.Context, start, end time.Time) []domain.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []domain.Event
	for _, e := range s.events {
		if (e.Timestamp.Equal(start) || e.Timestamp.After(start)) && !e.Timestamp.After(end) {
			result = append(result, e)
		}
	}
	return result
}
