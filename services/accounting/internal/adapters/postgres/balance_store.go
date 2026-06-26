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
		-- Миграция: создание таблицы и хранимых функций

		CREATE TABLE IF NOT EXISTS balances (
			campaign_id VARCHAR(255) PRIMARY KEY,
			amount BIGINT NOT NULL DEFAULT 0,
			scale INT NOT NULL DEFAULT 2
		);

		-- Функция получения баланса (необязательно, можно и прямым SELECT)
		CREATE OR REPLACE FUNCTION get_balance(campaign_id_in VARCHAR)
		RETURNS TABLE(amount BIGINT, scale INT) AS $$
		BEGIN
			RETURN QUERY SELECT b.amount, b.scale FROM balances b WHERE b.campaign_id = campaign_id_in;
		END;
		$$ LANGUAGE plpgsql;

		-- Функция списания средств
		CREATE OR REPLACE FUNCTION debit_balance(
			campaign_id_in VARCHAR,
			amount_in BIGINT,
			OUT success BOOLEAN,
			OUT remaining BIGINT,
			OUT scale_out INT,
			OUT error_msg TEXT
		) AS $$
		DECLARE
			current_amount BIGINT;
			current_scale INT;
		BEGIN
			-- Блокируем строку
			SELECT b.amount, b.scale INTO current_amount, current_scale
			FROM balances b
			WHERE b.campaign_id = campaign_id_in
			FOR UPDATE;

			IF NOT FOUND THEN
				success := false;
				error_msg := 'Campaign not found';
				RETURN;
			END IF;

			IF current_amount < amount_in THEN
				success := false;
				remaining := current_amount;
				scale_out := current_scale;
				error_msg := 'Insufficient funds';
				RETURN;
			END IF;

			-- Списываем
			UPDATE balances b SET amount = current_amount - amount_in
			WHERE b.campaign_id = campaign_id_in;

			success := true;
			remaining := current_amount - amount_in;
			scale_out := current_scale;
			error_msg := '';
		END;
		$$ LANGUAGE plpgsql;
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
