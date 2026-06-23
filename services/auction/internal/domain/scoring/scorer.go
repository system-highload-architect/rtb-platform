package scoring

import "rtb-platform/pkg/device"

type CampaignScoringData struct {
	ID           uint32
	BidCents     int64
	BillboardLat float64
	BillboardLng float64
	CreativeURL  string
}

type UserFeatures struct {
	Features   []float64
	Lat        float64
	Lng        float64
	DeviceType device.DeviceType
}

type Scorer interface {
	Score(campaign CampaignScoringData, user UserFeatures, distanceMeters float64) (score float64, optimalBid int64)
}
