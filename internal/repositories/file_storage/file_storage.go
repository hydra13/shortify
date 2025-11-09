package filestorage

import (
	"context"

	"github.com/rs/zerolog"

	repository "github.com/hydra13/shortify/internal/repositories"
)

type FileStorage struct {
	inMemory repository.Repository
	filePath string
	log      zerolog.Logger
}

func New(filePath string, inMemoryDB repository.Repository, log zerolog.Logger) (repository.Repository, error) {
	fs := &FileStorage{
		inMemory: inMemoryDB,
		filePath: filePath,
		log:      log,
	}

	err := fs.load()

	return fs, err
}

func (fs *FileStorage) Add(ctx context.Context, key string, value string) error {
	fs.inMemory.Add(ctx, key, value)

	return fs.write(ctx)
}

func (fs *FileStorage) Get(ctx context.Context, key string) (string, error) {
	return fs.inMemory.Get(ctx, key)
}

func (fs *FileStorage) GetAll(ctx context.Context) (map[string]string, error) {
	return fs.inMemory.GetAll(ctx)
}

func (fs *FileStorage) Delete(ctx context.Context, key string) error {
	err := fs.inMemory.Delete(ctx, key)
	if err != nil {
		return err
	}

	return fs.write(ctx)
}
