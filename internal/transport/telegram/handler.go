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
	GetByChatID(
		ctx context.Context,
		chatID int64,
	) ([]domain.Subscription, error)

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

	SendMessageWithKeyboard(
		chatID int64,
		text string,
		keyboard tgbotapi.ReplyKeyboardMarkup,
	) error
}

type Handler struct {
	bot                 MessageSender
	rateService         RateService
	subscriptionService SubscriptionService
	coinResolver        CoinResolver
	sessions            *sessionStore
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
		sessions:            newSessionStore(),
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

	session := h.sessions.get(chatID)

	if text != "❌ Отмена" &&
		text != "⬅️ Назад" &&
		!strings.HasPrefix(text, "/") {

		switch session.State {
		case stateWaitingRateCoin:
			h.handleCustomRateCoin(
				ctx,
				chatID,
				text,
			)
			return

		case stateWaitingSubscriptionCoin:
			h.handleSubscriptionCoin(
				ctx,
				chatID,
				text,
			)
			return

		case stateWaitingSubscriptionInterval:
			h.handleSubscriptionInterval(
				ctx,
				chatID,
				text,
			)
			return

		case stateWaitingSubscriptionDelete:
			h.handleSubscriptionDelete(
				ctx,
				chatID,
				text,
			)
			return
		}
	}

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

	case text == "💰 Курсы":
		h.handleRatesMenu(chatID)

	case text == "📋 Подписки":
		h.handleSubscriptions(ctx, chatID)

	case text == "🗑 Удалить подписку":
		h.handleSubscriptionDeleteMenu(
			ctx,
			chatID,
		)

	case text == "🗑 Удалить все":
		h.handleDeleteAllSubscriptions(
			ctx,
			chatID,
		)

	case text == "🔔 Подписаться":
		h.handleSubscriptionMenu(chatID)

	case text == "BTC" ||
		text == "ETH" ||
		text == "SOL" ||
		text == "TON" ||
		text == "DOGE" ||
		text == "ADA":
		h.handleRateBySymbol(
			ctx,
			chatID,
			"/rates "+text,
		)

	case text == "⬅️ Назад":
		h.sessions.clear(chatID)
		h.handleStart(chatID)

	case text == "❌ Отмена":
		h.sessions.clear(chatID)
		h.handleStart(chatID)

	case text == "🔎 Другая монета":
		h.sessions.set(
			chatID,
			userSession{
				State: stateWaitingRateCoin,
			},
		)

		h.sendMessage(
			chatID,
			"Введи тикер или CoinGecko ID криптовалюты.\n\nНапример:\nAVAX\nXMR\nlitecoin",
		)

	default:
		h.sendMessage(
			chatID,
			"Неизвестная команда. Используй /start.",
		)
	}
}

