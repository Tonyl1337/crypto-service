package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockSubscriptionProvider struct {
	subscriptions []domain.Subscription
	getErr        error

	markedChatID int64
	markedAt     time.Time
	markCalled   bool
	markErr      error
}

func (m *mockSubscriptionProvider) GetEnabled(
	ctx context.Context,
) ([]domain.Subscription, error) {
	return m.subscriptions, m.getErr
}

func (m *mockSubscriptionProvider) MarkSent(
	ctx context.Context,
	chatID int64,
	sentAt time.Time,
) error {
	m.markCalled = true
	m.markedChatID = chatID
	m.markedAt = sentAt

	return m.markErr
}

type mockRateProvider struct {
	rates []domain.Rate
	err   error
}

func (m *mockRateProvider) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {
	return m.rates, m.err
}

type mockMessageSender struct {
	called bool
	chatID int64
	text   string
	err    error
}

func (m *mockMessageSender) SendMessage(
	chatID int64,
	text string,
) error {
	m.called = true
	m.chatID = chatID
	m.text = text

	return m.err
}

func TestSubscriptionSender_Process_SendsMessageWhenDue(
	t *testing.T,
) {
	lastSent := time.Now().Add(-2 * time.Minute)

	subscriptions := &mockSubscriptionProvider{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Enabled:         true,
				IntervalMinutes: 1,
				LastSentAt:      &lastSent,
			},
		},
	}

	rates := &mockRateProvider{
		rates: []domain.Rate{
			{
				Symbol:   "BTC",
				Price:    63000,
				DayLow:   62000,
				DayHigh:  64000,
				Change1H: 0.5,
			},
		},
	}

	sender := &mockMessageSender{}

	subscriptionSender := NewSubscriptionSender(
		subscriptions,
		rates,
		sender,
		time.Second,
	)

	subscriptionSender.process(context.Background())

	if !sender.called {
		t.Fatal("expected message to be sent")
	}

	if sender.chatID != 123 {
		t.Fatalf(
			"expected chat ID 123, got %d",
			sender.chatID,
		)
	}

	if !subscriptions.markCalled {
		t.Fatal("expected MarkSent to be called")
	}
}

func TestSubscriptionSender_Process_DoesNotSendBeforeInterval(
	t *testing.T,
) {
	lastSent := time.Now().Add(-30 * time.Second)

	subscriptions := &mockSubscriptionProvider{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Enabled:         true,
				IntervalMinutes: 1,
				LastSentAt:      &lastSent,
			},
		},
	}

	rates := &mockRateProvider{}

	sender := &mockMessageSender{}

	subscriptionSender := NewSubscriptionSender(
		subscriptions,
		rates,
		sender,
		time.Second,
	)

	subscriptionSender.process(context.Background())

	if sender.called {
		t.Fatal("expected message not to be sent")
	}

	if subscriptions.markCalled {
		t.Fatal("expected MarkSent not to be called")
	}
}

func TestSubscriptionSender_Process_DoesNotMarkSentOnSendError(
	t *testing.T,
) {
	lastSent := time.Now().Add(-2 * time.Minute)

	subscriptions := &mockSubscriptionProvider{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Enabled:         true,
				IntervalMinutes: 1,
				LastSentAt:      &lastSent,
			},
		},
	}

	rates := &mockRateProvider{
		rates: []domain.Rate{
			{
				Symbol: "BTC",
				Price:  63000,
			},
		},
	}

	sender := &mockMessageSender{
		err: errors.New("telegram unavailable"),
	}

	subscriptionSender := NewSubscriptionSender(
		subscriptions,
		rates,
		sender,
		time.Second,
	)

	subscriptionSender.process(context.Background())

	if !sender.called {
		t.Fatal("expected SendMessage to be called")
	}

	if subscriptions.markCalled {
		t.Fatal(
			"expected MarkSent not to be called after send error",
		)
	}
}

func TestSubscriptionSender_Process_GetSubscriptionsError(
	t *testing.T,
) {
	subscriptions := &mockSubscriptionProvider{
		getErr: errors.New("database unavailable"),
	}

	rates := &mockRateProvider{}
	sender := &mockMessageSender{}

	subscriptionSender := NewSubscriptionSender(
		subscriptions,
		rates,
		sender,
		time.Second,
	)

	subscriptionSender.process(context.Background())

	if sender.called {
		t.Fatal("expected SendMessage not to be called")
	}

	if subscriptions.markCalled {
		t.Fatal("expected MarkSent not to be called")
	}
}
