package persist

import (
	"context"
	"encoding/json"
	"github.com/folivorra/goRedis/internal/model"
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

func (p *RedisPersister) Dump(ctx context.Context, data map[int]model.Item) error {
	return p.DumpTTL(ctx, data, 0)
}

func (p *RedisPersister) Load(ctx context.Context) (map[int]model.Item, error) {
	bytes, err := p.rdb.Get(ctx, p.key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	result := make(map[int]model.Item)
	if err = json.Unmarshal([]byte(bytes), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *RedisPersister) DumpTTL(ctx context.Context, data map[int]model.Item, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	expire := ttl

	return p.rdb.Set(ctx, p.key, bytes, expire).Err()
}

func (p *RedisPersister) Close() error {
	return p.rdb.Close()
}
