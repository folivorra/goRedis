package storage

import (
	"context"
	"github.com/folivorra/goRedis/application"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, app *application.App) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if pong := rdb.Ping(ctx); pong == nil {
		logger.ErrorLogger.Fatal("Redis connection error")
	}

	app.RegisterCleanup(func() {
		if err := rdb.Close(); err != nil {
			logger.ErrorLogger.Println(err)
		}
	})

	return rdb
}
