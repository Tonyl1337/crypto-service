package coingecko

type PriceResponse map[string]Coin

type Coin struct {
	Price     float64 `json:"usd"`
	Change24H float64 `json:"usd_24h_change"`
	High24H   float64 `json:"usd_24h_high"`
	Low24H    float64 `json:"usd_24h_low"`
}
