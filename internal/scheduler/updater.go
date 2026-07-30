package scheduler

import (
	"context"
	"log"
	"time"
)

type RateUpdater interface {
	UpdateRates(ctx context.Context) error
}

type Updater struct {
	service  RateUpdater
	interval time.Duration
}

func NewUpdater(
	service RateUpdater,
	interval time.Duration,
) *Updater {

	return &Updater{
		service:  service,
		interval: interval,
	}
}

func (u *Updater) Start(ctx context.Context) {

	go func() {

		if err := u.service.UpdateRates(ctx); err != nil {
			log.Println("initial update:", err)
		}

		ticker := time.NewTicker(u.interval)
		defer ticker.Stop()

		for {

			select {

			case <-ticker.C:

				if err := u.service.UpdateRates(ctx); err != nil {
					log.Println(err)
				}

			case <-ctx.Done():
				return
			}

		}

	}()
}