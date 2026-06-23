package scoring

import (
	"rtb-platform/pkg/geospatial"
	"rtb-platform/pkg/valuation"
)

// predictiveScorer — реализация Scorer, использующая пакеты valuation и geospatial.
type predictiveScorer struct {
	ltvModel  *valuation.LTVModel
	impVal    *valuation.ImpressionValue
	geoFactor *valuation.GeoFactor
}

// NewPredictiveScorer создаёт скорер с заданными коэффициентами.
// Параметры:
//
//	ltvCoeffs – коэффициенты линейной модели LTV,
//	ctr, cvr – базовые вероятности (или коэффициенты для ImpressionValue),
//	baseConversionValue – базовая ценность конверсии (в копейках),
//	geoDecay – параметр затухания для гео-фактора.
func NewPredictiveScorer(ltvCoeffs []float64, ctr, cvr, baseConversionValue, geoDecay float64) Scorer {
	return &predictiveScorer{
		ltvModel:  valuation.NewLTVModel(ltvCoeffs),
		impVal:    valuation.NewImpressionValue([]float64{ctr}, []float64{cvr}, baseConversionValue),
		geoFactor: valuation.NewGeoFactor(geoDecay, false),
	}
}

// Score вычисляет скор и оптимальную ставку.
// distanceMeters – расстояние от пользователя до рекламного щита кампании.
func (s *predictiveScorer) Score(campaign CampaignScoringData, user UserFeatures, distanceMeters float64) (float64, int64) {
	// LTV пользователя
	ltv := s.ltvModel.Predict(user.Features)

	// Оценка ценности показа (можно передать признаки рекламы, но здесь nil)
	imp := s.impVal.Value(user.Features, nil)

	// Гео-фактор: используем координаты пользователя и щита для вычисления фактора близости
	userPos := geospatial.Point{Lat: user.Lat, Lng: user.Lng}
	campPos := geospatial.Point{Lat: campaign.BillboardLat, Lng: campaign.BillboardLng}
	gf := s.geoFactor.Factor(userPos, campPos) // предполагаем, что Factor принимает два geospatial.Point

	// Итоговая ценность показа
	value := ltv * imp * gf
	// Оптимальная ставка = базовая ставка, умноженная на ценность (можно ограничить сверху)
	optimalBid := int64(float64(campaign.BidCents) * value)
	return value, optimalBid
}
