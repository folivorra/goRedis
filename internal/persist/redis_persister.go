package persist

import (
	"context"
	"encoding/json"
	"github.com/folivorra/dumpd/internal/model"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisPersister struct {
	rdb *redis.Client
	key string
}

func NewRedisPersister(rdb *redis.Client, key string) *RedisPersister {
	return &RedisPersister{rdb: rdb, key: key}
}

func (p *RedisPersister) Dump(ctx context.Context, data map[int64]model.Item, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	expire := ttl

	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	return p.rdb.Set(timeout, p.key, bytes, expire).Err()
}

func (p *RedisPersister) Load(ctx context.Context) (map[int64]model.Item, error) {
	timeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	bytes, err := p.rdb.Get(timeout, p.key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	result := make(map[int64]model.Item)
	if err = json.Unmarshal([]byte(bytes), &result); err != nil {
		return nil, err
	}

	return result, nil
}
