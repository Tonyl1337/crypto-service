package coingecko

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetPrices(t *testing.T) {
	client := NewClient()

	prices, err := client.GetPrices(context.Background())

	require.NoError(t, err)

	require.NotEmpty(t, prices)

	_, ok := prices["bitcoin"]
	require.True(t, ok)

	_, ok = prices["ethereum"]
	require.True(t, ok)
}