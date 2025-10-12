package inmemorydb

import (
	"github.com/hydra13/shortify/internal/repository"
)

type InMemoryDB struct {
	repository map[string]string
}

func New() repository.Repository {
	repo := make(map[string]string)

	return &InMemoryDB{
		repository: repo,
	}
}

func (r *InMemoryDB) Add(key string, value string) error {
	r.repository[key] = value

	return nil
}

func (r *InMemoryDB) Get(key string) (string, error) {
	value, found := r.repository[key]

	if !found {
		return "", repository.ErrKeyNotFound
	}

	return value, nil
}

func (r *InMemoryDB) Delete(key string) error {
	_, found := r.repository[key]

	if !found {
		return repository.ErrKeyNotFound
	}

	delete(r.repository, key)

	return nil
}