func (h *Handler) handleRatesMenu(
	chatID int64,
) {
	err := h.bot.SendMessageWithKeyboard(
		chatID,
		"Выберите криптовалюту 👇",
		ratesKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send rates keyboard: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptionMenu(
	chatID int64,
) {
	h.sessions.set(
		chatID,
		userSession{
			State: stateWaitingSubscriptionCoin,
		},
	)

	err := h.bot.SendMessageWithKeyboard(
		chatID,
		"Выберите криптовалюту для подписки 👇",
		subscriptionCoinKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send subscription keyboard: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptionCoin(
	ctx context.Context,
	chatID int64,
	text string,
) {
	coin, err := h.coinResolver.ResolveCoin(
		ctx,
		text,
	)
	if err != nil {
		switch {
		case errors.Is(err, coingecko.ErrCoinNotFound):
			h.sendMessage(
				chatID,
				"Не удалось найти такую криптовалюту.\n"+
					"Попробуй другой тикер или CoinGecko ID.",
			)

		case errors.Is(err, coingecko.ErrCoinAmbiguous):
			h.sendMessage(
				chatID,
				"Тикер неоднозначный.\n"+
					"Попробуй указать CoinGecko ID.",
			)

		default:
			log.Printf(
				"resolve subscription coin %q: %v",
				text,
				err,
			)

			h.sendMessage(
				chatID,
				"Не удалось найти криптовалюту. Попробуй ещё раз.",
			)
		}

		return
	}

	h.sessions.set(
		chatID,
		userSession{
			State:                  stateWaitingSubscriptionInterval,
			SubscriptionCoinID:     coin.ID,
			SubscriptionCoinSymbol: coin.Symbol,
		},
	)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		fmt.Sprintf(
			"Выбрана криптовалюта: %s\n\n"+
				"Как часто отправлять курс?",
			coin.Symbol,
		),
		intervalKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send interval keyboard: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptionInterval(
	ctx context.Context,
	chatID int64,
	text string,
) {
	session := h.sessions.get(chatID)

	intervalText := strings.TrimSpace(
		strings.TrimSuffix(text, "мин"),
	)

	interval, err := strconv.Atoi(intervalText)
	if err != nil || interval < 1 {
		h.sendMessage(
			chatID,
			"Выбери интервал кнопкой ниже.",
		)
		return
	}

	subscription := &domain.Subscription{
		ChatID:          chatID,
		Symbol:          session.SubscriptionCoinSymbol,
		CoinGeckoID:     session.SubscriptionCoinID,
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
			"Не удалось создать подписку.",
		)

		return
	}

	h.sessions.clear(chatID)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		fmt.Sprintf(
			"✅ Подписка создана!\n\n"+
				"%s будет отправляться каждые %d мин.",
			session.SubscriptionCoinSymbol,
			interval,
		),
		mainKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send subscription confirmation: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptions(
	ctx context.Context,
	chatID int64,
) {
	subscriptions, err := h.subscriptionService.GetByChatID(
		ctx,
		chatID,
	)
	if err != nil {
		log.Printf(
			"get Telegram subscriptions for chat %d: %v",
			chatID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось получить список подписок.",
		)

		return
	}

	if len(subscriptions) == 0 {
		err := h.bot.SendMessageWithKeyboard(
			chatID,
			"📋 У вас пока нет активных подписок.\n\n"+
				"Нажмите «🔔 Подписаться», чтобы создать первую.",
			mainKeyboard(),
		)
		if err != nil {
			log.Printf(
				"failed to send empty subscriptions message: %v",
				err,
			)
		}

		return
	}

	var builder strings.Builder

	builder.WriteString("📋 Ваши активные подписки:\n\n")

	for _, subscription := range subscriptions {
		if !subscription.Enabled {
			continue
		}

		fmt.Fprintf(
			&builder,
			"🪙 %s\n🔔 Каждые %d мин.\n\n",
			subscription.Symbol,
			subscription.IntervalMinutes,
		)
	}

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		strings.TrimSpace(builder.String()),
		subscriptionsKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send subscriptions message: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptionDeleteMenu(
	ctx context.Context,
	chatID int64,
) {
	subscriptions, err := h.subscriptionService.GetByChatID(
		ctx,
		chatID,
	)
	if err != nil {
		log.Printf(
			"get Telegram subscriptions for delete menu, chat %d: %v",
			chatID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось получить список подписок.",
		)

		return
	}

	var active []domain.Subscription

	for _, subscription := range subscriptions {
		if subscription.Enabled {
			active = append(active, subscription)
		}
	}

	if len(active) == 0 {
		h.sessions.clear(chatID)

		err := h.bot.SendMessageWithKeyboard(
			chatID,
			"У вас нет активных подписок.",
			mainKeyboard(),
		)
		if err != nil {
			log.Printf(
				"failed to send empty delete subscriptions message: %v",
				err,
			)
		}

		return
	}

	h.sessions.set(
		chatID,
		userSession{
			State: stateWaitingSubscriptionDelete,
		},
	)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		"Какую подписку удалить? 👇",
		subscriptionDeleteKeyboard(active),
	)
	if err != nil {
		log.Printf(
			"failed to send subscription delete keyboard: %v",
			err,
		)
	}
}

func (h *Handler) handleSubscriptionDelete(
	ctx context.Context,
	chatID int64,
	text string,
) {
	subscriptions, err := h.subscriptionService.GetByChatID(
		ctx,
		chatID,
	)
	if err != nil {
		log.Printf(
			"get Telegram subscriptions before delete, chat %d: %v",
			chatID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось получить список подписок.",
		)

		return
	}

	symbol := strings.ToUpper(
		strings.TrimSpace(text),
	)

	var selected *domain.Subscription

	for i := range subscriptions {
		subscription := &subscriptions[i]

		if subscription.Enabled &&
			strings.EqualFold(subscription.Symbol, symbol) {

			selected = subscription
			break
		}
	}

	if selected == nil {
		h.sendMessage(
			chatID,
			"Такой активной подписки нет. Выберите монету из списка.",
		)

		return
	}

	err = h.subscriptionService.Delete(
		ctx,
		chatID,
		selected.CoinGeckoID,
	)
	if err != nil {
		log.Printf(
			"delete Telegram subscription %s: %v",
			selected.CoinGeckoID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось удалить подписку.",
		)

		return
	}

	h.sessions.clear(chatID)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		fmt.Sprintf(
			"✅ Подписка %s удалена.",
			selected.Symbol,
		),
		mainKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send subscription delete confirmation: %v",
			err,
		)
	}
}

func (h *Handler) handleDeleteAllSubscriptions(
	ctx context.Context,
	chatID int64,
) {
	err := h.subscriptionService.DeleteAll(
		ctx,
		chatID,
	)
	if err != nil {
		log.Printf(
			"delete all Telegram subscriptions for chat %d: %v",
			chatID,
			err,
		)

		h.sendMessage(
			chatID,
			"Не удалось удалить подписки.",
		)

		return
	}

	h.sessions.clear(chatID)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		"✅ Все подписки удалены.",
		mainKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send delete all confirmation: %v",
			err,
		)
	}
}

func (h *Handler) handleStart(chatID int64) {
	err := h.bot.SendMessageWithKeyboard(
		chatID,
		"Привет! Я бот для отслеживания курсов криптовалют.\n\n"+
			"Выбери нужное действие в меню ниже 👇",
		mainKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send Telegram message with keyboard: %v",
			err,
		)
	}
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

func (h *Handler) handleCustomRateCoin(
	ctx context.Context,
	chatID int64,
	query string,
) {
	query = strings.TrimSpace(query)

	if query == "" {
		h.sendMessage(
			chatID,
			"Введи тикер или CoinGecko ID криптовалюты.",
		)

		return
	}

	coin, err := h.coinResolver.ResolveCoin(
		ctx,
		query,
	)
	if err != nil {
		switch {
		case errors.Is(err, coingecko.ErrCoinNotFound):
			h.sendMessage(
				chatID,
				"Криптовалюта не найдена.\nПопробуй ввести другой тикер.",
			)

		case errors.Is(err, coingecko.ErrCoinAmbiguous):
			h.sendMessage(
				chatID,
				"Тикер неоднозначен.\nПопробуй указать CoinGecko ID монеты.",
			)

		default:
			log.Printf(
				"resolve custom rate coin %q: %v",
				query,
				err,
			)

			h.sendMessage(
				chatID,
				"Не удалось найти криптовалюту.",
			)
		}

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

	h.sessions.clear(chatID)

	err = h.bot.SendMessageWithKeyboard(
		chatID,
		formatRate(*rate),
		ratesKeyboard(),
	)
	if err != nil {
		log.Printf(
			"failed to send rate with keyboard: %v",
			err,
		)
	}
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
