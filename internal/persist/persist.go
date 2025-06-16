package persist

import (
	"context"
	"github.com/folivorra/goRedis/internal/model"
	"time"
)

type Loader interface {
	Load(ctx context.Context) (map[int]model.Item, error)
}

type Dumper interface {
	Dump(ctx context.Context, data map[int]model.Item) error
}

type TTLSetter interface {
	DumpTTL(ctx context.Context, data map[int]model.Item, ttl time.Duration) error
}
