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
	p     *PostgresPersister
	ttl   time.Duration // TODO: prior
}

func NewManager(ctx context.Context, store storage.Storager, f *FilePersister, r *RedisPersister, p *PostgresPersister, ttl time.Duration) *Manager {
	m := &Manager{
		store: store,
		f:     f,
		r:     r,
		p:     p,
		ttl:   ttl,
	}
	m.restore(ctx)
	return m
}

func (m *Manager) restore(ctx context.Context) {
	if data, _ := m.r.Load(ctx); data != nil {
		m.store.Replace(data)
		logger.InfoLogger.Println("redis data restored")
		return
	}
	if data, _ := m.p.Load(ctx); data != nil {
		m.store.Replace(data)
		logger.InfoLogger.Println("postgres data restored")
		return
	}
	if data, _ := m.f.Load(ctx); data != nil {
		m.store.Replace(data)
		logger.InfoLogger.Println("file data restored")
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
	} // TODO: if error to next persister
}

func (m *Manager) Stop() {
	snap := m.store.Snapshot()
	ctx := context.Background() // TODO: переделать

	if err := m.r.Dump(ctx, snap); err != nil {
		logger.WarningLogger.Printf("dump to redis error: %s", err)
	}
	if err := m.p.Dump(ctx, snap); err != nil {
		logger.WarningLogger.Printf("dump to postgres error: %s", err)
	}
	if err := m.f.Dump(ctx, snap); err != nil {
		logger.WarningLogger.Printf("dump to file error: %s", err)
	} // TODO: to centralize
	if err := m.r.Close(); err != nil {
		logger.WarningLogger.Printf("close redis error: %s", err)
	}
	if err := m.p.Close(); err != nil {
		logger.WarningLogger.Printf("close postgres error: %s", err)
	}
}
