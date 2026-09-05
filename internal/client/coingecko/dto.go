package coingecko

type MarketResponse []MarketData

type MarketData struct {
	ID                      string  `json:"id"`
	Symbol                  string  `json:"symbol"`
	CurrentPrice            float64 `json:"current_price"`
	High24H                 float64 `json:"high_24h"`
	Low24H                  float64 `json:"low_24h"`
	PriceChangePercentage1H float64 `json:"price_change_percentage_1h_in_currency"`
}
