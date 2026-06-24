package domain

import "time"

// Event – событие аукциона, поступающее от внешних систем.
type Event struct {
	BidID         string
	CampaignID    uint32
	DeviceID      string
	PriceCents    int64
	Win           bool
	LtvScore      float64
	GeoFactor     float64
	ImpressionVal float64
	Timestamp     time.Time
}
