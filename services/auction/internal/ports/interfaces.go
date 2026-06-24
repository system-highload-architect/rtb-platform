package ports

import (
	"context"

	accountingv1 "rtb-platform/pb/accounting/v1"

	"rtb-platform/pkg/geospatial"

	"rtb-platform/services/auction/internal/domain"
)

// UserProfileRepo — получение профиля пользователя.
type UserProfileRepo interface {
	Get(ctx context.Context, deviceID string) (*UserProfile, error)
}

// UserProfile — представление пользователя на уровне порта.
type UserProfile struct {
	DeviceID  string
	IP        string
	UserAgent string
	Lat       float64
	Lng       float64
	Features  []float64
}

// CampaignRepo — хранилище активных кампаний.
type CampaignRepo interface {
	GetActive(ctx context.Context) ([]domain.Campaign, error)
	Update(campaigns []domain.Campaign)
	Stop()
}

type AccountingPort interface {
	Debit(ctx context.Context, req *accountingv1.DebitRequest) (*accountingv1.DebitResponse, error)
}

// FraudDetector — проверка на мошенничество.
type FraudDetector interface {
	IsSuspicious(ip, deviceID string) bool
}

// GeoResolver — получение координат билбордов кампаний.
type GeoResolver interface {
	GetBillboardLocation(campaignID uint32) (geospatial.Point, bool)
}

// EventPublisher — публикация событий.
type EventPublisher interface {
	PublishBidEvent(ctx context.Context, event domain.BidEvent) error
}
