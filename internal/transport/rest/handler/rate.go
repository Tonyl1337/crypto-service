package handler

import (
	"context"
	"net/http"

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

	http.Error(
		w,
		"not implemented",
		http.StatusNotImplemented,
	)
}