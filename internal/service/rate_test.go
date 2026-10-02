package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockRateRepository struct {
	savedRates []domain.Rate

	getLatestResult        []domain.Rate
	getBySymbolResult      []domain.Rate
	getByCoinGeckoIDResult []domain.Rate

	saveErr             error
	getLatestErr        error
	getBySymbolErr      error
	getByCoinGeckoIDErr error
}

func (m *mockRateRepository) Save(
	ctx context.Context,
	rate *domain.Rate,
) error {
	if m.saveErr != nil {
		return m.saveErr
	}

	m.savedRates = append(m.savedRates, *rate)

	return nil
}

func (m *mockRateRepository) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {
	return m.getLatestResult, m.getLatestErr
}

func (m *mockRateRepository) GetBySymbol(
	ctx context.Context,
	symbol string,
) ([]domain.Rate, error) {
	return m.getBySymbolResult, m.getBySymbolErr
}

func (m *mockRateRepository) GetByCoinGeckoID(
	ctx context.Context,
	coinGeckoID string,
) ([]domain.Rate, error) {
	return m.getByCoinGeckoIDResult, m.getByCoinGeckoIDErr
}

type mockExchangeClient struct {
	rates        []domain.Rate
	err          error
	requestedIDs []string
	called       bool
}

func (m *mockExchangeClient) GetRates(
	ctx context.Context,
	coinGeckoIDs []string,
) ([]domain.Rate, error) {
	m.called = true
	m.requestedIDs = append([]string(nil), coinGeckoIDs...)

	return m.rates, m.err
}

func TestRateService_UpdateRates(t *testing.T) {
	now := time.Now()

	rates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       63470,
			Change1H:    0.1,
			DayLow:      63267,
			DayHigh:     64329,
			CreatedAt:   now,
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       1883.48,
			Change1H:    -0.3,
			DayLow:      1876.31,
			DayHigh:     1918.99,
			CreatedAt:   now,
		},
	}

	repo := &mockRateRepository{}

	client := &mockExchangeClient{
		rates: rates,
	}

	rateService := NewRateService(
		repo,
		client,
	)

	err := rateService.UpdateRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.savedRates) != 2 {
		t.Fatalf(
			"expected 2 saved rates, got %d",
			len(repo.savedRates),
		)
	}

	if repo.savedRates[0].CoinGeckoID != "bitcoin" {
		t.Errorf(
			"expected first CoinGecko ID bitcoin, got %s",
			repo.savedRates[0].CoinGeckoID,
		)
	}

	if repo.savedRates[1].CoinGeckoID != "ethereum" {
		t.Errorf(
			"expected second CoinGecko ID ethereum, got %s",
			repo.savedRates[1].CoinGeckoID,
		)
	}

	assertIDs(
		t,
		client.requestedIDs,
		[]string{"bitcoin", "ethereum"},
	)
}

func TestRateService_UpdateRates_ClientError(t *testing.T) {
	expectedErr := errors.New("coin gecko error")

	repo := &mockRateRepository{}

	client := &mockExchangeClient{
		err: expectedErr,
	}

	rateService := NewRateService(
		repo,
		client,
	)

	err := rateService.UpdateRates(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}

	if len(repo.savedRates) != 0 {
		t.Fatalf(
			"expected 0 saved rates, got %d",
			len(repo.savedRates),
		)
	}
}

func TestRateService_UpdateRates_RepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New("database error")

	rates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       63470,
			Change1H:    0.1,
			DayLow:      63267,
			DayHigh:     64329,
			CreatedAt:   time.Now(),
		},
	}

	repo := &mockRateRepository{
		saveErr: expectedErr,
	}

	client := &mockExchangeClient{
		rates: rates,
	}

	rateService := NewRateService(
		repo,
		client,
	)

	err := rateService.UpdateRates(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}
}

func TestRateService_GetLatest(t *testing.T) {
	expectedRates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       63470,
		},
		{
			CoinGeckoID: "ethereum",
			Symbol:      "ETH",
			Price:       1883.48,
		},
	}

	repo := &mockRateRepository{
		getLatestResult: expectedRates,
	}

	rateService := NewRateService(
		repo,
		&mockExchangeClient{},
	)

	rates, err := rateService.GetLatest(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) != 2 {
		t.Fatalf(
			"expected 2 rates, got %d",
			len(rates),
		)
	}

	if rates[0].Symbol != "BTC" {
		t.Errorf(
			"expected BTC, got %s",
			rates[0].Symbol,
		)
	}
}

