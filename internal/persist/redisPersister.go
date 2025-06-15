package persist

import (
	"context"
	"encoding/json"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/logger"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/model"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/storage"
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

func (p *RedisPersister) Dump(data map[int]model.Item, ttl time.Duration) error {
	ctx := context.Background()

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err = p.rdb.Set(ctx, p.key, string(bytes), ttl).Err(); err != nil {
		return err
	}

	return nil
}

func (p *RedisPersister) Load() (map[int]model.Item, error) {
	ctx := context.Background()

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

func (p *RedisPersister) DumpForTTL(ctx context.Context, store *storage.InMemoryStorage) {
	ticker := time.NewTicker(7 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			ttl, err := p.rdb.TTL(ctx, p.key).Result()
			cancel()
			// от зависаний редиса
			if err != nil {
				logger.WarningLogger.Println("Failed to get TTL:", err)
				continue
			}

			if ttl >= -1 && ttl < 10*time.Second {
				snapshot := store.Snapshot()

				if err = p.Dump(snapshot, 2*time.Minute); err != nil {
					logger.ErrorLogger.Println("Failed to dump snapshot:", err)
				} else {
					logger.InfoLogger.Println("Snapshot dumped successfully")
				}
			}
		}
	}
}
