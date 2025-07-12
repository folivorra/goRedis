package storage

import (
	"context"
	"github.com/folivorra/dumpd/application"
	"github.com/folivorra/dumpd/internal/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedisClient(ctx context.Context, app *application.App) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if err := rdb.Ping(timeout).Err(); err != nil {
		logger.ErrorLogger.Println("redis init error:", err)
	}

	app.RegisterCleanup(func(ctx context.Context) {
		timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		if err := rdb.Ping(timeout).Err(); err != nil {
			logger.ErrorLogger.Println("redis connection error:", err)
		} else if err := rdb.Close(); err != nil {
			logger.ErrorLogger.Println("redis close error:", err)
		}
	})

	return rdb
}
