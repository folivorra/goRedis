package rest

import (
	"context"
	"github.com/folivorra/dumpd/application"
	"github.com/folivorra/dumpd/internal/config"
	"github.com/folivorra/dumpd/internal/logger"
	"github.com/folivorra/dumpd/internal/storage"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	router     *mux.Router
}

func NewServer(cfg *config.Config, app *application.App, store storage.Storager) *Server {
	c := NewItemController(store)
	r := mux.NewRouter()
	c.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    cfg.Server.HttpPort,
		Handler: r,
	}

	s := &Server{
		httpServer: srv,
		router:     r,
	}

	app.RegisterCleanup(func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			logger.ErrorLogger.Printf("Failed to shutdown http server: %v", err)
		}
	})

	return s
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
	logger.InfoLogger.Println("Shutting down http server on port 8080")
	return s.httpServer.Shutdown(ctx)
}