func TestRateService_GetBySymbol(t *testing.T) {
	expectedRates := []domain.Rate{
		{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       63470,
		},
	}

	repo := &mockRateRepository{
		getBySymbolResult: expectedRates,
	}

	rateService := NewRateService(
		repo,
		&mockExchangeClient{},
	)

	rates, err := rateService.GetBySymbol(
		context.Background(),
		"BTC",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) != 1 {
		t.Fatalf(
			"expected 1 rate, got %d",
			len(rates),
		)
	}

	if rates[0].Symbol != "BTC" {
		t.Errorf(
			"expected BTC, got %s",
			rates[0].Symbol,
		)
	}
}

func TestRateService_GetCurrentRate(t *testing.T) {
	expectedRate := domain.Rate{
		CoinGeckoID: "dogecoin",
		Symbol:      "DOGE",
		Price:       0.1,
		Change1H:    0.5,
		DayLow:      0.09,
		DayHigh:     0.11,
		CreatedAt:   time.Now(),
	}

	repo := &mockRateRepository{}

	client := &mockExchangeClient{
		rates: []domain.Rate{
			expectedRate,
		},
	}

	rateService := NewRateService(
		repo,
		client,
	)

	rate, err := rateService.GetCurrentRate(
		context.Background(),
		"dogecoin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rate == nil {
		t.Fatal("expected rate, got nil")
	}

	if rate.CoinGeckoID != "dogecoin" {
		t.Fatalf(
			"expected CoinGecko ID dogecoin, got %s",
			rate.CoinGeckoID,
		)
	}

	assertIDs(
		t,
		client.requestedIDs,
		[]string{"dogecoin"},
	)

	if len(repo.savedRates) != 1 {
		t.Fatalf(
			"expected 1 saved rate, got %d",
			len(repo.savedRates),
		)
	}

	if repo.savedRates[0].CoinGeckoID != "dogecoin" {
		t.Fatalf(
			"expected saved CoinGecko ID dogecoin, got %s",
			repo.savedRates[0].CoinGeckoID,
		)
	}
}

func assertIDs(
	t *testing.T,
	actual []string,
	expected []string,
) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf(
			"expected IDs %v, got %v",
			expected,
			actual,
		)
	}

	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf(
				"expected IDs %v, got %v",
				expected,
				actual,
			)
		}
	}
}

func TestRateService_GetByCoinGeckoID(t *testing.T) {
	expectedRates := []domain.Rate{
		{
			CoinGeckoID: "solana",
			Symbol:      "SOL",
			Price:       150.25,
		},
	}

	repo := &mockRateRepository{
		getByCoinGeckoIDResult: expectedRates,
	}

	client := &mockExchangeClient{}

	service := NewRateService(
		repo,
		client,
	)

	rates, err := service.GetByCoinGeckoID(
		context.Background(),
		"solana",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) != 1 {
		t.Fatalf(
			"expected 1 rate, got %d",
			len(rates),
		)
	}

	if rates[0].CoinGeckoID != "solana" {
		t.Errorf(
			"expected CoinGeckoID solana, got %s",
			rates[0].CoinGeckoID,
		)
	}

	if rates[0].Symbol != "SOL" {
		t.Errorf(
			"expected symbol SOL, got %s",
			rates[0].Symbol,
		)
	}
}

func TestRateService_GetCurrentRate_ClientError(t *testing.T) {
	expectedErr := errors.New("coingecko unavailable")

	repo := &mockRateRepository{}

	client := &mockExchangeClient{
		err: expectedErr,
	}

	rateService := NewRateService(
		repo,
		client,
	)

	rate, err := rateService.GetCurrentRate(
		context.Background(),
		"dogecoin",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}

	if rate != nil {
		t.Fatalf(
			"expected nil rate, got %+v",
			rate,
		)
	}

	assertIDs(
		t,
		client.requestedIDs,
		[]string{"dogecoin"},
	)

	if len(repo.savedRates) != 0 {
		t.Fatalf(
			"expected 0 saved rates, got %d",
			len(repo.savedRates),
		)
	}
}

