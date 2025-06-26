package persist

import (
	"context"
	"encoding/json"
	"github.com/folivorra/goRedis/internal/model"
	"os"
)

type FilePersister struct {
	path string
}

func NewFilePersister(path string) *FilePersister {
	return &FilePersister{path: path}
}

func (f *FilePersister) Dump(_ context.Context, data map[int64]model.Item) error {
	bytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if err = os.WriteFile(f.path, bytes, 0644); err != nil {
		return err
	}
	return nil
}

func (f *FilePersister) Load(_ context.Context) (map[int64]model.Item, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make(map[int64]model.Item)

	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
