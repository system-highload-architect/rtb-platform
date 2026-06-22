package idempotent

import (
	"time"

	"rtb-platform/pkg/timedcache"
)

// Store гарантирует, что операция с заданным ключом выполнится только один раз.
type Store struct {
	cache *timedcache.Cache[string, struct{}]
}

// NewStore создаёт хранилище идемпотентности с временем жизни ключа ttl.
func NewStore(ttl time.Duration) *Store {
	return &Store{
		cache: timedcache.New[string, struct{}](ttl),
	}
}

// Check возвращает true, если ключ ещё не использовался, и регистрирует его.
// Если ключ уже есть, возвращает false — операцию нужно отклонить.
func (s *Store) Check(key string) bool {
	if _, ok := s.cache.Get(key); ok {
		return false
	}
	s.cache.Set(key, struct{}{})
	return true
}

// Stop завершает работу кэша, останавливая демон и финализаторы.
func (s *Store) Stop() {
	s.cache.Stop()
}
