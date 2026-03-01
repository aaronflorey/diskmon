package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"diskmon/internal/storage"
	"diskmon/web"
)

type Server struct {
	http *http.Server
	log  *slog.Logger
}

func NewServer(addr string, logger *slog.Logger, db *storage.DuckDB) *Server {
	handler := NewRouter(logger, db, web.Assets())
	return &Server{
		log: logger,
		http: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.log.Info("api server listening", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("api server shutting down")
	return s.http.Shutdown(ctx)
}
