package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

const baseURL = "https://api.coingecko.com/api/v3"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (c *Client) GetRates(
	ctx context.Context,
	coinGeckoIDs []string,
) ([]domain.Rate, error) {

	if len(coinGeckoIDs) == 0 {
		return []domain.Rate{}, nil
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
		return nil, fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	var result MarketResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return ToDomain(result), nil
}
