package valuation

import (
	"math"

	"rtb-platform/pkg/geospatial"
)

// GeoFactor вычисляет пространственный коэффициент ценности.
type GeoFactor struct {
	// DecayRate управляет затуханием с расстоянием (чем больше, тем быстрее падает ценность).
	DecayRate float64
	// UseRoadMultiplier включает грубый учёт извилистости дорог (умножение расстояния на 1.4)
	UseRoadMultiplier bool
}

// NewGeoFactor создаёт фактор с настройками.
func NewGeoFactor(decayRate float64, useRoadMultiplier bool) *GeoFactor {
	return &GeoFactor{DecayRate: decayRate, UseRoadMultiplier: useRoadMultiplier}
}

// Factor возвращает коэффициент (0..1) для пары «устройство — объект».
func (g *GeoFactor) Factor(userPos, targetPos geospatial.Point) float64 {
	dist := geospatial.HaversineDistance(userPos, targetPos)
	if g.UseRoadMultiplier {
		dist *= 1.4
	}
	// Экспоненциальное затухание
	return math.Exp(-g.DecayRate * dist)
}
