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