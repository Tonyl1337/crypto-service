package service

import (
	"context"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type RateRepository interface {
	Save(ctx context.Context, rate *domain.Rate) error
	GetLatest(ctx context.Context) ([]domain.Rate, error)
	GetBySymbol(ctx context.Context, symbol string) ([]domain.Rate, error)
}

type RateService struct {
	repo   RateRepository
	client ExchangeClient
}

type ExchangeClient interface {
	GetPrices(ctx context.Context) (coingecko.PriceResponse, error)
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

func (s *RateService) UpdateRates(ctx context.Context) error {

	prices, err := s.client.GetPrices(ctx)
	if err != nil {
		return err
	}

	for symbol, coin := range prices {

		rate := &domain.Rate{
			Symbol: coingecko.NormalizeSymbol(symbol),
			Price:    coin.Price,
			Change24H: coin.Change24H,
			DayHigh:  coin.High24H,
			DayLow:   coin.Low24H,
		}

		if err := s.repo.Save(ctx, rate); err != nil {
			return err
		}
	}

	return nil
}