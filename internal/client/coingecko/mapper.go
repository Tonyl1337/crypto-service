package coingecko

import (
	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func NormalizeSymbol(id string) string {

	switch id {

	case "bitcoin":
		return "BTC"

	case "ethereum":
		return "ETH"

	default:
		return id
	}
}

func ToDomain(resp PriceResponse) []domain.Rate {

	rates := make([]domain.Rate, 0, len(resp))

	for symbol, coin := range resp {

		rates = append(rates, domain.Rate{
			Symbol:   NormalizeSymbol(symbol),
			Price:    coin.Price,
			Change1H: coin.Change24H,
			DayHigh:  coin.High24H,
			DayLow:   coin.Low24H,
		})
	}

	return rates
}
