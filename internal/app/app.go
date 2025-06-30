package app

import (
	"context"
	"fmt"
	"github.com/folivorra/goRedis/internal/cli"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/persist"
	"github.com/folivorra/goRedis/internal/storage"
	"github.com/folivorra/goRedis/internal/transport"
	"github.com/folivorra/goRedis/internal/transport/grpc"
	"github.com/folivorra/goRedis/internal/transport/rest"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	store       *storage.InMemoryStorage
	servers     []transport.ToServe
	persistence *persist.Manager
	cliManager  *cli.Manager
	shutdownCh  chan os.Signal
	cleanup     []func()
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

	// --- postgres ---
	post, err := storage.NewPostgresClient(ctx, cfg.Storage.PostgresDSN)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("init postgres error: %s", err)
	}

	// --- load ---
	p := persist.NewPriorityPersister(persist.NewPostgresPersister(post))
	r := persist.NewPriorityPersister(persist.NewRedisPersister(rdb, cfg.Storage.RedisKey))
	f := persist.NewPriorityPersister(persist.NewFilePersister(cfg.Storage.DumpFile))
	persisters := []*persist.PriorityPersister{p, r, f}

	pers := persist.NewManager(store, persisters, cfg.Storage.TTL)
	pers.Restore(ctx)

	// --- servers ---
	var servers []transport.ToServe

	// --- grpc-server ---
	grpcSrv, err := grpc.NewServer(cfg, store)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("init grpc server error: %s", err)
	}
	servers = append(servers, grpcSrv)

	// --- http-server ---
	httpSrv := rest.NewServer(cfg, store)
	servers = append(servers, httpSrv)

	// --- cli ---
	cliManager := cli.NewManager(store)

	return &App{
		store:       store,
		servers:     servers,
		persistence: pers,
		cliManager:  cliManager,
		shutdownCh:  make(chan os.Signal),
		rootCtx:     ctx,
		ctxCancel:   cancel,
	}, nil
}

func (a *App) Start() {
	a.persistence.Start(a.rootCtx)
	a.RegisterCleanup(func() {
		a.persistence.Stop()
	})

	a.cliManager.Start(a.rootCtx)

	for _, srv := range a.servers {
		srv := srv
		go func() {
			if err := srv.Start(); err != nil && err != http.ErrServerClosed {
				logger.ErrorLogger.Printf("server start error: %s", err)
				a.ctxCancel()
			}
		}()
		a.RegisterCleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				logger.ErrorLogger.Printf("server shutdown error: %s", err)
			}
		})
	}
}

func (a *App) Wait() {
	<-a.rootCtx.Done()
}

func (a *App) Shutdown() {
	logger.InfoLogger.Println("shutting down...")

	a.ctxCancel()

	for _, cleanup := range a.cleanup {
		cleanup()
	}

	logger.InfoLogger.Println("shutdown complete")
}

func (a *App) RegisterCleanup(f func()) {
	a.cleanup = append(a.cleanup, f)
}
