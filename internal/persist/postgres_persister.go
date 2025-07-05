package persist

import (
	"context"
	"database/sql"
	"github.com/folivorra/goRedis/internal/model"
	"time"
)

type PostgresPersister struct {
	db *sql.DB
}

func NewPostgresPersister(db *sql.DB) *PostgresPersister {
	return &PostgresPersister{db: db}
}

func (p *PostgresPersister) Dump(ctx context.Context, data map[int64]model.Item, _ time.Duration) error {
	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	queryDelete := `DELETE FROM items`
	_, err := p.db.ExecContext(timeout, queryDelete)
	if err != nil {
		return err
	}

	queryInsert := `INSERT INTO items (id, name, price) VALUES ($1, $2, $3)`
	for _, item := range data {
		_, err := p.db.ExecContext(ctx, queryInsert, item.ID, item.Name, item.Price)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresPersister) Load(ctx context.Context) (map[int64]model.Item, error) {
	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	rows, err := p.db.QueryContext(timeout, "SELECT id, name, price FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result map[int64]model.Item
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		if result == nil {
			result = make(map[int64]model.Item, 50)
		}
		result[item.ID] = item
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
