package domain

import "time"


type Subscription struct {
	ID              int64
	ChatID          int64
	Enabled         bool
	IntervalMinutes int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}