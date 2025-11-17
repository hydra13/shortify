package repository

import (
	"context"
	"errors"
)

var (
	ErrKeyNotFound    = errors.New("error: key not found")
	ErrImplementation = errors.New("error: not implemented")
)

type Repository interface {
	Add(ctx context.Context, key string, value string) error
	AddBatch(ctx context.Context, keyValue map[string]string) error
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	Delete(ctx context.Context, key string) error
}
