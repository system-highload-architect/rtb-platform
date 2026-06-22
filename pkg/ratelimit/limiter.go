package ratelimit

import (
	"sync"
	"time"
)

// Limiter хранит ведра для разных ключей (IP, user ID) и периодически чистит старые.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
	rate    float64
	burst   float64
	stopCh  chan struct{}
}

// NewLimiter создаёт лимитер и запускает фоновую очистку неиспользуемых вёдер.
func NewLimiter(rate, burst float64) *Limiter {
	l := &Limiter{
		buckets: make(map[string]*TokenBucket),
		rate:    rate,
		burst:   burst,
		stopCh:  make(chan struct{}),
	}
	go l.cleanup(1 * time.Minute)
	return l
}

// Allow проверяет, можно ли пропустить запрос для данного ключа.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = NewTokenBucket(l.rate, l.burst)
		l.buckets[key] = bucket
	}
	l.mu.Unlock()
	return bucket.Allow()
}

// Stop завершает фоновую очистку.
func (l *Limiter) Stop() {
	close(l.stopCh)
}

// cleanup удаляет ведра, к которым не было обращений более 2 минут.
func (l *Limiter) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.mu.Lock()
			for key, bucket := range l.buckets {
				lastAccess := bucket.lastAccess.Load()
				if time.Since(time.Unix(lastAccess, 0)) > 2*time.Minute {
					delete(l.buckets, key)
				}
			}
			l.mu.Unlock()
		}
	}
}
