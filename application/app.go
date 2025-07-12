package application

import (
	"context"
	"github.com/folivorra/dumpd/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cleanup    []func(ctx context.Context)
	shutdownCh chan os.Signal
}

func NewApp(ctx context.Context) *App {
	return &App{
		ctx:        ctx,
		shutdownCh: make(chan os.Signal),
	}
}

func (a *App) Start() {
	signal.Notify(a.shutdownCh, os.Interrupt, syscall.SIGTERM)
}

func (a *App) Wait() {
	<-a.shutdownCh
}

func (a *App) Shutdown() {
	logger.InfoLogger.Println("shutting down...")

	ctx := context.Background()
	for i := len(a.cleanup) - 1; i >= 0; i-- {
		a.cleanup[i](ctx)
	}

	logger.InfoLogger.Println("shutdown complete")
}

func (a *App) RegisterCleanup(f func(ctx context.Context)) {
	a.cleanup = append(a.cleanup, f)
}
