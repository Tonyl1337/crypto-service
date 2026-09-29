//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func TestSubscriptionRepository_Save(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscription := &domain.Subscription{
		ChatID:          123456,
		Symbol:          "DOGE",
		CoinGeckoID:     "dogecoin",
		Enabled:         true,
		IntervalMinutes: 5,
	}

	err = repo.Save(ctx, subscription)
	require.NoError(t, err)

	saved, err := repo.GetByChatIDAndCoin(
		ctx,
		123456,
		"dogecoin",
	)

	require.NoError(t, err)
	require.NotNil(t, saved)

	require.Equal(t, int64(123456), saved.ChatID)
	require.Equal(t, "DOGE", saved.Symbol)
	require.Equal(t, "dogecoin", saved.CoinGeckoID)
	require.True(t, saved.Enabled)
	require.Equal(t, 5, saved.IntervalMinutes)
	require.Nil(t, saved.LastSentAt)
	require.NotZero(t, saved.ID)
	require.False(t, saved.CreatedAt.IsZero())
	require.False(t, saved.UpdatedAt.IsZero())
}

func TestSubscriptionRepository_Save_Upsert(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	original := &domain.Subscription{
		ChatID:          777,
		Symbol:          "DOGE",
		CoinGeckoID:     "dogecoin",
		Enabled:         true,
		IntervalMinutes: 5,
	}

	err = repo.Save(ctx, original)
	require.NoError(t, err)

	sentAt := time.Now().Add(-time.Minute)

	err = repo.MarkSent(
		ctx,
		777,
		"dogecoin",
		sentAt,
	)
	require.NoError(t, err)

	beforeUpdate, err := repo.GetByChatIDAndCoin(
		ctx,
		777,
		"dogecoin",
	)
	require.NoError(t, err)
	require.NotNil(t, beforeUpdate.LastSentAt)

	originalID := beforeUpdate.ID

	updated := &domain.Subscription{
		ChatID:          777,
		Symbol:          "DOGE",
		CoinGeckoID:     "dogecoin",
		Enabled:         true,
		IntervalMinutes: 15,
	}

	err = repo.Save(ctx, updated)
	require.NoError(t, err)

	afterUpdate, err := repo.GetByChatIDAndCoin(
		ctx,
		777,
		"dogecoin",
	)
	require.NoError(t, err)

	// ON CONFLICT должен обновить существующую строку,
	// а не создать новую.
	require.Equal(t, originalID, afterUpdate.ID)

	require.Equal(t, int64(777), afterUpdate.ChatID)
	require.Equal(t, "DOGE", afterUpdate.Symbol)
	require.Equal(t, "dogecoin", afterUpdate.CoinGeckoID)
	require.True(t, afterUpdate.Enabled)

	// Интервал должен обновиться.
	require.Equal(t, 15, afterUpdate.IntervalMinutes)

	// Save при повторном включении подписки
	// должен сбросить last_sent_at.
	require.Nil(t, afterUpdate.LastSentAt)
}

func TestSubscriptionRepository_GetEnabled(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscriptions := []domain.Subscription{
		{
			ChatID:          100,
			Symbol:          "BTC",
			CoinGeckoID:     "bitcoin",
			Enabled:         true,
			IntervalMinutes: 5,
		},
		{
			ChatID:          200,
			Symbol:          "ETH",
			CoinGeckoID:     "ethereum",
			Enabled:         false,
			IntervalMinutes: 10,
		},
		{
			ChatID:          300,
			Symbol:          "DOGE",
			CoinGeckoID:     "dogecoin",
			Enabled:         true,
			IntervalMinutes: 15,
		},
	}

	for i := range subscriptions {
		err := repo.Save(ctx, &subscriptions[i])
		require.NoError(t, err)
	}

	result, err := repo.GetEnabled(ctx)

	require.NoError(t, err)
	require.Len(t, result, 2)

	require.Equal(t, int64(100), result[0].ChatID)
	require.Equal(t, "bitcoin", result[0].CoinGeckoID)
	require.True(t, result[0].Enabled)

	require.Equal(t, int64(300), result[1].ChatID)
	require.Equal(t, "dogecoin", result[1].CoinGeckoID)
	require.True(t, result[1].Enabled)
}

