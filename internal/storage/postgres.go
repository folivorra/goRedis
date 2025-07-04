package storage

import (
	"context"
	"database/sql"
	"github.com/folivorra/goRedis/application"
	"github.com/folivorra/goRedis/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresClient(ctx context.Context, app *application.App, dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.ErrorLogger.Fatalf("Postgres connection error: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		logger.ErrorLogger.Fatalf("Postgres connection error: %v", err)
	}

	app.RegisterCleanup(func(ctx context.Context) {
		if err := db.Close(); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}) // TODO: add timeout ctx to ops with db

	return db
}
