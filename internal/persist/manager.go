package persist

import (
	"context"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/storage"
	"sort"
	"time"
)

type Manager struct {
	store      storage.Storager
	persisters []*PriorityPersister
	ttl        time.Duration
}

func NewManager(store storage.Storager, persisters []*PriorityPersister, ttl time.Duration) *Manager {
	sort.Slice(persisters, func(i, j int) bool {
		return persisters[i].priority < persisters[j].priority
	})
	m := &Manager{
		store:      store,
		persisters: persisters,
		ttl:        ttl,
	}
	return m
}

func (m *Manager) Restore(ctx context.Context) {
	for _, p := range m.persisters {
		if data, _ := p.pers.Load(ctx); data != nil {
			m.store.Replace(data)
			logger.InfoLogger.Println(p.name, "data restored")
			return
		}
	}
}

func (m *Manager) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(m.ttl)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.dumpForTTL(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (m *Manager) dumpForTTL(ctx context.Context) {
	snap := m.store.Snapshot()

	for _, p := range m.persisters {
		if err := p.pers.Dump(ctx, snap, m.ttl); err != nil {
			logger.ErrorLogger.Println(p.name, "periodic dump failed:", err)
		} else {
			return
		}
	}
}

func (m *Manager) Stop() {
	snap := m.store.Snapshot()
	ctx := context.Background()

	for _, p := range m.persisters {
		if err := p.pers.Dump(ctx, snap, 0); err != nil {
			logger.ErrorLogger.Println(p.name, "final dump failed:", err)
		}
	}

	//if err := m.r.Close(); err != nil {
	//	logger.WarningLogger.Printf("close redis error: %s", err)
	//}
	//if err := m.p.Close(); err != nil {
	//	logger.WarningLogger.Printf("close postgres error: %s", err)
	//} //TODO: to main
}
