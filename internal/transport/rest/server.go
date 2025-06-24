package rest

import (
	"context"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/storage"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
}

func NewServer(cfg *config.Config, store storage.Storager) *Server {
	c := NewItemController(store)
	r := mux.NewRouter()
	c.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    cfg.Server.HttpPort,
		Handler: r,
	}

	return &Server{
		httpServer: srv,
		router:     r,
	}
}

func (s *Server) Start() error {
	logger.InfoLogger.Println("Starting REST server on port 8080")
	err := s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.InfoLogger.Println("Shutting down server on port 8080")
	return s.httpServer.Shutdown(ctx)
}
