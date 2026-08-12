package response

type Rate struct {
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
	DayLow   float64 `json:"day_low"`
	DayHigh  float64 `json:"day_high"`
	Change1H float64 `json:"change_24h"`
}
