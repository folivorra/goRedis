package server

import (
	"context"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/controller"
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
	c := controller.NewItemController(store)
	r := mux.NewRouter()
	c.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	return &Server{
		httpServer: srv,
		router:     r,
	}
}

func (s *Server) Start() error {
	logger.InfoLogger.Println("Starting server on port 8080")
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
