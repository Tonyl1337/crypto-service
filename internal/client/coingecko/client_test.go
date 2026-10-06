package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"

	"github.com/stretchr/testify/require"
)

func TestClient_GetRates(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				require.Equal(
					t,
					"/coins/markets",
					r.URL.Path,
				)

				require.Equal(
					t,
					"usd",
					r.URL.Query().Get(
						"vs_currency",
					),
				)

				require.Equal(
					t,
					"bitcoin,ethereum",
					r.URL.Query().Get("ids"),
				)

				require.Equal(
					t,
					"1h",
					r.URL.Query().Get(
						"price_change_percentage",
					),
				)

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write([]byte(`
[
	{
		"id": "bitcoin",
		"symbol": "btc",
		"current_price": 83405,
		"high_24h": 84945,
		"low_24h": 82581,
		"price_change_percentage_1h_in_currency": 0.0592
	},
	{
		"id": "ethereum",
		"symbol": "eth",
		"current_price": 2684.02,
		"high_24h": 2698.21,
		"low_24h": 2636.99,
		"price_change_percentage_1h_in_currency": 0.0313
	}
]
`))
			},
		),
	)
	defer server.Close()

	client := NewClient()

	client.baseURL = server.URL

	rates, err := client.GetRates(
		context.Background(),
		[]string{
			"bitcoin",
			"ethereum",
		},
	)

	require.NoError(t, err)

	require.Len(t, rates, 2)

	require.Equal(
		t,
		"bitcoin",
		rates[0].CoinGeckoID,
	)

	require.Equal(
		t,
		"BTC",
		rates[0].Symbol,
	)

	require.Equal(
		t,
		83405.0,
		rates[0].Price,
	)

	require.Equal(
		t,
		82581.0,
		rates[0].DayLow,
	)

	require.Equal(
		t,
		84945.0,
		rates[0].DayHigh,
	)

	require.Equal(
		t,
		0.0592,
		rates[0].Change1H,
	)

	require.Equal(
		t,
		"ethereum",
		rates[1].CoinGeckoID,
	)

	require.Equal(
		t,
		"ETH",
		rates[1].Symbol,
	)

	require.False(
		t,
		rates[0].CreatedAt.IsZero(),
	)
}

func TestClient_GetRates_RateLimited(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Retry-After",
				"60",
			)

			w.WriteHeader(
				http.StatusTooManyRequests,
			)

			_, _ = w.Write(
				[]byte(`{"status":"rate limited"}`),
			)
		}),
	)
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.GetRates(
		context.Background(),
		[]string{"bitcoin"},
	)

	require.Error(t, err)

	var apiErr *APIError

	require.ErrorAs(
		t,
		err,
		&apiErr,
	)

	require.Equal(
		t,
		http.StatusTooManyRequests,
		apiErr.StatusCode,
	)

	require.Equal(
		t,
		`{"status":"rate limited"}`,
		apiErr.Body,
	)

	require.Equal(
		t,
		time.Minute,
		apiErr.RetryAfter,
	)
}

func TestClient_GetRates_CooldownPreventsSecondRequest(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			atomic.AddInt32(&requestCount, 1)

			w.Header().Set(
				"Retry-After",
				"60",
			)

			w.WriteHeader(
				http.StatusTooManyRequests,
			)
		}),
	)
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.GetRates(
		context.Background(),
		[]string{"bitcoin"},
	)
	require.Error(t, err)

	_, err = client.GetRates(
		context.Background(),
		[]string{"ethereum"},
	)
	require.Error(t, err)

	var apiErr *APIError

	require.ErrorAs(
		t,
		err,
		&apiErr,
	)

	require.Equal(
		t,
		http.StatusTooManyRequests,
		apiErr.StatusCode,
	)

	require.Equal(
		t,
		int32(1),
		atomic.LoadInt32(&requestCount),
	)
}

func TestClient_CooldownSharedWithResolver(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			atomic.AddInt32(&requestCount, 1)

			w.Header().Set(
				"Retry-After",
				"60",
			)

			w.WriteHeader(
				http.StatusTooManyRequests,
			)
		}),
	)
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.GetRates(
		context.Background(),
		[]string{"bitcoin"},
	)
	require.Error(t, err)

	_, err = client.ResolveCoin(
		context.Background(),
		"ADA",
	)
	require.Error(t, err)

	var apiErr *APIError

	require.ErrorAs(
		t,
		err,
		&apiErr,
	)

	require.Equal(
		t,
		http.StatusTooManyRequests,
		apiErr.StatusCode,
	)

	require.Equal(
		t,
		int32(1),
		atomic.LoadInt32(&requestCount),
	)
}

func TestClient_ResolveCoin_UsesCacheDuringCooldown(t *testing.T) {
	client := NewClient()

	client.cacheCoin(
		"ADA",
		&domain.Coin{
			ID:     "cardano",
			Symbol: "ADA",
			Name:   "Cardano",
		},
	)

	client.setCooldown(time.Minute)

	coin, err := client.ResolveCoin(
		context.Background(),
		"ADA",
	)

	require.NoError(t, err)
	require.NotNil(t, coin)

	require.Equal(
		t,
		"cardano",
		coin.ID,
	)

	require.Equal(
		t,
		"ADA",
		coin.Symbol,
	)
}

func TestClient_GetRates_ResumesAfterCooldown(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			atomic.AddInt32(&requestCount, 1)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, _ = w.Write([]byte(`[
				{
					"id": "bitcoin",
					"symbol": "btc",
					"current_price": 83405,
					"high_24h": 84945,
					"low_24h": 82581,
					"price_change_percentage_1h_in_currency": 0.0592
				}
			]`))
		}),
	)
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	client.setCooldown(
		20 * time.Millisecond,
	)

	_, err := client.GetRates(
		context.Background(),
		[]string{"bitcoin"},
	)
	require.Error(t, err)

	require.Equal(
		t,
		int32(0),
		atomic.LoadInt32(&requestCount),
	)

	time.Sleep(
		30 * time.Millisecond,
	)

	rates, err := client.GetRates(
		context.Background(),
		[]string{"bitcoin"},
	)

	require.NoError(t, err)
	require.Len(t, rates, 1)

	require.Equal(
		t,
		int32(1),
		atomic.LoadInt32(&requestCount),
	)
}
