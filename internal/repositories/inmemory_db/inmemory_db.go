package inmemorydb

import (
	"context"
	"sync"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type Record struct {
	OriginalURL string
	UserID      string
}

type InMemoryDB struct {
	repository map[string]Record
	mutex      sync.RWMutex
}

func New() repository.Repository {
	repo := make(map[string]Record)

	return &InMemoryDB{
		repository: repo,
	}
}

func (r *InMemoryDB) Add(_ context.Context, shortURL string, originalURL string, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, found := r.repository[shortURL]
	if found {
		return repository.ErrConflict
	}

	r.repository[shortURL] = Record{
		OriginalURL: originalURL,
		UserID:      userID,
	}

	return nil
}

func (r *InMemoryDB) AddBatch(_ context.Context, urls map[string]string, userID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for shortURL, originalURL := range urls {
		r.repository[shortURL] = Record{
			OriginalURL: originalURL,
			UserID:      userID,
		}
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

	return value.OriginalURL, nil
}

func (r *InMemoryDB) GetAll(_ context.Context) (map[string]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	res := make(map[string]string, len(r.repository))
	for shortURL, record := range r.repository {
		res[shortURL] = record.OriginalURL
	}
	return res, nil
}

func (r *InMemoryDB) GetAllByUser(ctx context.Context, userID string) (map[string]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	res := make(map[string]string)

	for shortURL, record := range r.repository {
		if record.UserID == userID {
			res[shortURL] = record.OriginalURL
		}
	}

	return res, nil
}

func (r *InMemoryDB) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for shortURL, record := range r.repository {
		if record.OriginalURL == originalURL {
			return shortURL, nil
		}
	}
	return "", repository.ErrKeyNotFound
}

func (r *InMemoryDB) Delete(_ context.Context, shortURL string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, found := r.repository[shortURL]

	if !found {
		return repository.ErrKeyNotFound
	}

	delete(r.repository, shortURL)

	return nil
}
