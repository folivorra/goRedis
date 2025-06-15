package persist

import (
	"encoding/json"
	"github.com/folivorra/studyDir/tree/develop/goRedis/internal/model"
	"os"
)

type FilePersister struct {
	path string
}

func NewFilePersister(path string) *FilePersister {
	return &FilePersister{path: path}
}

func (f *FilePersister) Dump(data map[int]model.Item) error {
	bytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if err = os.WriteFile(f.path, bytes, 0644); err != nil {
		return err
	}
	return nil
}

func (f *FilePersister) Load() (map[int]model.Item, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make(map[int]model.Item)

	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
