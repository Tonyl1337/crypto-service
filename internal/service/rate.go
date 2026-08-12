package service

import (
	"context"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type RateRepository interface {
	Save(ctx context.Context, rate *domain.Rate) error
	GetLatest(ctx context.Context) ([]domain.Rate, error)
	GetBySymbol(ctx context.Context, symbol string) ([]domain.Rate, error)
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

type ExchangeClient interface {
	GetRates(ctx context.Context) ([]domain.Rate, error)
}

func NewRateService(
	repo RateRepository,
	client ExchangeClient,
) *RateService {

	return &RateService{
		repo: repo,
		client: client,
	}
}

func (s *RateService) UpdateRates(
	ctx context.Context,
) error {

	rates, err := s.client.GetRates(ctx)
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

func (s *RateService) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {

	return s.repo.GetLatest(ctx)
}

func (s *RateService) GetBySymbol(
	ctx context.Context,
	symbol string,
) ([]domain.Rate, error) {

	return s.repo.GetBySymbol(ctx, symbol)
}