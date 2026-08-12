package coingecko

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetRates(t *testing.T) {

	client := NewClient()

	rates, err := client.GetRates(context.Background())

	require.NoError(t, err)
	require.NotEmpty(t, rates)

	foundBTC := false
	foundETH := false

	for _, rate := range rates {

		if rate.Symbol == "BTC" {
			foundBTC = true
		}

		if rate.Symbol == "ETH" {
			foundETH = true
		}
	}

	require.True(t, foundBTC)
	require.True(t, foundETH)
}
