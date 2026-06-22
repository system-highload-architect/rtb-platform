package valuation

import (
	"math"

	"rtb-platform/pkg/fixedpoint"
)

// WinRateModel описывает вероятность выигрыша аукциона в зависимости от ставки.
// Может быть кусочно-заданной или аппроксимирована сигмоидой.
type WinRateModel struct {
	// Параметры сигмоиды: winRate = 1 / (1 + exp(-(bid - mid) / scale))
	Mid   float64
	Scale float64
	// Максимальная допустимая ставка (бюджетное ограничение)
	MaxBid fixedpoint.Money
}

// NewWinRateModel создаёт модель на основе исторических данных.
func NewWinRateModel(mid, scale float64, maxBid fixedpoint.Money) *WinRateModel {
	return &WinRateModel{Mid: mid, Scale: scale, MaxBid: maxBid}
}

// OptimalBid вычисляет оптимальную ставку, максимизирующую ожидаемую прибыль:
//
//	ожидаемая прибыль = (value - bid) * winRate(bid)
//
// Использует упрощённое аналитическое решение для сигмоидной winRate.
// Возвращает ставку в копейках.
func (w *WinRateModel) OptimalBid(value fixedpoint.Money) (fixedpoint.Money, error) {
	// Переводим ценность в float64 для вычислений (масштаб копеек)
	v := float64(value)

	// Если ценность нулевая или отрицательная, ставка = 0
	if v <= 0 {
		return 0, nil
	}

	// Для сигмоидной winRate ожидаемая прибыль:
	// P(b) = v * sigmoid(b) - b * sigmoid(b) = (v - b) / (1 + exp(-(b - mid)/scale))
	// Максимизируем: производная = -sigmoid(b) + (v - b) * sigmoid'(b) = 0
	// Аналитического решения нет; используем грубый поиск по нескольким точкам.
	// В продакшене можно заменить на константу, рассчитанную заранее.
	bestBid := 0.0
	bestProfit := -1.0
	step := v / 10.0
	for b := 0.0; b <= v; b += step {
		profit := (v - b) * sigmoid(b, w.Mid, w.Scale)
		if profit > bestProfit {
			bestProfit = profit
			bestBid = b
		}
	}

	// Убедимся, что ставка не превышает максимальную
	bidCents := int64(bestBid)
	maxCents := int64(w.MaxBid)
	if bidCents > maxCents {
		bidCents = maxCents
	}
	return fixedpoint.Money(bidCents), nil
}

func sigmoid(x, mid, scale float64) float64 {
	return 1.0 / (1.0 + math.Exp(-(x-mid)/scale))
}
