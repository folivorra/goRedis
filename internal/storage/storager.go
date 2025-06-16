package storage

import "github.com/folivorra/goRedis/internal/model"

type Storager interface {
	AddItem(item model.Item) (err error)
	GetAllItems() (items []model.Item, err error)
	UpdateItem(item model.Item) (err error)
	DeleteItem(id int) (err error)
	GetItem(id int) (item model.Item, err error)
	Snapshot() map[int]model.Item
	Replace(data map[int]model.Item)
}
