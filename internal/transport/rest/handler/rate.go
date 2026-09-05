package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Tonyl1337/crypto-service/internal/domain"
	"github.com/Tonyl1337/crypto-service/internal/transport/rest/response"
)

type RateService interface {
	GetLatest(ctx context.Context) ([]domain.Rate, error)
	GetBySymbol(ctx context.Context, symbol string) ([]domain.Rate, error)
}

type RateHandler struct {
	service RateService
}

func NewRateHandler(service RateService) *RateHandler {
	return &RateHandler{
		service: service,
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
	symbol := strings.ToUpper(r.PathValue("symbol"))

	if symbol != "BTC" && symbol != "ETH" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			errors.New("invalid cryptocurrency symbol"),
		)
		return
	}

	rates, err := h.service.GetBySymbol(r.Context(), symbol)
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
