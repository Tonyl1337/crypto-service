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
	chatID int64
	text   string
	err    error
}

func (m *mockBot) SendMessage(
	chatID int64,
	text string,
) error {
	m.chatID = chatID
	m.text = text

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
	require.Contains(t, bot.text, "Доступные команды:")
	require.Contains(t, bot.text, "/rates")
	require.Contains(t, bot.text, "/start_auto SOL 10")
	require.Contains(t, bot.text, "/stop_auto SOL")
	require.Contains(t, bot.text, "/stop_auto")
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
