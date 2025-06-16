package storage

import (
	"fmt"
	"github.com/folivorra/goRedis/internal/model"
	"sync"
)

type InMemoryStorage struct {
	mu    sync.RWMutex
	items map[int]model.Item
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		items: make(map[int]model.Item),
	}
}

func (s *InMemoryStorage) AddItem(item model.Item) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[item.ID]; ok {
		return fmt.Errorf("item already exists")
	}
	s.items[item.ID] = item
	return nil
}

func (s *InMemoryStorage) GetAllItems() (items []model.Item, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items = make([]model.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items, nil
}

func (s *InMemoryStorage) UpdateItem(item model.Item) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[item.ID]; !ok {
		return fmt.Errorf("item does not exist")
	}
	s.items[item.ID] = item
	return nil
}

func (s *InMemoryStorage) DeleteItem(id int) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return fmt.Errorf("item does not exist")
	}
	delete(s.items, id)
	return nil
}

func (s *InMemoryStorage) GetItem(id int) (item model.Item, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.items[id]; !ok {
		return model.Item{}, fmt.Errorf("item does not exist")
	}
	return s.items[id], nil
}

func (s *InMemoryStorage) Snapshot() map[int]model.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	temp := make(map[int]model.Item, len(s.items))
	for k, v := range s.items {
		temp[k] = v
	}
	return temp
}

func (s *InMemoryStorage) Replace(data map[int]model.Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[int]model.Item, len(data))

	for k, v := range data {
		s.items[k] = v
	}
}
