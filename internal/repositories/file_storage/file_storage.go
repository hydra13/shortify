package filestorage

import (
	"context"

	"github.com/rs/zerolog"

	repository "github.com/hydra13/shortify/internal/repositories"
	inmemory_db "github.com/hydra13/shortify/internal/repositories/inmemory_db"
)

type FileStorage struct {
	inMemory repository.Repository
	filePath string
	log      zerolog.Logger
}

func New(filePath string, log zerolog.Logger) (repository.Repository, error) {
	inMemoryDB := inmemory_db.New()

	fs := &FileStorage{
		inMemory: inMemoryDB,
		filePath: filePath,
		log:      log,
	}

	err := fs.load()

	return fs, err
}

func (fs *FileStorage) Add(ctx context.Context, key string, value string, userID string) error {
	fs.inMemory.Add(ctx, key, value, userID)

	return fs.write(ctx)
}

func (fs *FileStorage) AddBatch(ctx context.Context, batch map[string]string, userID string) error {
	fs.inMemory.AddBatch(ctx, batch, userID)

	return fs.write(ctx)
}

func (fs *FileStorage) Get(ctx context.Context, key string) (string, error) {
	return fs.inMemory.Get(ctx, key)
}

func (fs *FileStorage) GetAll(ctx context.Context) (map[string]string, error) {
	return fs.inMemory.GetAll(ctx)
}

func (fs *FileStorage) GetAllByUser(ctx context.Context, userID string) (map[string]string, error) {
	return fs.inMemory.GetAllByUser(ctx, userID)
}

func (fs *FileStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return fs.inMemory.GetShortURL(ctx, originalURL)
}

func (fs *FileStorage) Delete(ctx context.Context, key string) error {
	err := fs.inMemory.Delete(ctx, key)
	if err != nil {
		return err
	}

	return fs.write(ctx)
}
