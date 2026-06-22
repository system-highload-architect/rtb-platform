package breaker

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrBreakerOpen = errors.New("breaker: circuit is open")
)

// State представляет текущее состояние Circuit Breaker.
type State int32

const (
	Closed   State = 0
	Open     State = 1
	HalfOpen State = 2
)

// Breaker — Circuit Breaker для защиты внешних вызовов.
type Breaker struct {
	name             string
	threshold        int           // количество последовательных ошибок для размыкания
	timeout          time.Duration // сколько времени цепь разомкнута
	state            atomic.Int32
	consecutiveFails atomic.Int64
	lastFailureTime  atomic.Int64 // UnixNano
	halfOpenCount    atomic.Int64 // количество попыток в HalfOpen (обычно 0 или 1)
}

// New создаёт Breaker.
func New(name string, threshold int, timeout time.Duration) *Breaker {
	b := &Breaker{
		name:      name,
		threshold: threshold,
		timeout:   timeout,
	}
	b.state.Store(int32(Closed))
	return b
}

// State возвращает текущее состояние.
func (b *Breaker) State() State {
	return State(b.state.Load())
}

// Execute выполняет функцию fn с защитой от сбоев.
// Если цепь разомкнута, сразу возвращает ErrBreakerOpen.
// При успехе сбрасывает счётчик ошибок.
// При ошибке увеличивает счётчик и, если порог достигнут, размыкает цепь.
func (b *Breaker) Execute(ctx context.Context, fn func() error) error {
	state := b.State()
	if state == Open {
		// Проверим, не пора ли в Half-Open
		lastFail := b.lastFailureTime.Load()
		if time.Since(time.Unix(0, lastFail)) > b.timeout {
			// Переходим в Half-Open, но только один поток может попробовать
			if b.state.CompareAndSwap(int32(Open), int32(HalfOpen)) {
				b.halfOpenCount.Store(0)
				return b.tryCall(ctx, fn)
			}
			// Другой поток уже перевёл, ждём
			return ErrBreakerOpen
		}
		return ErrBreakerOpen
	}

	if state == HalfOpen {
		// Разрешаем только один пробный вызов
		if b.halfOpenCount.Add(1) == 1 {
			return b.tryCall(ctx, fn)
		}
		return ErrBreakerOpen
	}

	// Closed — нормальный вызов
	return b.tryCall(ctx, fn)
}

func (b *Breaker) tryCall(ctx context.Context, fn func() error) error {
	// Проверяем контекст
	if ctx.Err() != nil {
		return ctx.Err()
	}

	err := fn()
	if err != nil {
		// Ошибка
		b.onFailure()
		return err
	}

	// Успех — сбрасываем счётчик ошибок, возвращаемся в Closed
	b.consecutiveFails.Store(0)
	b.state.Store(int32(Closed))
	return nil
}

func (b *Breaker) onFailure() {
	fails := b.consecutiveFails.Add(1)
	b.lastFailureTime.Store(time.Now().UnixNano())
	if fails >= int64(b.threshold) {
		b.state.Store(int32(Open))
	}
}
