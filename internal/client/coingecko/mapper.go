package coingecko

import (
	"time"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func ToDomain(
	coins MarketResponse,
) []domain.Rate {

	rates := make([]domain.Rate, 0, len(coins))

	for _, coin := range coins {

		symbol := ""

		switch coin.ID {
		case "bitcoin":
			symbol = "BTC"

		case "ethereum":
			symbol = "ETH"

		default:
			continue
		}

		rates = append(rates, domain.Rate{
			Symbol:    symbol,
			Price:     coin.CurrentPrice,
			Change1H:  coin.PriceChangePercentage1H,
			DayLow:    coin.Low24H,
			DayHigh:   coin.High24H,
			CreatedAt: time.Now(),
		})
	}

	return rates
}