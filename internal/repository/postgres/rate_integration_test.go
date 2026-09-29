//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func TestRateRepository_Save(t *testing.T) {
	db := newTestDB(t)
	cleanDatabase(t, db)

	repo := NewRateRepository(db)

	createdAt := time.Now().Truncate(time.Microsecond)

	rate := &domain.Rate{
		CoinGeckoID: "bitcoin",
		Symbol:      "BTC",
		Price:       85000.25,
		Change1H:    1.2345,
		DayLow:      83000.50,
		DayHigh:     86000.75,
		CreatedAt:   createdAt,
	}

	err := repo.Save(
		context.Background(),
		rate,
	)

	require.NoError(t, err)

	var saved domain.Rate

	err = db.QueryRow(
		context.Background(),
		`
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
		`,
		"bitcoin",
	).Scan(
		&saved.ID,
		&saved.CoinGeckoID,
		&saved.Symbol,
		&saved.Price,
		&saved.Change1H,
		&saved.DayLow,
		&saved.DayHigh,
		&saved.CreatedAt,
	)

	require.NoError(t, err)

	require.NotZero(t, saved.ID)
	require.Equal(t, "bitcoin", saved.CoinGeckoID)
	require.Equal(t, "BTC", saved.Symbol)

	require.InDelta(t, 85000.25, saved.Price, 0.000001)
	require.InDelta(t, 1.2345, saved.Change1H, 0.000001)
	require.InDelta(t, 83000.50, saved.DayLow, 0.000001)
	require.InDelta(t, 86000.75, saved.DayHigh, 0.000001)

	require.Equal(
		t,
		createdAt.Format("2006-01-02 15:04:05.000000"),
		saved.CreatedAt.Format("2006-01-02 15:04:05.000000"),
	)
}

func TestRateRepository_GetByCoinGeckoID(t *testing.T) {
	db := newTestDB(t)

	_, err := db.Exec(
		context.Background(),
		"TRUNCATE TABLE rates RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewRateRepository(db)

	ctx := context.Background()

	now := time.Now()

	rates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       85000,
			Change1H:    0.5,
			DayLow:      83000,
			DayHigh:     86000,
			CreatedAt:   now.Add(-2 * time.Minute),
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       2700,
			Change1H:    0.3,
			DayLow:      2600,
			DayHigh:     2800,
			CreatedAt:   now.Add(-1 * time.Minute),
		},
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       85500,
			Change1H:    0.7,
			DayLow:      83000,
			DayHigh:     86000,
			CreatedAt:   now,
		},
	}

	for i := range rates {
		err := repo.Save(ctx, &rates[i])
		require.NoError(t, err)
	}

	result, err := repo.GetByCoinGeckoID(
		ctx,
		"bitcoin",
	)

	require.NoError(t, err)
	require.Len(t, result, 2)

	// Репозиторий должен вернуть только Bitcoin.
	require.Equal(t, "bitcoin", result[0].CoinGeckoID)
	require.Equal(t, "bitcoin", result[1].CoinGeckoID)

	require.Equal(t, "BTC", result[0].Symbol)
	require.Equal(t, "BTC", result[1].Symbol)

	// ORDER BY created_at DESC:
	// сначала должна идти самая свежая запись.
	require.InDelta(t, 85500.0, result[0].Price, 0.000001)
	require.InDelta(t, 85000.0, result[1].Price, 0.000001)
}

func TestRateRepository_GetBySymbol(t *testing.T) {
	db := newTestDB(t)

	_, err := db.Exec(
		context.Background(),
		"TRUNCATE TABLE rates RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewRateRepository(db)

	ctx := context.Background()
	now := time.Now()

	rates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       84000,
			Change1H:    0.1,
			DayLow:      82000,
			DayHigh:     86000,
			CreatedAt:   now.Add(-2 * time.Minute),
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       2700,
			Change1H:    0.2,
			DayLow:      2600,
			DayHigh:     2800,
			CreatedAt:   now.Add(-1 * time.Minute),
		},
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       85000,
			Change1H:    0.3,
			DayLow:      82000,
			DayHigh:     86000,
			CreatedAt:   now,
		},
	}

	for i := range rates {
		err := repo.Save(ctx, &rates[i])
		require.NoError(t, err)
	}

	result, err := repo.GetBySymbol(
		ctx,
		"BTC",
	)

	require.NoError(t, err)
	require.Len(t, result, 2)

	// Только BTC.
	require.Equal(t, "BTC", result[0].Symbol)
	require.Equal(t, "BTC", result[1].Symbol)

	require.Equal(t, "bitcoin", result[0].CoinGeckoID)
	require.Equal(t, "bitcoin", result[1].CoinGeckoID)

	// ORDER BY created_at DESC.
	require.InDelta(t, 85000.0, result[0].Price, 0.000001)
	require.InDelta(t, 84000.0, result[1].Price, 0.000001)
}

func TestRateRepository_GetLatest(t *testing.T) {
	db := newTestDB(t)

	_, err := db.Exec(
		context.Background(),
		"TRUNCATE TABLE rates RESTART IDENTITY",
	)
	require.NoError(t, err)

	repo := NewRateRepository(db)

	ctx := context.Background()
	now := time.Now()

	rates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       83000,
			Change1H:    0.1,
			DayLow:      82000,
			DayHigh:     86000,
			CreatedAt:   now.Add(-10 * time.Minute),
		},
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       85000,
			Change1H:    0.2,
			DayLow:      82000,
			DayHigh:     86000,
			CreatedAt:   now,
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       2600,
			Change1H:    -0.1,
			DayLow:      2500,
			DayHigh:     2800,
			CreatedAt:   now.Add(-5 * time.Minute),
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       2700,
			Change1H:    0.3,
			DayLow:      2500,
			DayHigh:     2800,
			CreatedAt:   now.Add(-1 * time.Minute),
		},
		{
			CoinGeckoID: "dogecoin",
			Symbol:      "DOGE",
			Price:       0.1,
			Change1H:    1.5,
			DayLow:      0.09,
			DayHigh:     0.11,
			CreatedAt:   now.Add(-2 * time.Minute),
		},
	}

	for i := range rates {
		err := repo.Save(ctx, &rates[i])
		require.NoError(t, err)
	}

	result, err := repo.GetLatest(ctx)

	require.NoError(t, err)
	require.Len(t, result, 3)

	resultByID := make(map[string]domain.Rate)

	for _, rate := range result {
		resultByID[rate.CoinGeckoID] = rate
	}

	require.Contains(t, resultByID, "bitcoin")
	require.Contains(t, resultByID, "ethereum")
	require.Contains(t, resultByID, "dogecoin")

	require.InDelta(
		t,
		85000.0,
		resultByID["bitcoin"].Price,
		0.000001,
	)

	require.InDelta(
		t,
		2700.0,
		resultByID["ethereum"].Price,
		0.000001,
	)

	require.InDelta(
		t,
		0.1,
		resultByID["dogecoin"].Price,
		0.000001,
	)
}
