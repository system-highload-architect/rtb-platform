package geodata

import (
	"rtb-platform/pkg/geospatial"
	"rtb-platform/services/auction/internal/domain"
)

// inmemGeoResolver хранит координаты билбордов в памяти (ключ — ID кампании).
type inmemGeoResolver struct {
	locations map[uint32]geospatial.Point
}

// NewInmemGeoResolver создаёт резолвер с начальными данными.
func NewInmemGeoResolver(initial map[uint32]geospatial.Point) domain.GeoResolver {
	return &inmemGeoResolver{locations: initial}
}

// GetBillboardLocation возвращает координаты билборда, если есть.
func (r *inmemGeoResolver) GetBillboardLocation(campaignID uint32) (geospatial.Point, bool) {
	p, ok := r.locations[campaignID]
	return p, ok
}
