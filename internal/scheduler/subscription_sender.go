package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

// Один общий SubscriptionSender раз в минуту читает активные подписки и решает
// кому уже пора отправлять сообщение

type SubscriptionProvider interface {
	GetEnabled(
		ctx context.Context,
	) ([]domain.Subscription, error)

	MarkSent(
		ctx context.Context,
		chatID int64,
		sentAT time.Time,
	) error
}

type RateProvider interface {
	GetLatest(
		ctx context.Context,
	) ([]domain.Rate, error)
}

type MessageSender interface {
	SendMessage(
		chatID int64,
		text string,
	) error
}

type SubscriptionSender struct {
	subscriptions SubscriptionProvider
	rates         RateProvider
	sender        MessageSender
	checkInterval time.Duration
}

func NewSubscriptionSender(
	subscriptions SubscriptionProvider,
	rates RateProvider,
	sender MessageSender,
	checkInterval time.Duration,
) *SubscriptionSender {

	return &SubscriptionSender{
		subscriptions: subscriptions,
		rates:         rates,
		sender:        sender,
		checkInterval: checkInterval,
	}
}

func (s *SubscriptionSender) process(
	ctx context.Context,
) {

	subscriptions, err := s.subscriptions.GetEnabled(ctx)
	if err != nil {
		log.Printf(
			"get enabled subscriptions: %v",
			err,
		)
		return
	}

	now := time.Now()

	for _, subscription := range subscriptions {

		if !shouldSend(subscription, now) {
			continue
		}

		rates, err := s.rates.GetLatest(ctx)
		if err != nil {
			log.Printf(
				"get rates for Telegram subscription: %v",
				err,
			)
			continue
		}

		if len(rates) == 0 {
			continue
		}

		message := formatRates(rates)

		if err := s.sender.SendMessage(
			subscription.ChatID,
			message,
		); err != nil {

			log.Printf(
				"send automatic Telegram message: %v",
				err,
			)

			continue
		}

		if err := s.subscriptions.MarkSent(
			ctx,
			subscription.ChatID,
			now,
		); err != nil {

			log.Printf(
				"mark Telegram subscription sent: %v",
				err,
			)
		}
	}
}

func (s *SubscriptionSender) Start(
	ctx context.Context,
) {

	ticker := time.NewTicker(s.checkInterval)

	go func() {
		defer ticker.Stop()

		for {
			select {

			case <-ticker.C:
				s.process(ctx)

			case <-ctx.Done():
				return
			}
		}
	}()
}

func shouldSend(
	subscription domain.Subscription,
	now time.Time,
) bool {

	interval := time.Duration(
		subscription.IntervalMinutes,
	) * time.Minute

	lastActivity := subscription.UpdatedAt

	if subscription.LastSentAt != nil {
		lastActivity = *subscription.LastSentAt
	}

	return now.Sub(lastActivity) >= interval
}

func formatRates(
	rates []domain.Rate,
) string {

	var builder strings.Builder

	builder.WriteString(
		"Автоматическое обновление курсов:\n\n",
	)

	for _, rate := range rates {

		builder.WriteString(
			fmt.Sprintf(
				"%s\n"+
					"Цена: $%.2f\n"+
					"Минимум за 24ч: $%.2f\n"+
					"Максимум за 24ч: $%.2f\n"+
					"Изменение за 1ч: %.2f%%\n\n",
				rate.Symbol,
				rate.Price,
				rate.DayLow,
				rate.DayHigh,
				rate.Change1H,
			),
		)
	}

	return builder.String()
}
