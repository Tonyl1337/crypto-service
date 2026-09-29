package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockSubscriptionRepository struct {
	subscription  *domain.Subscription
	subscriptions []domain.Subscription
	coinGeckoIDs  []string

	err error

	chatID          int64
	coinGeckoID     string
	subscriptionArg *domain.Subscription
	sentAt          time.Time

	getByChatIDAndCoinCalled     bool
	getEnabledCalled             bool
	getEnabledCoinGeckoIDsCalled bool
	saveCalled                   bool
	deleteCalled                 bool
	deleteAllCalled              bool
	markSentCalled               bool
}

func (m *mockSubscriptionRepository) GetByChatIDAndCoin(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) (*domain.Subscription, error) {
	m.getByChatIDAndCoinCalled = true
	m.chatID = chatID
	m.coinGeckoID = coinGeckoID

	return m.subscription, m.err
}

func (m *mockSubscriptionRepository) GetEnabled(
	ctx context.Context,
) ([]domain.Subscription, error) {
	m.getEnabledCalled = true

	return m.subscriptions, m.err
}

func (m *mockSubscriptionRepository) GetEnabledCoinGeckoIDs(
	ctx context.Context,
) ([]string, error) {
	m.getEnabledCoinGeckoIDsCalled = true

	return m.coinGeckoIDs, m.err
}

func (m *mockSubscriptionRepository) Save(
	ctx context.Context,
	subscription *domain.Subscription,
) error {
	m.saveCalled = true
	m.subscriptionArg = subscription

	return m.err
}

func (m *mockSubscriptionRepository) Delete(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) error {
	m.deleteCalled = true
	m.chatID = chatID
	m.coinGeckoID = coinGeckoID

	return m.err
}

func (m *mockSubscriptionRepository) DeleteAll(
	ctx context.Context,
	chatID int64,
) error {
	m.deleteAllCalled = true
	m.chatID = chatID

	return m.err
}

func (m *mockSubscriptionRepository) MarkSent(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
	sentAt time.Time,
) error {
	m.markSentCalled = true
	m.chatID = chatID
	m.coinGeckoID = coinGeckoID
	m.sentAt = sentAt

	return m.err
}

func TestSubscriptionService_GetByChatIDAndCoin(t *testing.T) {
	expected := &domain.Subscription{
		ChatID:      123,
		CoinGeckoID: "bitcoin",
		Symbol:      "BTC",
		Enabled:     true,
	}

	repo := &mockSubscriptionRepository{
		subscription: expected,
	}

	service := NewSubscriptionsService(repo)

	actual, err := service.GetByChatIDAndCoin(
		context.Background(),
		123,
		"bitcoin",
	)

	require.NoError(t, err)
	require.True(t, repo.getByChatIDAndCoinCalled)
	require.Equal(t, int64(123), repo.chatID)
	require.Equal(t, "bitcoin", repo.coinGeckoID)
	require.Equal(t, expected, actual)
}

func TestSubscriptionService_GetEnabled(t *testing.T) {
	expected := []domain.Subscription{
		{
			ChatID:      123,
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Enabled:     true,
		},
		{
			ChatID:      456,
			CoinGeckoID: "dogecoin",
			Symbol:      "DOGE",
			Enabled:     true,
		},
	}

	repo := &mockSubscriptionRepository{
		subscriptions: expected,
	}

	service := NewSubscriptionsService(repo)

	actual, err := service.GetEnabled(
		context.Background(),
	)

	require.NoError(t, err)
	require.True(t, repo.getEnabledCalled)
	require.Equal(t, expected, actual)
}

func TestSubscriptionService_GetEnabledCoinGeckoIDs(t *testing.T) {
	expected := []string{
		"bitcoin",
		"dogecoin",
		"solana",
	}

	repo := &mockSubscriptionRepository{
		coinGeckoIDs: expected,
	}

	service := NewSubscriptionsService(repo)

	actual, err := service.GetEnabledCoinGeckoIDs(
		context.Background(),
	)

	require.NoError(t, err)
	require.True(
		t,
		repo.getEnabledCoinGeckoIDsCalled,
	)
	require.Equal(t, expected, actual)
}

func TestSubscriptionService_Save(t *testing.T) {
	repo := &mockSubscriptionRepository{}

	service := NewSubscriptionsService(repo)

	subscription := &domain.Subscription{
		ChatID:          123,
		CoinGeckoID:     "solana",
		Symbol:          "SOL",
		Enabled:         true,
		IntervalMinutes: 5,
	}

	err := service.Save(
		context.Background(),
		subscription,
	)

	require.NoError(t, err)
	require.True(t, repo.saveCalled)
	require.Same(
		t,
		subscription,
		repo.subscriptionArg,
	)
}

func TestSubscriptionService_Delete(t *testing.T) {
	repo := &mockSubscriptionRepository{}

	service := NewSubscriptionsService(repo)

	err := service.Delete(
		context.Background(),
		123,
		"dogecoin",
	)

	require.NoError(t, err)
	require.True(t, repo.deleteCalled)
	require.Equal(t, int64(123), repo.chatID)
	require.Equal(
		t,
		"dogecoin",
		repo.coinGeckoID,
	)
}

func TestSubscriptionService_DeleteAll(t *testing.T) {
	repo := &mockSubscriptionRepository{}

	service := NewSubscriptionsService(repo)

	err := service.DeleteAll(
		context.Background(),
		777,
	)

	require.NoError(t, err)
	require.True(t, repo.deleteAllCalled)
	require.Equal(t, int64(777), repo.chatID)
}

func TestSubscriptionService_MarkSent(t *testing.T) {
	repo := &mockSubscriptionRepository{}

	service := NewSubscriptionsService(repo)

	sentAt := time.Date(
		2026,
		time.September,
		28,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	err := service.MarkSent(
		context.Background(),
		123,
		"solana",
		sentAt,
	)

	require.NoError(t, err)
	require.True(t, repo.markSentCalled)
	require.Equal(t, int64(123), repo.chatID)
	require.Equal(t, "solana", repo.coinGeckoID)
	require.Equal(t, sentAt, repo.sentAt)
}

func TestSubscriptionService_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database unavailable")

	repo := &mockSubscriptionRepository{
		err: expectedErr,
	}

	service := NewSubscriptionsService(repo)

	err := service.Delete(
		context.Background(),
		123,
		"bitcoin",
	)

	require.ErrorIs(t, err, expectedErr)
	require.True(t, repo.deleteCalled)
}
