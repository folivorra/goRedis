package persist

import (
	"context"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/storage"
	"time"
)

type Manager struct {
	store storage.Storager
	f     *FilePersister
	r     *RedisPersister
	ttl   time.Duration
}

func NewManager(ctx context.Context, store storage.Storager, f *FilePersister, r *RedisPersister, ttl time.Duration) *Manager {
	m := &Manager{
		store: store,
		f:     f,
		r:     r,
		ttl:   ttl,
	}
	m.restore(ctx)
	return m
}

func (m *Manager) restore(ctx context.Context) {
	if data, _ := m.r.Load(ctx); data != nil {
		m.store.Replace(data)
		return
	}
	if data, _ := m.f.Load(ctx); data != nil {
		m.store.Replace(data)
	}
}

func (m *Manager) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(m.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.dumpToRedis(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (m *Manager) dumpToRedis(ctx context.Context) {
	snap := m.store.Snapshot()
	if err := m.r.DumpTTL(ctx, snap, m.ttl); err != nil {
		logger.WarningLogger.Printf("periodic dump to redis error: %s", err)
	}
}

func (m *Manager) Stop() {
	snap := m.store.Snapshot()
	ctx := context.Background()

	if err := m.r.Dump(ctx, snap); err != nil {
		logger.WarningLogger.Printf("dump to redis error: %s", err)
	}
	if err := m.f.Dump(ctx, snap); err != nil {
		logger.WarningLogger.Printf("dump to file error: %s", err)
	}
}
