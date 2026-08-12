package rest

import (
	"context"
	stdhttp "net/http"
	"time"

	"github.com/Tonyl1337/crypto-service/internal/transport/rest/handler"
)

type Server struct {
	server *stdhttp.Server
}

func NewServer(
	address string,
	rateHandler *handler.RateHandler,
) *Server {

	mux := stdhttp.NewServeMux()

	mux.HandleFunc(
		"GET /rates",
		rateHandler.GetLatest,
	)

	mux.HandleFunc(
		"GET /rates/{symbol}",
		rateHandler.GetBySymbol,
	)

	server := &stdhttp.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		server: server,
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}