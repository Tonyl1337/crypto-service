package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/domain"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/response"
)

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

	handler := NewRateHandler(service)

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

	handler := NewRateHandler(service)

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

	handler := NewRateHandler(service)

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

	handler := NewRateHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/BTC",
		nil,
	)

	request.SetPathValue("symbol", "BTC")

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(recorder, request)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	require.Equal(
		t,
		"BTC",
		service.requestedSymbol,
	)

	var actual []response.Rate

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)
	require.Len(t, actual, 1)
	require.Equal(t, "BTC", actual[0].Symbol)
	require.Equal(t, 100000.0, actual[0].Price)
}

func TestRateHandler_GetBySymbol_Lowercase(t *testing.T) {
	service := &mockRateService{
		rates: []domain.Rate{
			{
				Symbol: "ETH",
				Price:  2000,
			},
		},
	}

	handler := NewRateHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/eth",
		nil,
	)

	request.SetPathValue("symbol", "eth")

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(recorder, request)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	require.Equal(
		t,
		"ETH",
		service.requestedSymbol,
	)
}

func TestRateHandler_GetBySymbol_InvalidSymbol(t *testing.T) {
	service := &mockRateService{}

	handler := NewRateHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/DOGE",
		nil,
	)

	request.SetPathValue("symbol", "DOGE")

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(recorder, request)

	require.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	require.Empty(
		t,
		service.requestedSymbol,
	)

	var actual response.Error

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&actual,
	)

	require.NoError(t, err)

	require.Equal(
		t,
		"invalid cryptocurrency symbol",
		actual.Error,
	)
}

func TestRateHandler_GetBySymbol_Error(t *testing.T) {
	service := &mockRateService{
		err: errors.New("database unavailable"),
	}

	handler := NewRateHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/rates/BTC",
		nil,
	)

	request.SetPathValue("symbol", "BTC")

	recorder := httptest.NewRecorder()

	handler.GetBySymbol(recorder, request)

	require.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	require.Equal(
		t,
		"BTC",
		service.requestedSymbol,
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
