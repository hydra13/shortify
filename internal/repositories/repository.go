package repository

import (
	"context"
	"errors"
)

var (
	ErrKeyNotFound    = errors.New("error: key not found")
	ErrImplementation = errors.New("error: not implemented")
	ErrConflict       = errors.New("error: conflict")
)

type Repository interface {
	Add(ctx context.Context, key string, value string, userID string) error
	AddBatch(ctx context.Context, keyValue map[string]string, userID string) error
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	GetAllByUser(ctx context.Context, userID string) (map[string]string, error)
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	Delete(ctx context.Context, key string) error
}
