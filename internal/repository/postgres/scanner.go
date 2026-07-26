package postgres

import (
	"github.com/jackc/pgx/v5"

	"github.com/Tonyl1337/crypto-service/internal/domain"
)

func scanRates(rows pgx.Rows) ([]domain.Rate, error) {
	var rates []domain.Rate

	for rows.Next() {
		var rate domain.Rate

		err := rows.Scan(
			&rate.ID,
			&rate.Symbol,
			&rate.Price,
			&rate.Change24H,
			&rate.DayLow,
			&rate.DayHigh,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		rates = append(rates, rate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rates, nil
}