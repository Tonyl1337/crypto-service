package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockBot struct {
	chatID   int64
	text     string
	keyboard *tgbotapi.ReplyKeyboardMarkup
	err      error
}

func (m *mockBot) SendMessage(
	chatID int64,
	text string,
) error {
	m.chatID = chatID
	m.text = text

	return m.err
}

func (m *mockBot) SendMessageWithKeyboard(
	chatID int64,
	text string,
	keyboard tgbotapi.ReplyKeyboardMarkup,
) error {
	m.chatID = chatID
	m.text = text
	m.keyboard = &keyboard

	return m.err
}

type mockRateService struct {
	rates []domain.Rate
	err   error

	getByCoinGeckoIDResult []domain.Rate
	getByCoinGeckoIDErr    error

	requestedSymbol      string
	requestedCoinGeckoID string

	currentRate       *domain.Rate
	currentRateErr    error
	currentRateCoinID string
}

func (m *mockRateService) GetCurrentRate(
	ctx context.Context,
	coinGeckoID string,
) (*domain.Rate, error) {
	m.currentRateCoinID = coinGeckoID

	return m.currentRate, m.currentRateErr
}

func (m *mockRateService) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {
	return m.rates, m.err
}

func (m *mockRateService) GetBySymbol(
	ctx context.Context,
	symbol string,
) ([]domain.Rate, error) {
	m.requestedSymbol = symbol

	return m.rates, m.err
}

func (m *mockRateService) GetByCoinGeckoID(
	ctx context.Context,
	coinGeckoID string,
) ([]domain.Rate, error) {
	m.requestedCoinGeckoID = coinGeckoID

	if m.getByCoinGeckoIDErr != nil {
		return nil, m.getByCoinGeckoIDErr
	}

	return m.getByCoinGeckoIDResult, nil
}

type mockSubscriptionService struct {
	saved              *domain.Subscription
	deletedChat        int64
	deletedCoinGeckoID string
	deleteCalled       bool
	deleteAllCalled    bool
	saveErr            error
	deleteErr          error

	subscriptions   []domain.Subscription
	getByChatIDErr  error
	requestedChatID int64
}

func (m *mockSubscriptionService) GetByChatID(
	ctx context.Context,
	chatID int64,
) ([]domain.Subscription, error) {
	m.requestedChatID = chatID

	if m.getByChatIDErr != nil {
		return nil, m.getByChatIDErr
	}

	return m.subscriptions, nil
}

func (m *mockSubscriptionService) Save(
	ctx context.Context,
	subscription *domain.Subscription,
) error {
	m.saved = subscription

	return m.saveErr
}

func (m *mockSubscriptionService) Delete(
	ctx context.Context,
	chatID int64,
	coinGeckoID string,
) error {
	m.deleteCalled = true
	m.deletedChat = chatID
	m.deletedCoinGeckoID = coinGeckoID

	return m.deleteErr
}

func (m *mockSubscriptionService) DeleteAll(
	ctx context.Context,
	chatID int64,
) error {
	m.deleteAllCalled = true
	m.deletedChat = chatID

	return m.deleteErr
}

type mockCoinResolver struct {
	coin      *domain.Coin
	err       error
	requested string
}

func (m *mockCoinResolver) ResolveCoin(
	ctx context.Context,
	query string,
) (*domain.Coin, error) {
	m.requested = query

	return m.coin, m.err
}

func makeUpdate(
	chatID int64,
	text string,
) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
		},
	}
}

func newTestHandler(
	bot *mockBot,
	rates *mockRateService,
	subscriptions *mockSubscriptionService,
	resolver *mockCoinResolver,
) *Handler {
	return NewHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)
}

func TestHandler_Start(t *testing.T) {
	bot := &mockBot{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start"),
	)

	require.Equal(t, int64(123), bot.chatID)
	require.Contains(
		t,
		bot.text,
		"Выбери нужное действие",
	)

	require.NotNil(t, bot.keyboard)
	require.True(t, bot.keyboard.ResizeKeyboard)

	require.Len(t, bot.keyboard.Keyboard, 2)

	require.Equal(
		t,
		"💰 Курсы",
		bot.keyboard.Keyboard[0][0].Text,
	)

	require.Equal(
		t,
		"🔔 Подписаться",
		bot.keyboard.Keyboard[0][1].Text,
	)

	require.Equal(
		t,
		"📋 Подписки",
		bot.keyboard.Keyboard[1][0].Text,
	)

	require.Equal(
		t,
		"ℹ️ Помощь",
		bot.keyboard.Keyboard[1][1].Text,
	)
}

