package storage

import "github.com/folivorra/goRedis/internal/model"

type Storager interface {
	CreateItem(item model.Item) (err error)
	GetAllItems() (items []model.Item, err error)
	UpdateItem(item model.Item) (err error)
	DeleteItem(id int64) (err error)
	GetItem(id int64) (item model.Item, err error)
	Snapshot() map[int64]model.Item
	Replace(data map[int64]model.Item)
}
