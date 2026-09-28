package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/domain"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/response"
)

type mockRateService struct {
	rates                []domain.Rate
	err                  error
	currentRate          *domain.Rate
	currentRateErr       error
	requestedCoinGeckoID string
}

func (m *mockRateService) GetLatest(
	ctx context.Context,
) ([]domain.Rate, error) {
	return m.rates, m.err
}

func (m *mockRateService) GetCurrentRate(
	ctx context.Context,
	coinGeckoID string,
) (*domain.Rate, error) {
	m.requestedCoinGeckoID = coinGeckoID

	return m.currentRate, m.currentRateErr
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

func TestRateHandler_GetLatest_Success(t *testing.T) {

	service := &mockRateService{
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

	handler := NewRateHandler(
		service,
		&mockCoinResolver{},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetLatest(
		recorder,
		request,
	)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var actual []response.Rate

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Len(
		t,
		actual,
		1,
	)

	require.Equal(
		t,
		"BTC",
		actual[0].Symbol,
	)

	require.Equal(
		t,
		100000.0,
		actual[0].Price,
	)
}

func TestRateHandler_GetLatest_Error(t *testing.T) {

	service := &mockRateService{
		err: errors.New("database unavailable"),
	}

	handler := NewRateHandler(
		service,
		&mockCoinResolver{},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetLatest(recorder, request)

	require.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	var actual response.Error

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		"database unavailable",
		actual.Error,
	)
}

func TestRateHandler_GetLatest_Empty(t *testing.T) {

	service := &mockRateService{
		rates: []domain.Rate{},
	}

	handler := NewRateHandler(
		service,
		&mockCoinResolver{},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.GetLatest(recorder, request)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var actual []response.Rate

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Empty(
		t,
		actual,
	)
}

func TestRateHandler_GetBySymbol_Success(t *testing.T) {
	service := &mockRateService{
		currentRate: &domain.Rate{
			CoinGeckoID: "dogecoin",
			Symbol:      "DOGE",
			Price:       0.1,
			DayLow:      0.09,
			DayHigh:     0.11,
			Change1H:    0.5,
		},
	}

	resolver := &mockCoinResolver{
		coin: &domain.Coin{
			ID:     "dogecoin",
			Symbol: "DOGE",
			Name:   "Dogecoin",
		},
	}

	handler := NewRateHandler(
		service,
		resolver,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/DOGE",
		nil,
	)

	request.SetPathValue(
		"symbol",
		"DOGE",
	)

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(
		recorder,
		request,
	)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	require.Equal(
		t,
		"DOGE",
		resolver.requested,
	)

	require.Equal(
		t,
		"dogecoin",
		service.requestedCoinGeckoID,
	)

	var actual []response.Rate

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)
	require.Len(t, actual, 1)

	require.Equal(
		t,
		"DOGE",
		actual[0].Symbol,
	)

	require.Equal(
		t,
		0.1,
		actual[0].Price,
	)
}

func TestRateHandler_GetBySymbol_NotFound(t *testing.T) {
	service := &mockRateService{}

	resolver := &mockCoinResolver{
		err: coingecko.ErrCoinNotFound,
	}

	handler := NewRateHandler(
		service,
		resolver,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/UNKNOWN",
		nil,
	)

	request.SetPathValue(
		"symbol",
		"UNKNOWN",
	)

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(
		recorder,
		request,
	)

	require.Equal(
		t,
		http.StatusNotFound,
		recorder.Code,
	)

	require.Empty(
		t,
		service.requestedCoinGeckoID,
	)
}