func TestHandler_Rates_Success(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		rates: []domain.Rate{
			{
				Symbol:   "BTC",
				Price:    100000,
				DayLow:   98000,
				DayHigh:  101000,
				Change1H: 2.5,
			},
			{
				Symbol:   "ETH",
				Price:    2000,
				DayLow:   1900,
				DayHigh:  2100,
				Change1H: -1.5,
			},
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates"),
	)

	require.Equal(t, int64(123), bot.chatID)
	require.Contains(t, bot.text, "Текущие курсы:")
	require.Contains(t, bot.text, "BTC")
	require.Contains(t, bot.text, "ETH")
	require.Contains(t, bot.text, "$100000.00")
	require.Contains(t, bot.text, "$2000.00")
}

func TestHandler_Rates_Empty(t *testing.T) {
	bot := &mockBot{}

	handler := newTestHandler(
		bot,
		&mockRateService{
			rates: []domain.Rate{},
		},
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates"),
	)

	require.Equal(
		t,
		"Данные о курсах пока отсутствуют.",
		bot.text,
	)
}

func TestHandler_Rates_Error(t *testing.T) {
	bot := &mockBot{}

	handler := newTestHandler(
		bot,
		&mockRateService{
			err: errors.New("database unavailable"),
		},
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates"),
	)

	require.Equal(
		t,
		"Не удалось получить курсы.",
		bot.text,
	)
}

func TestHandler_RateBySymbol_Success(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		currentRate: &domain.Rate{
			CoinGeckoID: "bitcoin",
			Symbol:      "BTC",
			Price:       85336,
			DayLow:      84000,
			DayHigh:     86000,
			Change1H:    0.25,
		},
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "bitcoin",
			Symbol: "BTC",
			Name:   "Bitcoin",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates btc"),
	)

	require.Equal(t, "btc", resolver.requested)
	require.Equal(t, "bitcoin", rates.currentRateCoinID)

	require.Contains(t, bot.text, "BTC")
	require.Contains(
		t,
		bot.text,
		"$85336.00",
	)
}

func TestHandler_RateBySymbol_ArbitraryCoinSuccess(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		currentRate: &domain.Rate{
			CoinGeckoID: "solana",
			Symbol:      "SOL",
			Price:       116.59,
			DayLow:      114,
			DayHigh:     120,
			Change1H:    0.5,
		},
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates SOL"),
	)

	require.Equal(t, "SOL", resolver.requested)
	require.Equal(t, "solana", rates.currentRateCoinID)

	require.Contains(t, bot.text, "SOL")
	require.Contains(
		t,
		bot.text,
		"$116.59",
	)
}

func TestHandler_RateBySymbol_CoinNotFound(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates UNKNOWN"),
	)

	require.Equal(t, "UNKNOWN", resolver.requested)
	require.Empty(t, rates.requestedCoinGeckoID)

	require.Equal(
		t,
		"Не удалось найти такую криптовалюту.",
		bot.text,
	)
}

func TestHandler_RateBySymbol_Error(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		currentRateErr: errors.New("coingecko unavailable"),
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "bitcoin",
			Symbol: "BTC",
			Name:   "Bitcoin",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates BTC"),
	)

	require.Equal(t, "BTC", resolver.requested)
	require.Equal(
		t,
		"bitcoin",
		rates.currentRateCoinID,
	)

	require.Equal(
		t,
		"Не удалось получить курс.",
		bot.text,
	)
}

func TestHandler_RateBySymbol_Empty(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		currentRate: nil,
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "dogecoin",
			Symbol: "DOGE",
			Name:   "Dogecoin",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		&mockSubscriptionService{},
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates DOGE"),
	)

	require.Equal(
		t,
		"dogecoin",
		rates.currentRateCoinID,
	)

	require.Equal(
		t,
		"CoinGecko не вернул курс DOGE.",
		bot.text,
	)
}

