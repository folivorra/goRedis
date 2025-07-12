package persist

import (
	"context"
	"github.com/folivorra/dumpd/application"
	"github.com/folivorra/dumpd/internal/logger"
	"github.com/folivorra/dumpd/internal/storage"
	"sort"
	"time"
)

type Manager struct {
	store      storage.Storager
	persisters []*PriorityPersister
	ttl        time.Duration
}

func NewManager(store storage.Storager, app *application.App, persisters []*PriorityPersister, ttl time.Duration) *Manager {
	sort.Slice(persisters, func(i, j int) bool {
		return persisters[i].priority < persisters[j].priority
	})

	m := &Manager{
		store:      store,
		persisters: persisters,
		ttl:        ttl,
	}

	app.RegisterCleanup(func(ctx context.Context) {
		m.Stop()
	})

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
				m.tryDump(ctx, m.ttl, true)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	ctx := context.Background()
	m.tryDump(ctx, 0, false)
}

func (m *Manager) tryDump(ctx context.Context, ttl time.Duration, stopOnSuccess bool) {
	snap := m.store.Snapshot()

	for _, p := range m.persisters {
		if txPers, ok := p.pers.(TxPersister); ok {
			tx, err := txPers.BeginTx(ctx, 4)
			if err != nil {
				logger.ErrorLogger.Println(p.name, "tx begin failed:", err)
				continue
			}
			txPers.UseTx(tx)

			if err := txPers.Dump(ctx, snap, ttl); err != nil {
				_ = tx.Rollback()
				logger.ErrorLogger.Println(p.name, "rollback:", err)
			} else {
				if err := tx.Commit(); err != nil {
					logger.ErrorLogger.Println(p.name, "commit failed:", err)
				} else if stopOnSuccess {
					return
				}
			}
		} else {
			if err := p.pers.Dump(ctx, snap, ttl); err != nil {
				logger.ErrorLogger.Println(p.name, "dump failed:", err)
			} else if stopOnSuccess {
				return
			}
		}
	}
}
