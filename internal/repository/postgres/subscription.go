package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(
	db *pgxpool.Pool,
) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) GetByChatIDAndCoin(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) (*domain.Subscription, error) {

	const query = `
		SELECT
			id,
			chat_id,
			symbol,
			coingecko_id,
			enabled,
			interval_minutes,
			created_at,
			updated_at,
			last_sent_at
		FROM telegram_subscriptions
		WHERE chat_id = $1
		  AND coingecko_id = $2
	`

	var subscription domain.Subscription

	err := r.db.QueryRow(
		ctx,
		query,
		chatID,
		coinGeckoID,
	).Scan(
		&subscription.ID,
		&subscription.ChatID,
		&subscription.Symbol,
		&subscription.CoinGeckoID,
		&subscription.Enabled,
		&subscription.IntervalMinutes,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
		&subscription.LastSentAt,
	)
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (r *SubscriptionRepository) GetByChatID(
	ctx context.Context,
	chatID int64,
) ([]domain.Subscription, error) {

	const query = `
		SELECT
			id,
			chat_id,
			symbol,
			coingecko_id,
			enabled,
			interval_minutes,
			created_at,
			updated_at,
			last_sent_at
		FROM telegram_subscriptions
		WHERE chat_id = $1
		  AND enabled = TRUE
		ORDER BY symbol
	`

	rows, err := r.db.Query(
		ctx,
		query,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make(
		[]domain.Subscription,
		0,
	)

	for rows.Next() {
		var subscription domain.Subscription

		err := rows.Scan(
			&subscription.ID,
			&subscription.ChatID,
			&subscription.Symbol,
			&subscription.CoinGeckoID,
			&subscription.Enabled,
			&subscription.IntervalMinutes,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.LastSentAt,
		)
		if err != nil {
			return nil, err
		}

		subscriptions = append(
			subscriptions,
			subscription,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Save(
	ctx context.Context,
	subscription *domain.Subscription,
) error {

	const query = `
		INSERT INTO telegram_subscriptions (
			chat_id,
			symbol,
			coingecko_id,
			enabled,
			interval_minutes
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chat_id, coingecko_id)
		DO UPDATE SET
			symbol = EXCLUDED.symbol,
			enabled = EXCLUDED.enabled,
			interval_minutes = EXCLUDED.interval_minutes,
			updated_at = NOW(),
			last_sent_at = NULL
	`

	_, err := r.db.Exec(
		ctx,
		query,
		subscription.ChatID,
		subscription.Symbol,
		subscription.CoinGeckoID,
		subscription.Enabled,
		subscription.IntervalMinutes,
	)

	return err
}

func (r *SubscriptionRepository) Delete(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) error {

	const query = `
		DELETE FROM telegram_subscriptions
		WHERE chat_id = $1
		  AND coingecko_id = $2
	`

	_, err := r.db.Exec(
		ctx,
		query,
		chatID,
		coinGeckoID,
	)

	return err
}

func (r *SubscriptionRepository) DeleteAll(
	ctx context.Context,
	chatID int64,
) error {

	const query = `
		DELETE FROM telegram_subscriptions
		WHERE chat_id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		chatID,
	)

	return err
}

func (r *SubscriptionRepository) GetEnabled(
	ctx context.Context,
) ([]domain.Subscription, error) {

	const query = `
		SELECT
			id,
			chat_id,
			symbol,
			coingecko_id,
			enabled,
			interval_minutes,
			created_at,
			updated_at,
			last_sent_at
		FROM telegram_subscriptions
		WHERE enabled = TRUE
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make(
		[]domain.Subscription,
		0,
	)

	for rows.Next() {
		var subscription domain.Subscription

		err := rows.Scan(
			&subscription.ID,
			&subscription.ChatID,
			&subscription.Symbol,
			&subscription.CoinGeckoID,
			&subscription.Enabled,
			&subscription.IntervalMinutes,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.LastSentAt,
		)
		if err != nil {
			return nil, err
		}

		subscriptions = append(
			subscriptions,
			subscription,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetEnabledCoinGeckoIDs(
	ctx context.Context,
) ([]string, error) {

	const query = `
		SELECT DISTINCT coingecko_id
		FROM telegram_subscriptions
		WHERE enabled = TRUE
		ORDER BY coingecko_id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	coinGeckoIDs := make([]string, 0)

	for rows.Next() {
		var coinGeckoID string

		if err := rows.Scan(&coinGeckoID); err != nil {
			return nil, err
		}

		coinGeckoIDs = append(
			coinGeckoIDs,
			coinGeckoID,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return coinGeckoIDs, nil
}

func (r *SubscriptionRepository) MarkSent(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
	sentAt time.Time,
) error {

	const query = `
		UPDATE telegram_subscriptions
		SET last_sent_at = $3
		WHERE chat_id = $1
		  AND coingecko_id = $2
	`

	_, err := r.db.Exec(
		ctx,
		query,
		chatID,
		coinGeckoID,
		sentAt,
	)

	return err
}
