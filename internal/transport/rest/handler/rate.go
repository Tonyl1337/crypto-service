package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Tonyl1337/crypto-service/internal/client/coingecko"
	"github.com/Tonyl1337/crypto-service/internal/domain"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/response"
)

type RateService interface {
	GetLatest(ctx context.Context) ([]domain.Rate, error)

	GetCurrentRate(
		ctx context.Context,
		coinGeckoID string,
	) (*domain.Rate, error)
}

type CoinResolver interface {
	ResolveCoin(
		ctx context.Context,
		query string,
	) (*domain.Coin, error)
}

type RateHandler struct {
	service  RateService
	resolver CoinResolver
}

func NewRateHandler(
	service RateService,
	resolver CoinResolver,
) *RateHandler {
	return &RateHandler{
		service:  service,
		resolver: resolver,
	}
}

func (h *RateHandler) GetLatest(
	w http.ResponseWriter,
	r *http.Request,
) {
	rates, err := h.service.GetLatest(r.Context())
	if err != nil {
		response.WriteError(
			w,
			http.StatusInternalServerError,
			err,
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		response.FromDomainList(rates),
	)
}

func (h *RateHandler) GetBySymbol(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := strings.TrimSpace(
		r.PathValue("symbol"),
	)

	if query == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			errors.New("cryptocurrency is required"),
		)
		return
	}

	coin, err := h.resolver.ResolveCoin(
		r.Context(),
		query,
	)
	if err != nil {
		switch {
		case errors.Is(err, coingecko.ErrCoinNotFound):
			response.WriteError(
				w,
				http.StatusNotFound,
				errors.New("cryptocurrency not found"),
			)

		case errors.Is(err, coingecko.ErrCoinAmbiguous):
			response.WriteError(
				w,
				http.StatusBadRequest,
				errors.New("cryptocurrency symbol is ambiguous"),
			)

		default:
			response.WriteError(
				w,
				http.StatusBadGateway,
				errors.New("failed to resolve cryptocurrency"),
			)
		}

		return
	}

	rate, err := h.service.GetCurrentRate(
		r.Context(),
		coin.ID,
	)
	if err != nil {
		response.WriteError(
			w,
			http.StatusBadGateway,
			errors.New("failed to get cryptocurrency rate"),
		)
		return
	}

	if rate == nil {
		response.WriteError(
			w,
			http.StatusNotFound,
			errors.New("cryptocurrency rate not found"),
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		response.FromDomainList(
			[]domain.Rate{*rate},
		),
	)
}