func TestSubscriptionRepository_GetEnabledCoinGeckoIDs(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscriptions := []domain.Subscription{
		{
			ChatID:          100,
			Symbol:          "DOGE",
			CoinGeckoID:     "dogecoin",
			Enabled:         true,
			IntervalMinutes: 5,
		},
		{
			ChatID:          200,
			Symbol:          "DOGE",
			CoinGeckoID:     "dogecoin",
			Enabled:         true,
			IntervalMinutes: 10,
		},
		{
			ChatID:          300,
			Symbol:          "SOL",
			CoinGeckoID:     "solana",
			Enabled:         true,
			IntervalMinutes: 15,
		},
		{
			ChatID:          400,
			Symbol:          "ETH",
			CoinGeckoID:     "ethereum",
			Enabled:         false,
			IntervalMinutes: 20,
		},
	}

	for i := range subscriptions {
		err := repo.Save(ctx, &subscriptions[i])
		require.NoError(t, err)
	}

	result, err := repo.GetEnabledCoinGeckoIDs(ctx)

	require.NoError(t, err)

	require.Equal(
		t,
		[]string{
			"dogecoin",
			"solana",
		},
		result,
	)
}

func TestSubscriptionRepository_MarkSent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscription := &domain.Subscription{
		ChatID:          123,
		Symbol:          "DOGE",
		CoinGeckoID:     "dogecoin",
		Enabled:         true,
		IntervalMinutes: 5,
	}

	err = repo.Save(ctx, subscription)
	require.NoError(t, err)

	before, err := repo.GetByChatIDAndCoin(
		ctx,
		123,
		"dogecoin",
	)
	require.NoError(t, err)
	require.Nil(t, before.LastSentAt)

	sentAt := time.Now().UTC().Truncate(time.Microsecond)

	err = repo.MarkSent(
		ctx,
		123,
		"dogecoin",
		sentAt,
	)
	require.NoError(t, err)

	after, err := repo.GetByChatIDAndCoin(
		ctx,
		123,
		"dogecoin",
	)
	require.NoError(t, err)
	require.NotNil(t, after.LastSentAt)

	// PostgreSQL column is timestamp without time zone,
	// поэтому сравниваем календарные компоненты времени.
	require.Equal(
		t,
		sentAt.Format("2006-01-02 15:04:05.999999"),
		after.LastSentAt.Format("2006-01-02 15:04:05.999999"),
	)
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscriptions := []domain.Subscription{
		{
			ChatID:          123,
			Symbol:          "DOGE",
			CoinGeckoID:     "dogecoin",
			Enabled:         true,
			IntervalMinutes: 5,
		},
		{
			ChatID:          123,
			Symbol:          "SOL",
			CoinGeckoID:     "solana",
			Enabled:         true,
			IntervalMinutes: 10,
		},
	}

	for i := range subscriptions {
		err := repo.Save(ctx, &subscriptions[i])
		require.NoError(t, err)
	}

	err = repo.Delete(
		ctx,
		123,
		"dogecoin",
	)
	require.NoError(t, err)

	// DOGE должен исчезнуть.
	_, err = repo.GetByChatIDAndCoin(
		ctx,
		123,
		"dogecoin",
	)
	require.Error(t, err)

	// SOL должен остаться.
	solana, err := repo.GetByChatIDAndCoin(
		ctx,
		123,
		"solana",
	)
	require.NoError(t, err)

	require.Equal(t, int64(123), solana.ChatID)
	require.Equal(t, "SOL", solana.Symbol)
	require.Equal(t, "solana", solana.CoinGeckoID)
}

func TestSubscriptionRepository_DeleteAll(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"TRUNCATE TABLE telegram_subscriptions RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewSubscriptionRepository(db)

	subscriptions := []domain.Subscription{
		{
			ChatID:          123,
			Symbol:          "DOGE",
			CoinGeckoID:     "dogecoin",
			Enabled:         true,
			IntervalMinutes: 5,
		},
		{
			ChatID:          123,
			Symbol:          "SOL",
			CoinGeckoID:     "solana",
			Enabled:         true,
			IntervalMinutes: 10,
		},
		{
			ChatID:          999,
			Symbol:          "ETH",
			CoinGeckoID:     "ethereum",
			Enabled:         true,
			IntervalMinutes: 15,
		},
	}

	for i := range subscriptions {
		err := repo.Save(ctx, &subscriptions[i])
		require.NoError(t, err)
	}

	err = repo.DeleteAll(ctx, 123)
	require.NoError(t, err)

	// Обе подписки chatID=123 должны исчезнуть.
	_, err = repo.GetByChatIDAndCoin(
		ctx,
		123,
		"dogecoin",
	)
	require.Error(t, err)

	_, err = repo.GetByChatIDAndCoin(
		ctx,
		123,
		"solana",
	)
	require.Error(t, err)

	// Подписка другого пользователя должна остаться.
	ethereum, err := repo.GetByChatIDAndCoin(
		ctx,
		999,
		"ethereum",
	)
	require.NoError(t, err)

	require.Equal(t, int64(999), ethereum.ChatID)
	require.Equal(t, "ETH", ethereum.Symbol)
	require.Equal(t, "ethereum", ethereum.CoinGeckoID)
}
