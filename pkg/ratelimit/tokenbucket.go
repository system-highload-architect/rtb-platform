package ratelimit

import (
	"sync/atomic"
	"time"
)

const (
	tokenScale = 1_000_000 // токены как целые, 1.0 = 1_000_000
)

// TokenBucket — lock-free реализация Token Bucket.
type TokenBucket struct {
	rate       float64 // токенов в секунду (константа после создания)
	burst      float64
	burstInt   int64
	rateInt    int64         // токенов в секунду * tokenScale
	state      atomic.Uint64 // упакованное состояние: старшие 32 бита — время (сек), младшие — количество токенов * scale
	lastAccess atomic.Int64  // Unix секунды последнего Allow (для очистки)
}

func NewTokenBucket(rate, burst float64) *TokenBucket {
	tb := &TokenBucket{
		rate:     rate,
		burst:    burst,
		burstInt: int64(burst * tokenScale),
		rateInt:  int64(rate * tokenScale),
	}
	// Начальное состояние: полное ведро, время "сейчас"
	tb.state.Store(packState(time.Now().Unix(), int64(burst*tokenScale)))
	tb.lastAccess.Store(time.Now().Unix())
	return tb
}

// packState упаковывает время (секунды) и токены (int64) в одно uint64.
func packState(nowSec int64, tokens int64) uint64 {
	return uint64(nowSec)<<32 | uint64(tokens)
}

// Allow пытается взять один токен. true — разрешено.
func (tb *TokenBucket) Allow() bool {
	now := time.Now().Unix()
	rate := tb.rateInt
	burst := tb.burstInt

	// Обновим lastAccess при любом вызове (для cleanup)
	tb.lastAccess.Store(now)

	for {
		old := tb.state.Load()
		lastSec := int64(old >> 32)
		tokens := int64(old & 0xFFFFFFFF)

		// Добавляем накопившиеся токены
		elapsed := now - lastSec
		if elapsed > 0 {
			tokens += elapsed * rate
			if tokens > burst {
				tokens = burst
			}
		}
		// Тратим один токен
		if tokens < tokenScale {
			return false
		}
		newTokens := tokens - tokenScale
		newState := packState(now, newTokens)
		if tb.state.CompareAndSwap(old, newState) {
			return true
		}
		// CAS не удался — повторяем
	}
}
