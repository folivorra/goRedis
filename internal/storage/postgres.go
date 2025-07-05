package storage

import (
	"context"
	"database/sql"
	"github.com/folivorra/goRedis/application"
	"github.com/folivorra/goRedis/internal/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresClient(ctx context.Context, app *application.App, dsn string) *sql.DB {
	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger.Println("postgres init error: %v", err)
	}

	if err := db.PingContext(timeout); err != nil {
		logger.ErrorLogger.Println("postgres init error: %v", err)
	}

	app.RegisterCleanup(func(ctx context.Context) {
		timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		if err := db.PingContext(timeout); err != nil {
			logger.ErrorLogger.Println("postgres connection error: %v", err)
		} else if err := db.Close(); err != nil {
			logger.ErrorLogger.Println("postgres close error: %v", err)
		}
	})

	return db
}
