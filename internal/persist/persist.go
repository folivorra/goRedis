package persist

import (
	"context"
	"github.com/folivorra/goRedis/internal/model"
	"time"
)

type Loader interface {
	Load(ctx context.Context) (map[int64]model.Item, error)
}

type Dumper interface {
	Dump(ctx context.Context, data map[int64]model.Item, ttl time.Duration) error
}

type Persister interface {
	Loader
	Dumper
}
