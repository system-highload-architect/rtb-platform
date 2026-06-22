package experiment

import (
	"hash/fnv"
)

// Experiments хранит набор A/B-экспериментов с долями трафика.
// Конфигурация задаётся один раз при старте сервиса и не изменяется.
type Experiments struct {
	flags map[string]float64 // имя эксперимента → доля (0.0 .. 1.0)
}

// New создаёт Experiments из карты (имя → доля).
func New(flags map[string]float64) *Experiments {
	if flags == nil {
		flags = make(map[string]float64)
	}
	return &Experiments{flags: flags}
}

// IsInExperiment проверяет, попадает ли userID в эксперимент с именем name.
// Если эксперимент не определён, возвращает false.
// Детерминирован: один и тот же userID всегда вернёт один результат.
func (e *Experiments) IsInExperiment(userID, name string) bool {
	ratio, ok := e.flags[name]
	if !ok || ratio <= 0 {
		return false
	}
	if ratio >= 1.0 {
		return true
	}
	// Вычисляем хэш от пары (userID, name) и преобразуем в число в диапазоне [0, 1)
	h := fnv.New64a()
	h.Write([]byte(userID))
	h.Write([]byte(name))
	hashVal := h.Sum64()
	// Нормализуем: используем старшие 53 бита для double (как в Java)
	// Простое деление на MaxUint64 даст равномерное распределение.
	normalized := float64(hashVal) / float64(^uint64(0))
	return normalized < ratio
}
