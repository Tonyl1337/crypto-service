package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RateService interface {
	GetLatest(ctx context.Context) ([]domain.Rate, error)

	GetBySymbol(
		ctx context.Context,
		symbol string,
	) ([]domain.Rate, error)

	GetByCoinGeckoID(
		ctx context.Context,
		coinGeckoID string,
	) ([]domain.Rate, error)

	GetCurrentRate(
		ctx context.Context,
		coinGeckoID string,
	) (*domain.Rate, error)
}

type CoinResolver interface {
	ResolveCoin(
		ctx context.Context,
		query string,
	) (*domain.Coin, error)
}

type SubscriptionService interface {
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
}

type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

type Handler struct {
	bot                 MessageSender
	rateService         RateService
	subscriptionService SubscriptionService
	coinResolver        CoinResolver
}

func NewHandler(
	bot MessageSender,
	rateService RateService,
	subscriptionService SubscriptionService,
	coinResolver CoinResolver,
) *Handler {
	return &Handler{
		bot:                 bot,
		rateService:         rateService,
		subscriptionService: subscriptionService,
		coinResolver:        coinResolver,
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

	case text == "/stop_auto" ||
		text == "/stop-auto" ||
		strings.HasPrefix(text, "/stop_auto ") ||
		strings.HasPrefix(text, "/stop-auto "):
		h.handleStopAuto(ctx, chatID, text)

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
			"/rates — текущие курсы BTC и ETH\n"+
			"/rates BTC — курс выбранной криптовалюты\n"+
			"/rates SOL — можно использовать другие монеты\n"+
			"/start_auto SOL 10 — отправлять курс SOL каждые 10 минут\n"+
			"/stop_auto SOL — отключить отправку SOL\n"+
			"/stop_auto — отключить все автоматические отправки",
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
			"Использование: /rates BTC",
		)

		return
	}

	query := parts[1]

	coin, err := h.coinResolver.ResolveCoin(
		ctx,
		query,
	)
	if err != nil {
		log.Printf(
			"resolve coin %s: %v",
			query,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось найти такую криптовалюту.",
		)

		return
	}

	rate, err := h.rateService.GetCurrentRate(
		ctx,
		coin.ID,
	)
	if err != nil {
		log.Printf(
			"get current rate %s: %v",
			coin.ID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось получить курс.",
		)

		return
	}

	if rate == nil {
		h.sendMessage(
			chatID,
			fmt.Sprintf(
				"CoinGecko не вернул курс %s.",
				coin.Symbol,
			),
		)

		return
	}

	h.sendMessage(
		chatID,
		formatRate(*rate),
	)
}

func (h *Handler) handleStartAuto(
	ctx context.Context,
	chatID int64,
	text string,
) {
	parts := strings.Fields(text)

	if len(parts) != 3 {
		h.sendMessage(
			chatID,
			"Использование: /start_auto SOL 10",
		)

		return
	}

	coinQuery := parts[1]

	interval, err := strconv.Atoi(parts[2])
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

	coin, err := h.coinResolver.ResolveCoin(
		ctx,
		coinQuery,
	)
	if err != nil {
		switch {
		case errors.Is(err, coingecko.ErrCoinNotFound):
			h.sendMessage(
				chatID,
				"Криптовалюта не найдена.",
			)

		case errors.Is(err, coingecko.ErrCoinAmbiguous):
			h.sendMessage(
				chatID,
				"Тикер неоднозначен. Укажи CoinGecko ID монеты.",
			)

		default:
			log.Printf(
				"resolve coin %q: %v",
				coinQuery,
				err,
			)

			h.sendMessage(
				chatID,
				"Не удалось найти криптовалюту.",
			)
		}

		return
	}

	subscription := &domain.Subscription{
		ChatID:          chatID,
		Symbol:          coin.Symbol,
		CoinGeckoID:     coin.ID,
		Enabled:         true,
		IntervalMinutes: interval,
	}

	if err := h.subscriptionService.Save(
		ctx,
		subscription,
	); err != nil {
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
			"Автоматическая отправка %s включена каждые %d мин.",
			coin.Symbol,
			interval,
		),
	)
}

func (h *Handler) handleStopAuto(
	ctx context.Context,
	chatID int64,
	text string,
) {
	parts := strings.Fields(text)

	switch len(parts) {
	case 1:
		if err := h.subscriptionService.DeleteAll(
			ctx,
			chatID,
		); err != nil {
			log.Printf(
				"delete all Telegram subscriptions: %v",
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
			"Все автоматические отправки отключены.",
		)

	case 2:
		coinQuery := parts[1]

		coin, err := h.coinResolver.ResolveCoin(
			ctx,
			coinQuery,
		)
		if err != nil {
			switch {
			case errors.Is(err, coingecko.ErrCoinNotFound):
				h.sendMessage(
					chatID,
					"Криптовалюта не найдена.",
				)

			case errors.Is(err, coingecko.ErrCoinAmbiguous):
				h.sendMessage(
					chatID,
					"Тикер неоднозначен. Укажи CoinGecko ID монеты.",
				)

			default:
				log.Printf(
					"resolve coin %q: %v",
					coinQuery,
					err,
				)

				h.sendMessage(
					chatID,
					"Не удалось найти криптовалюту.",
				)
			}

			return
		}

		if err := h.subscriptionService.Delete(
			ctx,
			chatID,
			coin.ID,
		); err != nil {
			log.Printf(
				"delete Telegram subscription %s: %v",
				coin.ID,
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
			fmt.Sprintf(
				"Автоматическая отправка %s отключена.",
				coin.Symbol,
			),
		)

	default:
		h.sendMessage(
			chatID,
			"Использование: /stop_auto или /stop_auto SOL",
		)
	}
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
