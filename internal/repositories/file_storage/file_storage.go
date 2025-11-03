package filestorage

import (
	"context"
	"sync"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type FileStorage struct {
	repository map[string]*Record
	values     []*Record
	mutex      sync.RWMutex
	freeID     RecordID
	filePath   string
}

func New(filePath string) repository.Repository {
	repo := make(map[string]*Record)
	vals := make([]*Record, 0)

	fs := &FileStorage{
		repository: repo,
		values:     vals,
		filePath:   filePath,
		freeID:     1,
	}

	fs.load()

	return fs
}

func (fs *FileStorage) Add(_ context.Context, key string, value string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	rec := &Record{
		ID:          fs.freeID,
		ShortURL:    key,
		OriginalURL: value,
	}
	fs.repository[key] = rec
	fs.values = append(fs.values, rec)
	fs.freeID++
	fs.write()

	return nil
}

func (fs *FileStorage) Get(_ context.Context, key string) (string, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	value, found := fs.repository[key]

	if !found {
		return "", repository.ErrKeyNotFound
	}

	return value.OriginalURL, nil
}

func (fs *FileStorage) Delete(_ context.Context, key string) error {
	fs.mutex.RLock()
	rec, found := fs.repository[key]
	fs.mutex.RUnlock()

	if !found {
		return repository.ErrKeyNotFound
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	delete(fs.repository, key)

	vals := make([]*Record, 0, len(fs.values)-1)
	for _, v := range fs.values {
		if v.ID != rec.ID {
			vals = append(vals, v)
		}
	}
	fs.values = vals

	if fs.freeID == rec.ID {
		fs.freeID--
	}

	fs.write()

	return nil
}
