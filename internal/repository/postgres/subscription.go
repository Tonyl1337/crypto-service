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

func (r *SubscriptionRepository) GetByChatID(
	ctx context.Context,
	chatID int64,
) (*domain.Subscription, error) {

	const query = `
		SELECT
			id,
			chat_id,
			enabled,
			interval_minutes,
			created_at,
			updated_at,
			last_sent_at
		FROM telegram_subscriptions
		WHERE chat_id = $1
	`

	var subscription domain.Subscription

	err := r.db.QueryRow(
		ctx,
		query,
		chatID,
	).Scan(
		&subscription.ID,
		&subscription.ChatID,
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

func (r *SubscriptionRepository) Save(
	ctx context.Context,
	subscription *domain.Subscription,
) error {

	const query = `
		INSERT INTO telegram_subscriptions (
			chat_id,
			enabled,
			interval_minutes
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id)
		DO UPDATE SET
			enabled = EXCLUDED.enabled,
			interval_minutes = EXCLUDED.interval_minutes,
			updated_at = NOW()
	`

	_, err := r.db.Exec(
		ctx,
		query,
		subscription.ChatID,
		subscription.Enabled,
		subscription.IntervalMinutes,
	)

	return err
}

func (r *SubscriptionRepository) Delete(
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
			enabled,
			interval_minutes,
			created_at,
			updated_at,
			last_sent_at
		FROM telegram_subscriptions
		WHERE enabled = TRUE
		ORDER BY id;
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

func (r *SubscriptionRepository) MarkSent(
	ctx context.Context,
	chatID int64,
	sentAt time.Time,
) error {

	const query = `
		UPDATE telegram_subscriptions
		SET last_sent_at = $2
		WHERE chat_id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		chatID,
		sentAt,
	)

	return err
}
