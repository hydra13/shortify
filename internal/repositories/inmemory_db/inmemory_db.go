package inmemorydb

import (
	"context"
	"fmt"
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

func (r *InMemoryDB) printDebug(action string) {
	fmt.Println("DEBUG: repo:", r.repository, "; action:", action)
}

func (r *InMemoryDB) Add(_ context.Context, key string, value string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.repository[key] = value

	r.printDebug("Add")

	return nil
}

func (r *InMemoryDB) Get(_ context.Context, key string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	r.printDebug("Get")
	value, found := r.repository[key]

	if !found {
		return "", repository.ErrKeyNotFound
	}

	return value, nil
}

func (r *InMemoryDB) Delete(_ context.Context, key string) error {
	r.mutex.RLock()
	r.printDebug("Delete")
	_, found := r.repository[key]
	r.mutex.RUnlock()

	if !found {
		return repository.ErrKeyNotFound
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.repository, key)

	return nil
}
