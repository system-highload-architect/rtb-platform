package clickhouse

import (
	"context"
	"time"

	"rtb-platform/services/analytics/internal/domain"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type CHStore struct {
	conn driver.Conn
}

func NewCHStore(ctx context.Context, dsn string) (*CHStore, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn}, // например, "localhost:9000"
		Auth: clickhouse.Auth{
			Database: "rtb",
		},
		Debug: false,
	})
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}
	s := &CHStore{conn: conn}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *CHStore) migrate(ctx context.Context) error {
	return s.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			bid_id        String,
			campaign_id   UInt32,
			device_id     String,
			price_cents   Int64,
			win           UInt8,
			ltv_score     Float64,
			geo_factor    Float64,
			impression_val Float64,
			timestamp     DateTime
		) ENGINE = MergeTree()
		ORDER BY (campaign_id, timestamp)
	`)
}

func (s *CHStore) Add(event domain.Event) {
	ctx := context.Background()
	err := s.conn.Exec(ctx, `
        INSERT INTO events (bid_id, campaign_id, device_id, price_cents, win, ltv_score, geo_factor, impression_val, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, event.BidID, event.CampaignID, event.DeviceID, event.PriceCents,
		event.Win, event.LtvScore, event.GeoFactor, event.ImpressionVal, event.Timestamp)
	if err != nil {
		// логируем ошибку
	}
}

func (s *CHStore) Query(ctx context.Context, start, end time.Time) []domain.Event {
	query := `
		SELECT bid_id, campaign_id, device_id, price_cents, win, ltv_score, geo_factor, impression_val, timestamp
		FROM events
		WHERE timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp
	`
	rows, err := s.conn.Query(ctx, query, start, end)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		var win uint8
		err := rows.Scan(&e.BidID, &e.CampaignID, &e.DeviceID, &e.PriceCents, &win,
			&e.LtvScore, &e.GeoFactor, &e.ImpressionVal, &e.Timestamp)
		if err != nil {
			continue
		}
		e.Win = win == 1
		events = append(events, e)
	}
	return events
}

func (s *CHStore) Close() error {
	return s.conn.Close()
}
