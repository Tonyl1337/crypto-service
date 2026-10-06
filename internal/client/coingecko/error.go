package coingecko

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf(
			"CoinGecko unexpected status %d: body=%q retry-after=%s",
			e.StatusCode,
			e.Body,
			e.RetryAfter,
		)
	}

	return fmt.Sprintf(
		"CoinGecko unexpected status %d: body=%q",
		e.StatusCode,
		e.Body,
	)
}

func parseRetryAfter(
	header string,
	now time.Time,
) time.Duration {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(header); err == nil {
		if seconds <= 0 {
			return 0
		}

		return time.Duration(seconds) * time.Second
	}

	retryAt, err := http.ParseTime(header)
	if err != nil {
		return 0
	}

	delay := retryAt.Sub(now)
	if delay <= 0 {
		return 0
	}

	return delay
}

func newAPIError(
	resp *http.Response,
) error {
	body, err := io.ReadAll(
		io.LimitReader(resp.Body, 4096),
	)
	if err != nil {
		return fmt.Errorf(
			"CoinGecko unexpected status %d; failed to read response body: %w",
			resp.StatusCode,
			err,
		)
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Body:       strings.TrimSpace(string(body)),
		RetryAfter: parseRetryAfter(
			resp.Header.Get("Retry-After"),
			time.Now(),
		),
	}
}
