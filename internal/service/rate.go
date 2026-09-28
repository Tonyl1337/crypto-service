package service

import (
	"context"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type RateRepository interface {
	Save(ctx context.Context, rate *domain.Rate) error
	GetLatest(ctx context.Context) ([]domain.Rate, error)
	GetBySymbol(ctx context.Context, symbol string) ([]domain.Rate, error)
	GetByCoinGeckoID(
		ctx context.Context,
		coinGeckoID string,
	) ([]domain.Rate, error)
}

type ExchangeClient interface {
	GetRates(
		ctx context.Context,
		coinGeckoIDs []string,
	) ([]domain.Rate, error)
}

type RateResponse struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	DayLow   float64 `json:"day_low"`
	DayHigh  float64 `json:"day_high"`
	Change1H float64 `json:"change_1h"`
}

type RateService struct {
	repo   RateRepository
	client ExchangeClient
}

func NewRateService(
	repo RateRepository,
	client ExchangeClient,
) *RateService {
	return &RateService{
		repo:   repo,
		client: client,
	}
}

var baseCoinGeckoIDs = []string{
	"bitcoin",
	"ethereum",
}

func (s *RateService) UpdateRates(
	ctx context.Context,
) error {

	rates, err := s.client.GetRates(
		ctx,
		baseCoinGeckoIDs,
	)
	if err != nil {
		return err
	}

	for i := range rates {
		if err := s.repo.Save(ctx, &rates[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *RateService) GetCurrentRate(
	ctx context.Context,
	coinGeckoID string,
) (*domain.Rate, error) {

	rates, err := s.client.GetRates(
		ctx,
		[]string{coinGeckoID},
	)
	if err != nil {
		return nil, err
	}

	if len(rates) == 0 {
		return nil, nil
	}

	rate := rates[0]

	if err := s.repo.Save(ctx, &rate); err != nil {
		return nil, err
	}

	return &rate, nil
}

func (s *RateService) GetByCoinGeckoID(
	ctx context.Context,
	coinGeckoID string,
) ([]domain.Rate, error) {

	return s.repo.GetByCoinGeckoID(
		ctx,
		coinGeckoID,
	)
}

func (s *RateService) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {

	rates, err := s.repo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Rate, 0, len(baseCoinGeckoIDs))

	for _, rate := range rates {
		if isBaseCoinGeckoID(rate.CoinGeckoID) {
			result = append(result, rate)
		}
	}

	return result, nil
}

func (s *RateService) GetBySymbol(
	ctx context.Context,
	symbol string,
) ([]domain.Rate, error) {

	return s.repo.GetBySymbol(ctx, symbol)
}

func isBaseCoinGeckoID(id string) bool {
	for _, baseID := range baseCoinGeckoIDs {
		if id == baseID {
			return true
		}
	}

	return false
}
