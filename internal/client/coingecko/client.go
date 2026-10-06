package coingecko

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

const baseURL = "https://api.coingecko.com/api/v3"

type cachedCoin struct {
	coin      domain.Coin
	expiresAt time.Time
}

type Client struct {
	httpClient *http.Client
	baseURL    string

	cacheMu   sync.RWMutex
	coinCache map[string]cachedCoin
	cacheTTL  time.Duration

	cooldownMu    sync.RWMutex
	cooldownUntil time.Time
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:   baseURL,
		coinCache: make(map[string]cachedCoin),
		cacheTTL:  24 * time.Hour,
	}
}

func (c *Client) cooldownRemaining() time.Duration {
	c.cooldownMu.RLock()
	until := c.cooldownUntil
	c.cooldownMu.RUnlock()

	remaining := time.Until(until)
	if remaining <= 0 {
		return 0
	}

	return remaining
}

func (c *Client) setCooldown(duration time.Duration) {
	if duration <= 0 {
		return
	}

	until := time.Now().Add(duration)

	c.cooldownMu.Lock()
	defer c.cooldownMu.Unlock()

	if until.After(c.cooldownUntil) {
		c.cooldownUntil = until
	}
}

func (c *Client) handleAPIError(resp *http.Response) error {
	err := newAPIError(resp)

	apiErr, ok := err.(*APIError)
	if !ok {
		return err
	}

	if apiErr.StatusCode == http.StatusTooManyRequests {
		retryAfter := apiErr.RetryAfter

		if retryAfter <= 0 {
			retryAfter = time.Minute
		}

		c.setCooldown(retryAfter)
	}

	return apiErr
}

func (c *Client) GetRates(
	ctx context.Context,
	coinGeckoIDs []string,
) ([]domain.Rate, error) {

	if len(coinGeckoIDs) == 0 {
		return []domain.Rate{}, nil
	}

	if remaining := c.cooldownRemaining(); remaining > 0 {
		return nil, &APIError{
			StatusCode: http.StatusTooManyRequests,
			Body:       "CoinGecko request skipped during cooldown",
			RetryAfter: remaining,
		}
	}

	endpoint, err := url.Parse(
		c.baseURL + "/coins/markets",
	)
	if err != nil {
		return nil, err
	}

	query := endpoint.Query()
	query.Set("vs_currency", "usd")
	query.Set(
		"ids",
		strings.Join(coinGeckoIDs, ","),
	)
	query.Set(
		"price_change_percentage",
		"1h",
	)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	var result MarketResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return ToDomain(result), nil
}

func (c *Client) getCachedCoin(
	query string,
) (*domain.Coin, bool) {

	key := normalizeCoinQuery(query)

	c.cacheMu.RLock()
	entry, ok := c.coinCache[key]
	c.cacheMu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.cacheMu.Lock()

		// Проверяем ещё раз после получения write-lock.
		entry, ok = c.coinCache[key]
		if ok && time.Now().After(entry.expiresAt) {
			delete(c.coinCache, key)
		}

		c.cacheMu.Unlock()

		return nil, false
	}

	coin := entry.coin

	return &coin, true
}

func (c *Client) cacheCoin(
	query string,
	coin *domain.Coin,
) {
	if coin == nil {
		return
	}

	entry := cachedCoin{
		coin:      *coin,
		expiresAt: time.Now().Add(c.cacheTTL),
	}

	keys := []string{
		normalizeCoinQuery(query),
		normalizeCoinQuery(coin.ID),
		normalizeCoinQuery(coin.Symbol),
	}

	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	for _, key := range keys {
		if key == "" {
			continue
		}

		c.coinCache[key] = entry
	}
}

func normalizeCoinQuery(
	query string,
) string {

	return strings.ToLower(
		strings.TrimSpace(query),
	)
}
