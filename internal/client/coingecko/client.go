package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tonyl1337/crypto-service/internal/domain"
	"net/http"
	"time"
)

const baseURL = "https://api.coingecko.com/api/v3"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetRates(
	ctx context.Context,
) ([]domain.Rate, error) {

	url := baseURL +
		"/coins/markets" +
		"?vs_currency=usd" +
		"&ids=bitcoin,ethereum" +
		"&price_change_percentage=1h"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
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
