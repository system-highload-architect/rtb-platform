package sampler

import (
	"math/rand"
	"sync"
)

// Sampler принимает решение о выборке события с заданной вероятностью.
type Sampler struct {
	mu   sync.Mutex
	rate float64 // вероятность от 0.0 до 1.0
	rng  *rand.Rand
}

// NewSampler создаёт сэмплер с вероятностью rate.
// rate должен быть в интервале [0, 1].
func NewSampler(rate float64) *Sampler {
	return &Sampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(rand.Int63())),
	}
}

// Sample возвращает true, если событие нужно сохранить/обработать.
func (s *Sampler) Sample() bool {
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0.0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rng.Float64() < s.rate
}
