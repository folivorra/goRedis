package app

import (
	"context"
	"fmt"
	"github.com/folivorra/goRedis/internal/cli"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/persist"
	"github.com/folivorra/goRedis/internal/server"
	"github.com/folivorra/goRedis/internal/storage"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	store       *storage.InMemoryStorage
	server      *server.Server
	persistence *persist.Manager
	cliManager  *cli.Manager
	shutdownCh  chan os.Signal
	rootCtx     context.Context
	ctxCancel   context.CancelFunc
}

func NewApp(cfg *config.Config) (*App, error) {
	// --- context ---
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// --- logger ---
	if err := logger.Init(cfg.Logger.LogFile); err != nil {
		cancel()
		return nil, fmt.Errorf("init logger error: %s", err)
	}

	// --- storage ---
	store := storage.NewInMemoryStorage()

	// --- redis ---
	rdb := storage.NewRedisClient()

	// --- load ---
	r := persist.NewRedisPersister(rdb, cfg.Storage.RedisKey)
	f := persist.NewFilePersister(cfg.Storage.DumpFile)
	pers := persist.NewManager(ctx, store, f, r, cfg.Storage.TTL)

	// --- http-server ---
	srv := server.NewServer(cfg, store)

	// --- cli ---
	cliManager := cli.NewManager(store)

	return &App{
		store:       store,
		server:      srv,
		persistence: pers,
		cliManager:  cliManager,
		shutdownCh:  make(chan os.Signal),
		rootCtx:     ctx,
		ctxCancel:   cancel,
	}, nil
}

func (a *App) Start() {
	a.persistence.Start(a.rootCtx)

	a.cliManager.Start(a.rootCtx)

	go func() {
		if err := a.server.Start(); err != nil {
			logger.ErrorLogger.Printf("server start error: %s", err)
			a.ctxCancel()
		}
	}()
}

func (a *App) Wait() {
	<-a.rootCtx.Done()
}

func (a *App) Shutdown() {
	logger.InfoLogger.Println("shutting down...")

	a.ctxCancel()

	a.persistence.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		logger.ErrorLogger.Printf("server shutdown error: %s", err)
	}
	logger.InfoLogger.Println("shutdown complete")
}
