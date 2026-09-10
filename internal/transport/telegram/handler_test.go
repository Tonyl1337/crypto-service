package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockBot struct {
	chatID int64
	text   string
	err    error
}

func (m *mockBot) SendMessage(chatID int64, text string) error {
	m.chatID = chatID
	m.text = text

	return m.err
}

type mockRateService struct {
	rates           []domain.Rate
	err             error
	requestedSymbol string
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

type mockSubscriptionService struct {
	saved       *domain.Subscription
	deletedChat int64
	saveErr     error
	deleteErr   error
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
) error {
	m.deletedChat = chatID

	return m.deleteErr
}

func makeUpdate(chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{
				ID: chatID,
			},
		},
	}
}

func TestHandler_Start(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		rates,
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start"),
	)

	require.Equal(t, int64(123), bot.chatID)
	require.Contains(
		t,
		bot.text,
		"Доступные команды:",
	)
	require.Contains(t, bot.text, "/rates")
	require.Contains(t, bot.text, "/start_auto")
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

	handler := NewHandler(
		bot,
		rates,
		&mockSubscriptionService{},
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

	handler := NewHandler(
		bot,
		&mockRateService{
			rates: []domain.Rate{},
		},
		&mockSubscriptionService{},
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

	handler := NewHandler(
		bot,
		&mockRateService{
			err: errors.New("database unavailable"),
		},
		&mockSubscriptionService{},
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
		rates: []domain.Rate{
			{
				Symbol:   "BTC",
				Price:    100000,
				DayLow:   98000,
				DayHigh:  101000,
				Change1H: 2.5,
			},
		},
	}

	handler := NewHandler(
		bot,
		rates,
		&mockSubscriptionService{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates btc"),
	)

	require.Equal(t, "BTC", rates.requestedSymbol)
	require.Contains(t, bot.text, "BTC")
	require.Contains(t, bot.text, "$100000.00")
}

func TestHandler_RateBySymbol_InvalidSymbol(t *testing.T) {
	bot := &mockBot{}
	rates := &mockRateService{}

	handler := NewHandler(
		bot,
		rates,
		&mockSubscriptionService{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates DOGE"),
	)

	require.Empty(t, rates.requestedSymbol)

	require.Equal(
		t,
		"Поддерживаются только BTC и ETH.",
		bot.text,
	)
}

func TestHandler_RateBySymbol_Error(t *testing.T) {
	bot := &mockBot{}

	rates := &mockRateService{
		err: errors.New("database unavailable"),
	}

	handler := NewHandler(
		bot,
		rates,
		&mockSubscriptionService{},
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/rates BTC"),
	)

	require.Equal(t, "BTC", rates.requestedSymbol)

	require.Equal(
		t,
		"Не удалось получить курс.",
		bot.text,
	)
}

func TestHandler_StartAuto_Success(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(777, "/start_auto 5"),
	)

	require.NotNil(t, subscriptions.saved)

	require.Equal(
		t,
		int64(777),
		subscriptions.saved.ChatID,
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
		"Автоматическая отправка включена каждые 5 мин.",
		bot.text,
	)
}

func TestHandler_StartAuto_HyphenAlias(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(777, "/start-auto 10"),
	)

	require.NotNil(t, subscriptions.saved)

	require.Equal(
		t,
		10,
		subscriptions.saved.IntervalMinutes,
	)
}

func TestHandler_StartAuto_InvalidNumber(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto abc"),
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

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto 0"),
	)

	require.Nil(t, subscriptions.saved)

	require.Equal(
		t,
		"Интервал должен быть не меньше 1 минуты.",
		bot.text,
	)
}

func TestHandler_StartAuto_SaveError(t *testing.T) {
	bot := &mockBot{}

	subscriptions := &mockSubscriptionService{
		saveErr: errors.New("database unavailable"),
	}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(123, "/start_auto 5"),
	)

	require.NotNil(t, subscriptions.saved)

	require.Equal(
		t,
		"Не удалось включить автоматическую отправку.",
		bot.text,
	)
}

func TestHandler_StopAuto_Success(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
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

	require.Equal(
		t,
		"Автоматическая отправка отключена.",
		bot.text,
	)
}

func TestHandler_StopAuto_HyphenAlias(t *testing.T) {
	bot := &mockBot{}
	subscriptions := &mockSubscriptionService{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
	)

	handler.Handle(
		context.Background(),
		makeUpdate(555, "/stop-auto"),
	)

	require.Equal(
		t,
		int64(555),
		subscriptions.deletedChat,
	)
}

func TestHandler_StopAuto_Error(t *testing.T) {
	bot := &mockBot{}

	subscriptions := &mockSubscriptionService{
		deleteErr: errors.New("database unavailable"),
	}

	handler := NewHandler(
		bot,
		&mockRateService{},
		subscriptions,
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

	require.Equal(
		t,
		"Не удалось отключить автоматическую отправку.",
		bot.text,
	)
}

func TestHandler_UnknownCommand(t *testing.T) {
	bot := &mockBot{}

	handler := NewHandler(
		bot,
		&mockRateService{},
		&mockSubscriptionService{},
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

	handler := NewHandler(
		bot,
		&mockRateService{},
		&mockSubscriptionService{},
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
	require.True(t, strings.Contains(actual, "Цена: $100000.00"))
	require.True(t, strings.Contains(actual, "Минимум за 24ч: $98000.00"))
	require.True(t, strings.Contains(actual, "Максимум за 24ч: $101000.00"))
	require.True(t, strings.Contains(actual, "Изменение за 1ч: 2.50%"))
}
