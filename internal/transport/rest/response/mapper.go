package response

import "github.com/Tonyl1337/crypto-service/internal/domain"

func FromDomain(rate domain.Rate) Rate {
	return Rate{
		Symbol:   rate.Symbol,
		Price:    rate.Price,
		DayLow:   rate.DayLow,
		DayHigh:  rate.DayHigh,
		Change1H: rate.Change1H,
	}
}

func FromDomainList(rates []domain.Rate) []Rate {

	result := make([]Rate, 0, len(rates))

	for _, rate := range rates {
		result = append(result, FromDomain(rate))
	}

	return result
}
