package domain

import (
	"context"
	"strconv"
	"time"
)

// EventStore – интерфейс хранилища событий.
type EventStore interface {
	Add(event Event)
	Query(ctx context.Context, start, end time.Time) []Event
}

// ReportService – формирует отчёты на основе событий.
type ReportService struct {
	store EventStore
}

func NewReportService(store EventStore) *ReportService {
	return &ReportService{store: store}
}

// ReportRow – строка отчёта.
type ReportRow struct {
	DimensionValues map[string]string
	MetricValues    map[string]float64
}

// GenerateReport возвращает строки отчёта, сгруппированные по измерениям.
func (s *ReportService) GenerateReport(ctx context.Context, start, end time.Time, dimensions []string, metrics []string) ([]*ReportRow, error) {
	events := s.store.Query(ctx, start, end)

	type aggKey struct {
		campaignID string
		deviceType string
	}
	aggs := make(map[aggKey]*aggregate)

	for _, e := range events {
		key := aggKey{
			campaignID: strconv.FormatUint(uint64(e.CampaignID), 10),
			// deviceType можно добавить, если будет поле в Event
		}
		if _, ok := aggs[key]; !ok {
			aggs[key] = &aggregate{}
		}
		a := aggs[key]
		if e.Win {
			a.impressions++
			a.clicks++ // условно, в реальности клики отдельно
			a.spend += float64(e.PriceCents) / 100.0
		}
	}

	var rows []*ReportRow
	for key, a := range aggs {
		dimVals := map[string]string{
			"campaign_id": key.campaignID,
		}
		metVals := map[string]float64{
			"impressions": float64(a.impressions),
			"clicks":      float64(a.clicks),
			"spend":       a.spend,
		}
		rows = append(rows, &ReportRow{
			DimensionValues: dimVals,
			MetricValues:    metVals,
		})
	}
	return rows, nil
}

type aggregate struct {
	impressions int64
	clicks      int64
	spend       float64
}