func TestHandler_StartAuto_Success(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(777, "/start_auto SOL 5"),
	)

	require.Equal(t, "SOL", resolver.requested)

	require.NotNil(t, subscriptions.saved)

	require.Equal(
		t,
		int64(777),
		subscriptions.saved.ChatID,
	)

	require.Equal(
		t,
		"solana",
		subscriptions.saved.CoinGeckoID,
	)

	require.Equal(
		t,
		"SOL",
		subscriptions.saved.Symbol,
	)

	require.True(
		t,
		subscriptions.saved.Enabled,
	)

	require.Equal(
		t,
		5,
		subscriptions.saved.IntervalMinutes,
	)

	require.Equal(
		t,
		"Автоматическая отправка SOL включена каждые 5 мин.",
		bot.text,
	)
}

func TestHandler_StartAuto_HyphenAlias(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "dogecoin",
			Symbol: "DOGE",
			Name:   "Dogecoin",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(777, "/start-auto DOGE 10"),
	)

	require.Equal(t, "DOGE", resolver.requested)
	require.NotNil(t, subscriptions.saved)
	require.Equal(t, "dogecoin", subscriptions.saved.CoinGeckoID)
	require.Equal(t, "DOGE", subscriptions.saved.Symbol)
	require.Equal(t, 10, subscriptions.saved.IntervalMinutes)
}

func TestHandler_StartAuto_InvalidArguments(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto SOL"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Использование: /start_auto SOL 10",
		bot.text,
	)
}

func TestHandler_StartAuto_InvalidNumber(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto SOL abc"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Интервал должен быть числом.",
		bot.text,
	)
}

func TestHandler_StartAuto_TooSmall(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto SOL 0"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Интервал должен быть не меньше 1 минуты.",
		bot.text,
	)
}

func TestHandler_StartAuto_CoinNotFound(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto UNKNOWN 5"),
	)

	require.Equal(t, "UNKNOWN", resolver.requested)
	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Криптовалюта не найдена.",
		bot.text,
	)
}

func TestHandler_StartAuto_CoinAmbiguous(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinAmbiguous,
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto ABC 5"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Тикер неоднозначен. Укажи CoinGecko ID монеты.",
		bot.text,
	)
}

func TestHandler_StartAuto_ResolverError(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: errors.New("coingecko unavailable"),
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto SOL 5"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Не удалось найти криптовалюту.",
		bot.text,
	)
}

func TestHandler_StartAuto_SaveError(t *testing.T) {
	bot := &mockBot{}

	subscriptions := &mockSubscriptionService{
		saveErr: errors.New("database unavailable"),
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto SOL 5"),
	)

	require.NotNil(t, subscriptions.saved)

	require.Equal(
		t,
		"Не удалось включить автоматическую отправку.",
		bot.text,
	)
}

func TestHandler_StopAuto_AllSuccess(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto"),
	)

	require.Equal(
		t,
		int64(555),
		subscriptions.deletedChat,
	)

	require.True(t, subscriptions.deleteAllCalled)
	require.False(t, subscriptions.deleteCalled)

	require.Equal(
		t,
		"Все автоматические отправки отключены.",
		bot.text,
	)
}

func TestHandler_StopAuto_AllHyphenAlias(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop-auto"),
	)

	require.True(t, subscriptions.deleteAllCalled)
	require.False(t, subscriptions.deleteCalled)
}

func TestHandler_StopAuto_OneSuccess(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto SOL"),
	)

	require.Equal(t, "SOL", resolver.requested)

	require.True(t, subscriptions.deleteCalled)
	require.False(t, subscriptions.deleteAllCalled)

	require.Equal(
		t,
		int64(555),
		subscriptions.deletedChat,
	)

	require.Equal(
		t,
		"solana",
		subscriptions.deletedCoinGeckoID,
	)

	require.Equal(
		t,
		"Автоматическая отправка SOL отключена.",
		bot.text,
	)
}

func TestHandler_StopAuto_OneHyphenAlias(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "dogecoin",
			Symbol: "DOGE",
			Name:   "Dogecoin",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop-auto DOGE"),
	)

	require.True(t, subscriptions.deleteCalled)
	require.Equal(
		t,
		"dogecoin",
		subscriptions.deletedCoinGeckoID,
	)
}

func TestHandler_StopAuto_CoinNotFound(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto UNKNOWN"),
	)

	require.False(t, subscriptions.deleteCalled)

	require.Equal(
		t,
		"Криптовалюта не найдена.",
		bot.text,
	)
}

