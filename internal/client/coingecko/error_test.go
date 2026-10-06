package coingecko

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseRetryAfter_Seconds(t *testing.T) {
	now := time.Date(
		2026,
		time.October,
		6,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	got := parseRetryAfter("60", now)

	require.Equal(
		t,
		time.Minute,
		got,
	)
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	now := time.Date(
		2026,
		time.October,
		6,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	retryAt := now.Add(2 * time.Minute)

	got := parseRetryAfter(
		retryAt.Format(http.TimeFormat),
		now,
	)

	require.Equal(
		t,
		2*time.Minute,
		got,
	)
}

func TestParseRetryAfter_Invalid(t *testing.T) {
	now := time.Now()

	require.Zero(
		t,
		parseRetryAfter("", now),
	)

	require.Zero(
		t,
		parseRetryAfter("invalid", now),
	)

	require.Zero(
		t,
		parseRetryAfter("0", now),
	)

	require.Zero(
		t,
		parseRetryAfter("-10", now),
	)
}
