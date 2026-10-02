package coingecko

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestResolveSearchResult_BySymbol(t *testing.T) {
	coins := []SearchCoin{
		{
			ID:     "solana",
			Name:   "Solana",
			Symbol: "SOL",
		},
	}

	coin, err := resolveSearchResult(
		"SOL",
		coins,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coin.ID != "solana" {
		t.Fatalf(
			"expected solana, got %s",
			coin.ID,
		)
	}

	if coin.Symbol != "SOL" {
		t.Fatalf(
			"expected SOL, got %s",
			coin.Symbol,
		)
	}
}

func TestResolveSearchResult_ByID(t *testing.T) {
	coins := []SearchCoin{
		{
			ID:     "the-open-network",
			Name:   "Toncoin",
			Symbol: "TON",
		},
	}

	coin, err := resolveSearchResult(
		"the-open-network",
		coins,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coin.ID != "the-open-network" {
		t.Fatalf(
			"expected the-open-network, got %s",
			coin.ID,
		)
	}

	if coin.Symbol != "TON" {
		t.Fatalf(
			"expected TON, got %s",
			coin.Symbol,
		)
	}
}

func TestResolveSearchResult_NotFound(t *testing.T) {
	coins := []SearchCoin{
		{
			ID:     "bitcoin",
			Name:   "Bitcoin",
			Symbol: "BTC",
		},
	}

	_, err := resolveSearchResult(
		"UNKNOWN",
		coins,
	)

	if !errors.Is(err, ErrCoinNotFound) {
		t.Fatalf(
			"expected ErrCoinNotFound, got %v",
			err,
		)
	}
}

func TestResolveSearchResult_AmbiguousSymbol(
	t *testing.T,
) {
	coins := []SearchCoin{
		{
			ID:     "coin-one",
			Name:   "Coin One",
			Symbol: "ABC",
		},
		{
			ID:     "coin-two",
			Name:   "Coin Two",
			Symbol: "ABC",
		},
	}

	_, err := resolveSearchResult(
		"ABC",
		coins,
	)

	if !errors.Is(err, ErrCoinAmbiguous) {
		t.Fatalf(
			"expected ErrCoinAmbiguous, got %v",
			err,
		)
	}
}

func TestResolveSearchResult_IDWinsOverAmbiguousSymbol(
	t *testing.T,
) {
	coins := []SearchCoin{
		{
			ID:     "abc",
			Name:   "ABC Coin",
			Symbol: "ABC",
		},
		{
			ID:     "another-coin",
			Name:   "Another Coin",
			Symbol: "ABC",
		},
	}

	coin, err := resolveSearchResult(
		"abc",
		coins,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coin.ID != "abc" {
		t.Fatalf(
			"expected abc, got %s",
			coin.ID,
		)
	}
}

func TestResolveCoin_UsesCache(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Path != "/search" {
				t.Fatalf(
					"expected /search, got %s",
					r.URL.Path,
				)
			}

			atomic.AddInt32(&requestCount, 1)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			_, err := w.Write([]byte(`{
				"coins": [
					{
						"id": "avalanche-2",
						"name": "Avalanche",
						"symbol": "AVAX",
						"market_cap_rank": 25
					}
				]
			}`))
			if err != nil {
				t.Fatalf(
					"write response: %v",
					err,
				)
			}
		}),
	)
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		coinCache:  make(map[string]cachedCoin),
		cacheTTL:   time.Hour,
	}

	first, err := client.ResolveCoin(
		t.Context(),
		"AVAX",
	)
	if err != nil {
		t.Fatalf(
			"first ResolveCoin: %v",
			err,
		)
	}

	second, err := client.ResolveCoin(
		t.Context(),
		"AVAX",
	)
	if err != nil {
		t.Fatalf(
			"second ResolveCoin: %v",
			err,
		)
	}

	if first.ID != "avalanche-2" {
		t.Fatalf(
			"expected avalanche-2, got %s",
			first.ID,
		)
	}

	if second.ID != "avalanche-2" {
		t.Fatalf(
			"expected avalanche-2 from cache, got %s",
			second.ID,
		)
	}

	if got := atomic.LoadInt32(&requestCount); got != 1 {
		t.Fatalf(
			"expected 1 HTTP request, got %d",
			got,
		)
	}
}