func TestHandler_StopAuto_CoinAmbiguous(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinAmbiguous,
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto ABC"),
	)

	require.False(t, subscriptions.deleteCalled)

	require.Equal(
		t,
		"Тикер неоднозначен. Укажи CoinGecko ID монеты.",
		bot.text,
	)
}

func TestHandler_StopAuto_ResolverError(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: errors.New("coingecko unavailable"),
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto SOL"),
	)

	require.False(t, subscriptions.deleteCalled)

	require.Equal(
		t,
		"Не удалось найти криптовалюту.",
		bot.text,
	)
}

func TestHandler_StopAuto_DeleteOneError(t *testing.T) {
	bot := &mockBot{}

	subscriptions := &mockSubscriptionService{
		deleteErr: errors.New("database unavailable"),
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto SOL"),
	)

	require.True(t, subscriptions.deleteCalled)

	require.Equal(
		t,
		"Не удалось отключить автоматическую отправку.",
		bot.text,
	)
}

func TestHandler_StopAuto_DeleteAllError(t *testing.T) {
	bot := &mockBot{}

	subscriptions := &mockSubscriptionService{
		deleteErr: errors.New("database unavailable"),
	}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto"),
	)

	require.True(t, subscriptions.deleteAllCalled)

	require.Equal(
		t,
		"Не удалось отключить автоматическую отправку.",
		bot.text,
	)
}

func TestHandler_StopAuto_InvalidArguments(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		subscriptions,
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop_auto SOL extra"),
	)

	require.False(t, subscriptions.deleteCalled)
	require.False(t, subscriptions.deleteAllCalled)

	require.Equal(
		t,
		"Использование: /stop_auto или /stop_auto SOL",
		bot.text,
	)
}

func TestHandler_UnknownCommand(t *testing.T) {
	bot := &mockBot{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/hello"),
	)

	require.Equal(
		t,
		"Неизвестная команда. Используй /start.",
		bot.text,
	)
}

func TestHandler_NoMessage(t *testing.T) {
	bot := &mockBot{}

	handler := newTestHandler(
		bot,
		&mockRateService{},
		&mockSubscriptionService{},
		&mockCoinResolver{},
	)

	handler.Handle(
		context.Background(),
		tgbotapi.Update{},
	)

	require.Empty(t, bot.text)
	require.Equal(t, int64(0), bot.chatID)
}

func TestFormatRate(t *testing.T) {
	rate := domain.Rate{
		Symbol:   "BTC",
		Price:    100000,
		DayLow:   98000,
		DayHigh:  101000,
		Change1H: 2.5,
	}

	actual := formatRate(rate)

	require.True(t, strings.Contains(actual, "BTC"))
	require.True(
		t,
		strings.Contains(actual, "Цена: $100000.00"),
	)
	require.True(
		t,
		strings.Contains(actual, "Минимум за 24ч: $98000.00"),
	)
	require.True(
		t,
		strings.Contains(actual, "Максимум за 24ч: $101000.00"),
	)
	require.True(
		t,
		strings.Contains(actual, "Изменение за 1ч: 2.50%"),
	)
}

func TestHandler_CustomRateButton_SetsWaitingState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🔎 Другая монета"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingRateCoin,
		session.State,
	)

	require.Contains(
		t,
		bot.text,
		"Введи тикер или CoinGecko ID",
	)
}

func TestHandler_CustomRate_Success(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		currentRate: &domain.Rate{
			CoinGeckoID: "avalanche-2",
			Symbol:      "AVAX",
			Price:       11.18,
			DayLow:      10.74,
			DayHigh:     11.20,
			Change1H:    0.42,
		},
	}

	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "avalanche-2",
			Symbol: "AVAX",
			Name:   "Avalanche",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingRateCoin,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "AVAX"),
	)

	require.Equal(
		t,
		"AVAX",
		resolver.requested,
	)

	require.Equal(
		t,
		"avalanche-2",
		rates.currentRateCoinID,
	)

	require.Contains(
		t,
		bot.text,
		"AVAX",
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateNone,
		session.State,
	)
}

func TestHandler_CustomRate_NotFound_KeepsWaitingState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingRateCoin,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "UNKNOWN"),
	)

	require.Equal(
		t,
		"UNKNOWN",
		resolver.requested,
	)

	require.Contains(
		t,
		bot.text,
		"не найд",
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingRateCoin,
		session.State,
	)
}

