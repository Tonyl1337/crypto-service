package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type RateRepository struct {
	db *pgxpool.Pool
}

func NewRateRepository(db *pgxpool.Pool) *RateRepository {
	return &RateRepository{
		db: db,
	}
}

func (r *RateRepository) GetByCoinGeckoID(
	ctx context.Context,
	coinGeckoID string,
) ([]domain.Rate, error) {

	const query = `
		SELECT
			id,
			coingecko_id,
			symbol,
			price,
			change_1h,
			day_low,
			day_high,
			created_at
		FROM rates
		WHERE coingecko_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.db.Query(
		ctx,
		query,
		coinGeckoID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRates(rows)
}

func (r *RateRepository) Save(
	ctx context.Context,
	rate *domain.Rate,
) error {

	const query = `
		INSERT INTO rates (
			coingecko_id,
			symbol,
			price,
			change_1h,
			day_low,
			day_high,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		rate.CoinGeckoID,
		rate.Symbol,
		rate.Price,
		rate.Change1H,
		rate.DayLow,
		rate.DayHigh,
		rate.CreatedAt,
	)

	return err
}

func (r *RateRepository) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {

	const query = `
		SELECT DISTINCT ON (coingecko_id)
		id,
		coingecko_id,
		symbol,
		price,
		change_1h,
		day_low,
		day_high,
		created_at
	FROM rates
	ORDER BY coingecko_id, created_at DESC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRates(rows)
}

func (r *RateRepository) GetBySymbol(
	ctx context.Context,
	symbol string,
) ([]domain.Rate, error) {

	const query = `
		SELECT
		id,
		coingecko_id,
		symbol,
		price,
		change_1h,
		day_low,
		day_high,
		created_at
	FROM rates
	WHERE symbol = $1
	ORDER BY created_at DESC;
	`

	rows, err := r.db.Query(ctx, query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRates(rows)
}
