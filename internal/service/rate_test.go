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

	getLatestResult   []domain.Rate
	getBySymbolResult []domain.Rate

	saveErr        error
	getLatestErr   error
	getBySymbolErr error
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

type mockExchangeClient struct {
	rates []domain.Rate
	err   error
}

func (m *mockExchangeClient) GetRates(
	ctx context.Context,
) ([]domain.Rate, error) {
	return m.rates, m.err
}

func TestRateService_UpdateRates(t *testing.T) {
	now := time.Now()

	rates := []domain.Rate{
		{
			Symbol:    "BTC",
			Price:     63470,
			Change1H:  0.1,
			DayLow:    63267,
			DayHigh:   64329,
			CreatedAt: now,
		},
		{
			Symbol:    "ETH",
			Price:     1883.48,
			Change1H:  -0.3,
			DayLow:    1876.31,
			DayHigh:   1918.99,
			CreatedAt: now,
		},
	}

	repo := &mockRateRepository{}
	client := &mockExchangeClient{
		rates: rates,
	}

	service := NewRateService(repo, client)

	err := service.UpdateRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.savedRates) != 2 {
		t.Fatalf(
			"expected 2 saved rates, got %d",
			len(repo.savedRates),
		)
	}

	if repo.savedRates[0].Symbol != "BTC" {
		t.Errorf(
			"expected first symbol BTC, got %s",
			repo.savedRates[0].Symbol,
		)
	}

	if repo.savedRates[1].Symbol != "ETH" {
		t.Errorf(
			"expected second symbol ETH, got %s",
			repo.savedRates[1].Symbol,
		)
	}
}

func TestRateService_UpdateRates_ClientError(t *testing.T) {
	expectedErr := errors.New("coin gecko error")

	repo := &mockRateRepository{}

	client := &mockExchangeClient{
		err: expectedErr,
	}

	service := NewRateService(repo, client)

	err := service.UpdateRates(context.Background())
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

func TestRateService_UpdateRates_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	rates := []domain.Rate{
		{
			Symbol:    "BTC",
			Price:     63470,
			Change1H:  0.1,
			DayLow:    63267,
			DayHigh:   64329,
			CreatedAt: time.Now(),
		},
	}

	repo := &mockRateRepository{
		saveErr: expectedErr,
	}

	client := &mockExchangeClient{
		rates: rates,
	}

	service := NewRateService(repo, client)

	err := service.UpdateRates(context.Background())
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
			Symbol: "BTC",
			Price:  63470,
		},
		{
			Symbol: "ETH",
			Price:  1883.48,
		},
	}

	repo := &mockRateRepository{
		getLatestResult: expectedRates,
	}

	client := &mockExchangeClient{}

	service := NewRateService(repo, client)

	rates, err := service.GetLatest(context.Background())
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
			Symbol: "BTC",
			Price:  63470,
		},
	}

	repo := &mockRateRepository{
		getBySymbolResult: expectedRates,
	}

	client := &mockExchangeClient{}

	service := NewRateService(repo, client)

	rates, err := service.GetBySymbol(
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
