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

func (p *RedisPersister) DumpTTL(ctx context.Context, data map[int]model.Item, ttlSeconds int) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var expire time.Duration
	if ttlSeconds > 0 {
		expire = time.Duration(ttlSeconds) * time.Second
	}

	return p.rdb.Set(ctx, p.key, bytes, expire).Err()

	//ticker := time.NewTicker(7 * time.Second)
	//defer ticker.Stop()
	//
	//for {
	//	select {
	//	case <-ctx.Done():
	//		return
	//	case <-ticker.C:
	//		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	//		ttl, err := p.rdb.TTL(ctx, p.key).Result()
	//		cancel()
	//		// от зависаний редиса
	//		if err != nil {
	//			logger.WarningLogger.Println("Failed to get TTL:", err)
	//			continue
	//		}
	//
	//		if ttl >= -1 && ttl < 10*time.Second {
	//			snapshot := store.Snapshot()
	//
	//			if err = p.Dump(snapshot, 2*time.Minute); err != nil {
	//				logger.ErrorLogger.Println("Failed to dump snapshot:", err)
	//			} else {
	//				logger.InfoLogger.Println("Snapshot dumped successfully")
	//			}
	//		}
	//	}
	//}
}
