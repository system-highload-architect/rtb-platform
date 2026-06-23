package domain

import (
	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/pkg/geospatial"
)

// Campaign — рекламная кампания, участвующая в аукционе.
type Campaign struct {
	ID           uint32
	BidCents     int64            // ставка в копейках
	DailyBudget  fixedpoint.Money // дневной бюджет
	CreativeURL  string
	BillboardLat float64
	BillboardLng float64
}

// BidEvent — событие о результате аукциона.
type BidEvent struct {
	BidID         string
	CampaignID    uint32
	DeviceID      string
	PriceCents    int64
	Win           bool
	LtvScore      float64
	GeoFactor     float64
	ImpressionVal float64
	Timestamp     int64
}

// BidResponse — результат аукциона.
type BidResponse struct {
	CampaignID  uint32
	BidPrice    int64
	CreativeURL string
	Error       string
}

// GeoResolver — интерфейс для получения координат рекламного щита.
type GeoResolver interface {
	GetBillboardLocation(campaignID uint32) (geospatial.Point, bool)
}
