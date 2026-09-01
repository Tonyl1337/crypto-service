package service

import (
	"context"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type SubscriptionRepository interface {
	GetByChatID(ctx context.Context, chatID int64) (*domain.Subscription, error)

	GetEnabled(ctx context.Context) ([]domain.Subscription, error)

	Save(ctx context.Context, subscription *domain.Subscription) error

	Delete(ctx context.Context, chatID int64) error

	MarkSent(ctx context.Context, chatID int64, sentAt time.Time) error
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

func (s *SubscriptionService) GetByChatID( // для определения айдишника чата
	ctx context.Context,
	chatID int64,
) (*domain.Subscription, error) {

	return s.repo.GetByChatID(ctx, chatID)
}

func (s *SubscriptionService) GetEnabled( //для включения подписки
	ctx context.Context,
) ([]domain.Subscription, error) {

	return s.repo.GetEnabled(ctx)
}

func (s *SubscriptionService) Save( // для сохранения подписки
	ctx context.Context,
	subscription *domain.Subscription,
) error {

	return s.repo.Save(ctx, subscription)
}

func (s *SubscriptionService) Delete( // для удаления подписки
	ctx context.Context,
	chatID int64,
) error {

	return s.repo.Delete(ctx, chatID)
}

func (s *SubscriptionService) MarkSent(
	ctx context.Context,
	chatID int64,
	sentAt time.Time,
) error {
	return s.repo.MarkSent(
		ctx,
		chatID,
		sentAt,
	)
}
