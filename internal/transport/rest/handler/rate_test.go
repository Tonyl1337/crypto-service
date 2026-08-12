package handler

import (
	"context"
	"errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Tonyl1337/crypto-service/internal/transport/rest/response"
	"github.com/Tonyl1337/crypto-service/internal/domain"
)

type mockRateService struct {
	rates []domain.Rate
	err   error
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

	return nil, nil
}


func TestRateHandler_GetLatest_Success(t *testing.T) {

service := &mockRateService{
	rates: []domain.Rate{
		{
			Symbol:    "BTC",
			Price:     100000,
			DayLow:    98000,
			DayHigh:   101000,
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