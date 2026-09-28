package coingecko

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

var (
	ErrCoinNotFound  = errors.New("coin not found")
	ErrCoinAmbiguous = errors.New("coin is ambiguous")
)

func (c *Client) ResolveCoin(
	ctx context.Context,
	query string,
) (*domain.Coin, error) {

	query = strings.TrimSpace(query)

	if query == "" {
		return nil, ErrCoinNotFound
	}

	endpoint, err := url.Parse(baseURL + "/search")
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("query", query)
	endpoint.RawQuery = params.Encode()

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

	var result SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return resolveSearchResult(query, result.Coins)
}

func resolveSearchResult(
	query string,
	coins []SearchCoin,
) (*domain.Coin, error) {

	normalizedQuery := strings.ToLower(
		strings.TrimSpace(query),
	)

	// CoinGecko ID имеет наивысший приоритет.
	for _, coin := range coins {
		if strings.EqualFold(coin.ID, normalizedQuery) {
			return searchCoinToDomain(coin), nil
		}
	}

	var symbolMatches []SearchCoin

	for _, coin := range coins {
		if strings.EqualFold(
			coin.Symbol,
			normalizedQuery,
		) {
			symbolMatches = append(
				symbolMatches,
				coin,
			)
		}
	}

	if len(symbolMatches) == 0 {
		return nil, ErrCoinNotFound
	}

	if len(symbolMatches) == 1 {
		return searchCoinToDomain(
			symbolMatches[0],
		), nil
	}

	// Если несколько монет имеют одинаковый тикер,
	// выбираем монету с лучшим MarketCapRank.
	var best *SearchCoin

	for i := range symbolMatches {
		coin := &symbolMatches[i]

		if coin.MarketCapRank == nil {
			continue
		}

		if best == nil ||
			best.MarketCapRank == nil ||
			*coin.MarketCapRank < *best.MarketCapRank {

			best = coin
		}
	}

	if best != nil {
		return searchCoinToDomain(*best), nil
	}

	return nil, ErrCoinAmbiguous
}

func searchCoinToDomain(
	coin SearchCoin,
) *domain.Coin {

	return &domain.Coin{
		ID:     coin.ID,
		Symbol: strings.ToUpper(coin.Symbol),
		Name:   coin.Name,
	}
}
