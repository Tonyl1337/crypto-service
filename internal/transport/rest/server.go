package rest

import (
	"context"
	"errors"
	"log"
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

	mux.HandleFunc(
		"GET /openapi.yaml",
		func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			stdhttp.ServeFile(
				w,
				r,
				"api/openapi.yaml",
			)
		},
	)

	mux.HandleFunc(
		"GET /swagger/",
		func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			w.Header().Set(
				"Content-Type",
				"text/html; charset=utf-8",
			)

			_, _ = w.Write([]byte(`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Crypto Service API</title>
	
		<link
			rel="stylesheet"
			href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"
		>
	</head>
	
	<body>
		<div id="swagger-ui"></div>
	
		<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	
		<script>
			window.onload = function() {
				SwaggerUIBundle({
					url: "/openapi.yaml",
					dom_id: "#swagger-ui"
				});
			};
		</script>
	</body>
	</html>`))
		},
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

func (s *Server) Run(
	ctx context.Context,
) error {

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			log.Printf(
				"HTTP server shutdown: %v",
				err,
			)
		}
	}()

	log.Printf(
		"HTTP server started on %s",
		s.server.Addr,
	)

	err := s.server.ListenAndServe()

	if errors.Is(err, stdhttp.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
