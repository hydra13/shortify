package inmemorydb

import (
	"context"
	"sync"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type InMemoryDB struct {
	repository map[string]string
	mutex      sync.RWMutex
}

func New() repository.Repository {
	repo := make(map[string]string)

	return &InMemoryDB{
		repository: repo,
	}
}

func (r *InMemoryDB) Add(_ context.Context, key string, value string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.repository[key] = value

	return nil
}

func (r *InMemoryDB) AddBatch(_ context.Context, keyValue map[string]string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for k, v := range keyValue {
		r.repository[k] = v
	}
	return nil
}

func (r *InMemoryDB) Get(_ context.Context, key string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	value, found := r.repository[key]

	if !found {
		return "", repository.ErrKeyNotFound
	}

	return value, nil
}

func (r *InMemoryDB) GetAll(_ context.Context) (map[string]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	res := make(map[string]string, len(r.repository))
	for k, v := range r.repository {
		res[k] = v
	}
	return res, nil
}

func (r *InMemoryDB) Delete(_ context.Context, key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, found := r.repository[key]

	if !found {
		return repository.ErrKeyNotFound
	}

	delete(r.repository, key)

	return nil
}
