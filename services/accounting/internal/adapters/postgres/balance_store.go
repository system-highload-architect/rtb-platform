package postgres

import (
	"context"
	"errors"

	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/services/accounting/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(ctx context.Context, dsn string) (*pgStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	s := &pgStore{pool: pool}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *pgStore) migrate(ctx context.Context) error {
	// Можно загрузить SQL-файл миграции или выполнить встроенный SQL
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS balances (
			campaign_id VARCHAR(255) PRIMARY KEY,
			amount BIGINT NOT NULL DEFAULT 0,
			scale INT NOT NULL DEFAULT 2
		);

		CREATE OR REPLACE FUNCTION debit_balance(...) ... ; -- здесь полный код функции из миграции
	`)
	return err
}

func (s *pgStore) Get(campaignID string) (fixedpoint.Money, bool) {
	var amount int64
	err := s.pool.QueryRow(context.Background(),
		"SELECT amount FROM balances WHERE campaign_id = $1", campaignID,
	).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false
		}
		return 0, false
	}
	return fixedpoint.Money(amount), true
}

func (s *pgStore) Debit(campaignID string, amount fixedpoint.Money) error {
	var success bool
	var remaining int64
	var scaleOut int32
	var errorMsg string

	err := s.pool.QueryRow(context.Background(),
		"SELECT * FROM debit_balance($1, $2)",
		campaignID, int64(amount),
	).Scan(&success, &remaining, &scaleOut, &errorMsg)
	if err != nil {
		return err
	}
	if !success {
		if errorMsg == "Campaign not found" {
			return domain.ErrCampaignNotFound
		}
		if errorMsg == "Insufficient funds" {
			return domain.ErrInsufficientFunds
		}
		return errors.New(errorMsg)
	}
	return nil
}

func (s *pgStore) Set(campaignID string, balance fixedpoint.Money) {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO balances (campaign_id, amount, scale)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (campaign_id) DO UPDATE SET amount = $2, scale = $3`,
		campaignID, int64(balance), int32(2),
	)
	if err != nil {
		// логируем
	}
}

func (s *pgStore) Close() {
	s.pool.Close()
}
