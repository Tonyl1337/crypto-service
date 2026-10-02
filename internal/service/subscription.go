package service

import (
	"context"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type SubscriptionRepository interface {
	GetByChatIDAndCoin(
		ctx context.Context,
		chatID int64,
		coinGeckoID string,
	) (*domain.Subscription, error)

	GetByChatID(
		ctx context.Context,
		chatID int64,
	) ([]domain.Subscription, error)

	GetEnabled(
		ctx context.Context,
	) ([]domain.Subscription, error)

	GetEnabledCoinGeckoIDs(
		ctx context.Context,
	) ([]string, error)

	Save(
		ctx context.Context,
		subscription *domain.Subscription,
	) error

	Delete(
		ctx context.Context,
		chatID int64,
		coinGeckoID string,
	) error

	DeleteAll(
		ctx context.Context,
		chatID int64,
	) error

	MarkSent(
		ctx context.Context,
		chatID int64,
		coinGeckoID string,
		sentAt time.Time,
	) error
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionsService(
	repo SubscriptionRepository,
) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) GetByChatIDAndCoin(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) (*domain.Subscription, error) {
	return s.repo.GetByChatIDAndCoin(
		ctx,
		chatID,
		coinGeckoID,
	)
}

func (s *SubscriptionService) GetByChatID(
	ctx context.Context,
	chatID int64,
) ([]domain.Subscription, error) {
	return s.repo.GetByChatID(
		ctx,
		chatID,
	)
}

func (s *SubscriptionService) GetEnabled(
	ctx context.Context,
) ([]domain.Subscription, error) {
	return s.repo.GetEnabled(ctx)
}

func (s *SubscriptionService) GetEnabledCoinGeckoIDs(
	ctx context.Context,
) ([]string, error) {
	return s.repo.GetEnabledCoinGeckoIDs(ctx)
}

func (s *SubscriptionService) Save(
	ctx context.Context,
	subscription *domain.Subscription,
) error {
	return s.repo.Save(ctx, subscription)
}

func (s *SubscriptionService) Delete(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) error {
	return s.repo.Delete(
		ctx,
		chatID,
		coinGeckoID,
	)
}

func (s *SubscriptionService) DeleteAll(
	ctx context.Context,
	chatID int64,
) error {
	return s.repo.DeleteAll(
		ctx,
		chatID,
	)
}

func (s *SubscriptionService) MarkSent(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
	sentAt time.Time,
) error {
	return s.repo.MarkSent(
		ctx,
		chatID,
		coinGeckoID,
		sentAt,
	)
}
