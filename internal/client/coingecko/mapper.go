package coingecko

import (
	"strings"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func ToDomain(
	coins MarketResponse,
) []domain.Rate {

	rates := make([]domain.Rate, 0, len(coins))

	now := time.Now()

	for _, coin := range coins {
		rates = append(rates, domain.Rate{
			CoinGeckoID: coin.ID,
			Symbol:      strings.ToUpper(coin.Symbol),
			Price:       coin.CurrentPrice,
			Change1H:    coin.PriceChangePercentage1H,
			DayLow:      coin.Low24H,
			DayHigh:     coin.High24H,
			CreatedAt:   now,
		})
	}

	return rates
}