func TestRateService_GetCurrentRate_SaveError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &mockRateRepository{
		saveErr: expectedErr,
	}

	client := &mockExchangeClient{
		rates: []domain.Rate{
			{
				CoinGeckoID: "dogecoin",
				Symbol:      "DOGE",
				Price:       0.1,
				Change1H:    0.5,
				DayLow:      0.09,
				DayHigh:     0.11,
				CreatedAt:   time.Now(),
			},
		},
	}

	rateService := NewRateService(
		repo,
		client,
	)

	rate, err := rateService.GetCurrentRate(
		context.Background(),
		"dogecoin",
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}

	if rate != nil {
		t.Fatalf(
			"expected nil rate, got %+v",
			rate,
		)
	}

	assertIDs(
		t,
		client.requestedIDs,
		[]string{"dogecoin"},
	)
}

func TestRateService_GetLatest_ReturnsOnlyBaseCoins(t *testing.T) {
	repo := &mockRateRepository{
		getLatestResult: []domain.Rate{
			{
				CoinGeckoID: "bitcoin",
				Symbol:      "BTC",
				Price:       82879,
			},
			{
				CoinGeckoID: "dogecoin",
				Symbol:      "DOGE",
				Price:       0.09,
			},
			{
				CoinGeckoID: "ethereum",
				Symbol:      "ETH",
				Price:       2642.81,
			},
			{
				CoinGeckoID: "solana",
				Symbol:      "SOL",
				Price:       116.59,
			},
		},
	}

	client := &mockExchangeClient{}

	rateService := NewRateService(
		repo,
		client,
	)

	rates, err := rateService.GetLatest(
		context.Background(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) != 2 {
		t.Fatalf(
			"expected 2 rates, got %d",
			len(rates),
		)
	}

	if rates[0].CoinGeckoID != "bitcoin" {
		t.Errorf(
			"expected bitcoin, got %s",
			rates[0].CoinGeckoID,
		)
	}

	if rates[1].CoinGeckoID != "ethereum" {
		t.Errorf(
			"expected ethereum, got %s",
			rates[1].CoinGeckoID,
		)
	}
}

func TestRateService_UpdateRates_UsesFreshCache(t *testing.T) {
	repo := &mockRateRepository{
		getLatestResult: []domain.Rate{
			{
				CoinGeckoID: "bitcoin",
				Symbol:      "BTC",
				CreatedAt:   time.Now(),
			},
			{
				CoinGeckoID: "ethereum",
				Symbol:      "ETH",
				CreatedAt:   time.Now(),
			},
		},
	}

	client := &mockExchangeClient{}

	rateService := NewRateService(
		repo,
		client,
	)

	err := rateService.UpdateRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.called {
		t.Fatal("expected CoinGecko not to be called")
	}
}

func TestRateService_UpdateRates_UpdatesOnlyStaleCoins(t *testing.T) {
	repo := &mockRateRepository{
		getLatestResult: []domain.Rate{
			{
				CoinGeckoID: "bitcoin",
				Symbol:      "BTC",
				CreatedAt:   time.Now(),
			},
			{
				CoinGeckoID: "ethereum",
				Symbol:      "ETH",
				CreatedAt:   time.Now().Add(-10 * time.Minute),
			},
		},
	}

	client := &mockExchangeClient{
		rates: []domain.Rate{
			{
				CoinGeckoID: "ethereum",
				Symbol:      "ETH",
				Price:       2700,
				CreatedAt:   time.Now(),
			},
		},
	}

	rateService := NewRateService(
		repo,
		client,
	)

	err := rateService.UpdateRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertIDs(
		t,
		client.requestedIDs,
		[]string{"ethereum"},
	)
}

func TestRateService_GetCurrentRate_FallsBackToCachedRate(
	t *testing.T,
) {
	cachedRate := domain.Rate{
		CoinGeckoID: "dogecoin",
		Symbol:      "DOGE",
		Price:       0.10,
		CreatedAt:   time.Now().Add(-10 * time.Minute),
	}

	repo := &mockRateRepository{
		getByCoinGeckoIDResult: []domain.Rate{
			cachedRate,
		},
	}

	client := &mockExchangeClient{
		err: errors.New("unexpected status: 429"),
	}

	rateService := NewRateService(
		repo,
		client,
	)

	rate, err := rateService.GetCurrentRate(
		context.Background(),
		"dogecoin",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rate == nil {
		t.Fatal("expected cached rate, got nil")
	}

	if rate.Price != cachedRate.Price {
		t.Fatalf(
			"expected cached price %.2f, got %.2f",
			cachedRate.Price,
			rate.Price,
		)
	}

	if !client.called {
		t.Fatal("expected CoinGecko to be attempted")
	}
}
