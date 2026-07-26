package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
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


func (c *Client) GetPrices(ctx context.Context) (PriceResponse, error) {

	url := baseURL +
		"/simple/price" +
		"?ids=bitcoin,ethereum" +
		"&vs_currencies=usd" +
		"&include_24hr_change=true" +
		"&include_24hr_high=true" +
		"&include_24hr_low=true"

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
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result PriceResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}