func TestHandler_SubscribeButton_SetsWaitingCoinState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🔔 Подписаться"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionCoin,
		session.State,
	)

	require.Contains(
		t,
		bot.text,
		"криптовалют",
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_SubscriptionCoin_SetsWaitingIntervalState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "solana",
			Symbol: "SOL",
			Name:   "Solana",
		},
	}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingSubscriptionCoin,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "SOL"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionInterval,
		session.State,
	)

	require.Equal(
		t,
		"solana",
		session.SubscriptionCoinID,
	)

	require.Equal(
		t,
		"SOL",
		session.SubscriptionCoinSymbol,
	)

	require.Equal(
		t,
		"SOL",
		resolver.requested,
	)

	require.Contains(
		t,
		bot.text,
		"SOL",
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_SubscriptionCoin_NotFound_KeepsState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingSubscriptionCoin,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "UNKNOWN"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionCoin,
		session.State,
	)

	require.Nil(
		t,
		subscriptions.saved,
	)
}

func TestHandler_SubscriptionInterval_CreatesSubscription(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State:                  stateWaitingSubscriptionInterval,
			SubscriptionCoinID:     "solana",
			SubscriptionCoinSymbol: "SOL",
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "5 мин"),
	)

	require.NotNil(
		t,
		subscriptions.saved,
	)

	require.Equal(
		t,
		int64(123),
		subscriptions.saved.ChatID,
	)

	require.Equal(
		t,
		"solana",
		subscriptions.saved.CoinGeckoID,
	)

	require.Equal(
		t,
		"SOL",
		subscriptions.saved.Symbol,
	)

	require.Equal(
		t,
		5,
		subscriptions.saved.IntervalMinutes,
	)

	require.True(
		t,
		subscriptions.saved.Enabled,
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateNone,
		session.State,
	)

	require.Contains(
		t,
		bot.text,
		"Подписка создана",
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_SubscriptionInterval_Invalid_KeepsState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State:                  stateWaitingSubscriptionInterval,
			SubscriptionCoinID:     "solana",
			SubscriptionCoinSymbol: "SOL",
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "abc"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionInterval,
		session.State,
	)

	require.Nil(
		t,
		subscriptions.saved,
	)
}

func TestHandler_SubscriptionInterval_SaveError_KeepsState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}

	subscriptions := &mockSubscriptionService{
		saveErr: errors.New("database unavailable"),
	}

	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State:                  stateWaitingSubscriptionInterval,
			SubscriptionCoinID:     "solana",
			SubscriptionCoinSymbol: "SOL",
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "5 мин"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionInterval,
		session.State,
	)

	require.NotNil(
		t,
		subscriptions.saved,
	)

	require.Contains(
		t,
		bot.text,
		"Не удалось создать подписку",
	)
}

func TestHandler_SubscriptionCancel_ClearsState(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State:                  stateWaitingSubscriptionInterval,
			SubscriptionCoinID:     "solana",
			SubscriptionCoinSymbol: "SOL",
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "❌ Отмена"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateNone,
		session.State,
	)

	require.Nil(
		t,
		subscriptions.saved,
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_Subscriptions_Success(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Symbol:          "BTC",
				CoinGeckoID:     "bitcoin",
				Enabled:         true,
				IntervalMinutes: 5,
			},
			{
				ChatID:          123,
				Symbol:          "SOL",
				CoinGeckoID:     "solana",
				Enabled:         true,
				IntervalMinutes: 30,
			},
		},
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "📋 Подписки"),
	)

	require.Equal(t, int64(123), subscriptions.requestedChatID)

	require.Contains(
		t,
		bot.text,
		"BTC",
	)

	require.Contains(
		t,
		bot.text,
		"5 мин",
	)

	require.Contains(
		t,
		bot.text,
		"SOL",
	)

	require.Contains(
		t,
		bot.text,
		"30 мин",
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_Subscriptions_Empty(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "📋 Подписки"),
	)

	require.Contains(
		t,
		bot.text,
		"нет активных подписок",
	)

	require.NotNil(t, bot.keyboard)
}

