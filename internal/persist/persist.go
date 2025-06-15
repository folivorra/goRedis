package persist

import "github.com/folivorra/studyDir/tree/develop/goRedis/internal/model"

type Persister interface {
	Load() (map[int]model.Item, error)
	Dump(data map[int]model.Item) error
}
