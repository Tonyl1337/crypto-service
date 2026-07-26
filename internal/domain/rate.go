package domain

import "time"

type Rate struct {
	ID        int64
	Symbol    string
	Price     float64
	Change24H  float64
	DayLow    float64
	DayHigh   float64
	CreatedAt time.Time
}