func TestHandler_Subscriptions_Error(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		getByChatIDErr: errors.New("database unavailable"),
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "📋 Подписки"),
	)

	require.Contains(
		t,
		bot.text,
		"Не удалось получить список подписок",
	)
}

func TestHandler_SubscriptionDeleteMenu_Success(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Symbol:          "BTC",
				CoinGeckoID:     "bitcoin",
				Enabled:         true,
				IntervalMinutes: 5,
			},
			{
				ChatID:          123,
				Symbol:          "SOL",
				CoinGeckoID:     "solana",
				Enabled:         true,
				IntervalMinutes: 10,
			},
		},
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🗑 Удалить подписку"),
	)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionDelete,
		session.State,
	)

	require.NotNil(t, bot.keyboard)
	require.Contains(t, bot.text, "Какую подписку удалить")
}

func TestHandler_SubscriptionDeleteMenu_Empty(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🗑 Удалить подписку"),
	)

	session := handler.sessions.get(123)

	require.Equal(t, stateNone, session.State)
	require.Contains(t, bot.text, "нет активных подписок")
}

func TestHandler_SubscriptionDeleteMenu_Error(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		getByChatIDErr: errors.New("database unavailable"),
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🗑 Удалить подписку"),
	)

	require.Contains(
		t,
		bot.text,
		"Не удалось получить список подписок",
	)
}

func TestHandler_SubscriptionDelete_Success(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		subscriptions: []domain.Subscription{
			{
				ChatID:          123,
				Symbol:          "SOL",
				CoinGeckoID:     "solana",
				Enabled:         true,
				IntervalMinutes: 5,
			},
		},
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingSubscriptionDelete,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "SOL"),
	)

	require.True(t, subscriptions.deleteCalled)
	require.Equal(t, int64(123), subscriptions.deletedChat)
	require.Equal(t, "solana", subscriptions.deletedCoinGeckoID)

	session := handler.sessions.get(123)

	require.Equal(t, stateNone, session.State)
	require.Contains(t, bot.text, "Подписка SOL удалена")
}

func TestHandler_SubscriptionDelete_NotFound(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		subscriptions: []domain.Subscription{
			{
				ChatID:      123,
				Symbol:      "BTC",
				CoinGeckoID: "bitcoin",
				Enabled:     true,
			},
		},
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingSubscriptionDelete,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "SOL"),
	)

	require.False(t, subscriptions.deleteCalled)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionDelete,
		session.State,
	)

	require.Contains(
		t,
		bot.text,
		"Такой активной подписки нет",
	)
}

func TestHandler_SubscriptionDelete_Error(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		subscriptions: []domain.Subscription{
			{
				ChatID:      123,
				Symbol:      "SOL",
				CoinGeckoID: "solana",
				Enabled:     true,
			},
		},
		deleteErr: errors.New("database unavailable"),
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.sessions.set(
		123,
		userSession{
			State: stateWaitingSubscriptionDelete,
		},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "SOL"),
	)

	require.True(t, subscriptions.deleteCalled)

	session := handler.sessions.get(123)

	require.Equal(
		t,
		stateWaitingSubscriptionDelete,
		session.State,
	)

	require.Contains(
		t,
		bot.text,
		"Не удалось удалить подписку",
	)
}

func TestHandler_DeleteAllSubscriptions_Success(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🗑 Удалить все"),
	)

	require.True(t, subscriptions.deleteAllCalled)
	require.Equal(t, int64(123), subscriptions.deletedChat)
	require.Contains(t, bot.text, "Все подписки удалены")
}

func TestHandler_DeleteAllSubscriptions_Error(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{
		deleteErr: errors.New("database unavailable"),
	}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "🗑 Удалить все"),
	)

	require.True(t, subscriptions.deleteAllCalled)
	require.Contains(
		t,
		bot.text,
		"Не удалось удалить подписки",
	)
}

func TestHandler_Help(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}
	resolver := &mockCoinResolver{}

	handler := newTestHandler(
		bot,
		rates,
		subscriptions,
		resolver,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "ℹ️ Помощь"),
	)

	require.Contains(
		t,
		bot.text,
		"Помощь",
	)

	require.Contains(
		t,
		bot.text,
		"/rates",
	)

	require.Contains(
		t,
		bot.text,
		"/start_auto",
	)

	require.Contains(
		t,
		bot.text,
		"/stop_auto",
	)
}
