package valuation

import (
	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/pkg/geospatial"
)

// Scorer — комплексная оценка ценности показа для конкретного пользователя и кампании.
type Scorer struct {
	LTV        *LTVModel
	Impression *ImpressionValue
	Geo        *GeoFactor
	WinRate    *WinRateModel
}

// NewScorer собирает все компоненты воедино.
func NewScorer(ltv *LTVModel, imp *ImpressionValue, geo *GeoFactor, winRate *WinRateModel) *Scorer {
	return &Scorer{
		LTV:        ltv,
		Impression: imp,
		Geo:        geo,
		WinRate:    winRate,
	}
}

// Score вычисляет итоговый рейтинг показа и оптимальную ставку.
// Возвращает сырой рейтинг (score) и готовую ставку в fixedpoint.
func (s *Scorer) Score(
	userFeatures []float64,
	adFeatures []float64,
	userPos, targetPos geospatial.Point,
) (score float64, bid fixedpoint.Money, err error) {
	// 1. LTV пользователя
	ltv := s.LTV.Predict(userFeatures)

	// 2. Ценность показа
	imp := s.Impression.Value(userFeatures, adFeatures)

	// 3. Гео-фактор
	geo := s.Geo.Factor(userPos, targetPos)

	// Итоговая ценность показа (условные единицы)
	score = (ltv + imp) * geo

	// Переводим в деньги (условно: 1 единица = 1 копейка)
	valueCents := int64(score)
	if valueCents < 0 {
		valueCents = 0
	}
	value := fixedpoint.Money(valueCents)

	// 4. Оптимальная ставка
	bid, err = s.WinRate.OptimalBid(value)
	return
}
