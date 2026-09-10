package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Tonyl1337/crypto-service/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RateService interface {
	GetLatest(ctx context.Context) ([]domain.Rate, error)
	GetBySymbol(
		ctx context.Context,
		symbol string,
	) ([]domain.Rate, error)
}

type SubscriptionService interface {
	Save(
		ctx context.Context,
		subscription *domain.Subscription,
	) error

	Delete(
		ctx context.Context,
		chatID int64,
	) error
}

type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

type Handler struct {
	bot                 MessageSender
	rateService         RateService
	subscriptionService SubscriptionService
}

func NewHandler(
	bot MessageSender,
	rateService RateService,
	subscriptionService SubscriptionService,
) *Handler {
	return &Handler{
		bot:                 bot,
		rateService:         rateService,
		subscriptionService: subscriptionService,
	}
}

func (h *Handler) Handle(
	ctx context.Context,
	update tgbotapi.Update,
) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	log.Printf(
		"Telegram message from chat %d: %s",
		chatID,
		text,
	)

	switch {
	case text == "/start":
		h.handleStart(chatID)

	case text == "/rates":
		h.handleRates(ctx, chatID)

	case strings.HasPrefix(text, "/rates "):
		h.handleRateBySymbol(ctx, chatID, text)

	case strings.HasPrefix(text, "/start_auto ") ||
		strings.HasPrefix(text, "/start-auto "):
		h.handleStartAuto(ctx, chatID, text)

	case text == "/stop_auto" || text == "/stop-auto":
		h.handleStopAuto(ctx, chatID)

	default:
		h.sendMessage(
			chatID,
			"Неизвестная команда. Используй /start.",
		)
	}
}

func (h *Handler) handleStart(chatID int64) {
	h.sendMessage(
		chatID,
		"Привет! Я бот для отслеживания курсов криптовалют.\n\n"+
			"Доступные команды:\n"+
			"/rates — текущие курсы\n"+
			"/rates BTC — курс Bitcoin\n"+
			"/rates ETH — курс Ethereum\n"+
			"/start_auto 10 — отправлять курсы каждые 10 минут\n"+
			"/stop_auto — отключить автоматическую отправку",
	)
}

func (h *Handler) handleRates(
	ctx context.Context,
	chatID int64,
) {
	rates, err := h.rateService.GetLatest(ctx)
	if err != nil {
		log.Printf("get latest rates: %v", err)

		h.sendMessage(
			chatID,
			"Не удалось получить курсы.",
		)

		return
	}

	if len(rates) == 0 {
		h.sendMessage(
			chatID,
			"Данные о курсах пока отсутствуют.",
		)

		return
	}

	var builder strings.Builder

	builder.WriteString("Текущие курсы:\n\n")

	for _, rate := range rates {
		builder.WriteString(formatRate(rate))
		builder.WriteString("\n")
	}

	h.sendMessage(
		chatID,
		builder.String(),
	)
}

func formatRate(rate domain.Rate) string {
	return fmt.Sprintf(
		"%s\n"+
			"Цена: $%.2f\n"+
			"Минимум за 24ч: $%.2f\n"+
			"Максимум за 24ч: $%.2f\n"+
			"Изменение за 1ч: %.2f%%\n",
		rate.Symbol,
		rate.Price,
		rate.DayLow,
		rate.DayHigh,
		rate.Change1H,
	)
}

func (h *Handler) handleRateBySymbol(
	ctx context.Context,
	chatID int64,
	text string,
) {
	parts := strings.Fields(text)

	if len(parts) != 2 {
		h.sendMessage(
			chatID,
			"Использование: /rates BTC или /rates ETH",
		)

		return
	}

	symbol := strings.ToUpper(parts[1])

	if symbol != "BTC" && symbol != "ETH" {
		h.sendMessage(
			chatID,
			"Поддерживаются только BTC и ETH.",
		)

		return
	}

	rates, err := h.rateService.GetBySymbol(
		ctx,
		symbol,
	)
	if err != nil {
		log.Printf(
			"get rate by symbol %s: %v",
			symbol,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось получить курс.",
		)

		return
	}

	if len(rates) == 0 {
		h.sendMessage(
			chatID,
			"Данные по этой валюте отсутствуют.",
		)

		return
	}

	h.sendMessage(
		chatID,
		formatRate(rates[0]),
	)
}

func (h *Handler) handleStartAuto(
	ctx context.Context,
	chatID int64,
	text string,
) {
	parts := strings.Fields(text)

	if len(parts) != 2 {
		h.sendMessage(
			chatID,
			"Использование: /start_auto 10",
		)

		return
	}

	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		h.sendMessage(
			chatID,
			"Интервал должен быть числом.",
		)

		return
	}

	if interval < 1 {
		h.sendMessage(
			chatID,
			"Интервал должен быть не меньше 1 минуты.",
		)

		return
	}

	subscription := &domain.Subscription{
		ChatID:          chatID,
		Enabled:         true,
		IntervalMinutes: interval,
	}

	err = h.subscriptionService.Save(
		ctx,
		subscription,
	)
	if err != nil {
		log.Printf(
			"save Telegram subscription: %v",
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось включить автоматическую отправку.",
		)

		return
	}

	h.sendMessage(
		chatID,
		fmt.Sprintf(
			"Автоматическая отправка включена каждые %d мин.",
			interval,
		),
	)
}

func (h *Handler) handleStopAuto(
	ctx context.Context,
	chatID int64,
) {
	err := h.subscriptionService.Delete(
		ctx,
		chatID,
	)
	if err != nil {
		log.Printf(
			"delete Telegram subscription: %v",
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось отключить автоматическую отправку.",
		)

		return
	}

	h.sendMessage(
		chatID,
		"Автоматическая отправка отключена.",
	)
}

func (h *Handler) sendMessage(
	chatID int64,
	text string,
) {
	if err := h.bot.SendMessage(chatID, text); err != nil {
		log.Printf(
			"send Telegram message: %v",
			err,
		)
	}
}
