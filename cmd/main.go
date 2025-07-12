package main

import (
	"context"
	"github.com/folivorra/goRedis/application"
	"github.com/folivorra/goRedis/internal/cli"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/persist"
	"github.com/folivorra/goRedis/internal/storage"
	"github.com/folivorra/goRedis/internal/transport"
	"github.com/folivorra/goRedis/internal/transport/grpc"
	"github.com/folivorra/goRedis/internal/transport/rest"
	"log"
	"net/http"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgPath := os.Getenv("APP_CONFIG")
	if cfgPath == "" {
		cfgPath = "/app/config/app_config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	a := application.NewApp(ctx)
	defer a.Shutdown()

	if err := logger.Init(cfg.Logger.LogFile); err != nil {
		cancel()
		log.Fatal(err)
	}

	store := storage.NewInMemoryStorage()

	rdb := storage.NewRedisClient(ctx, a)

	post := storage.NewPostgresClient(ctx, a, cfg.Storage.PostgresDSN)

	p := persist.NewPriorityPersister(persist.NewPostgresPersister(post))
	r := persist.NewPriorityPersister(persist.NewRedisPersister(rdb, cfg.Storage.RedisKey))
	f := persist.NewPriorityPersister(persist.NewFilePersister(cfg.Storage.DumpFile))
	persisters := []*persist.PriorityPersister{p, r, f}

	pers := persist.NewManager(store, a, persisters, cfg.Storage.TTL)
	pers.Restore(ctx)

	var servers []transport.Server

	grpcSrv, err := grpc.NewServer(cfg, a, store)
	if err != nil {
		cancel()
		logger.ErrorLogger.Fatalf("init grpc server error: %s", err)
	}
	servers = append(servers, grpcSrv)

	httpSrv := rest.NewServer(cfg, a, store)
	servers = append(servers, httpSrv)

	cliManager := cli.NewManager(store)

	pers.Start(ctx)
	cliManager.Start(ctx)
	for _, server := range servers {
		srv := server
		go func() {
			if err := srv.Start(); err != nil && err != http.ErrServerClosed {
				logger.ErrorLogger.Fatalf("start grpc server error: %s", err)
			}
		}()
	}

	a.Start()
	a.Wait()
}
