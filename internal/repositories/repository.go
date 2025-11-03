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
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type RepositoryErrorMock struct{}

func (e *RepositoryErrorMock) Add(context.Context, string, string) error {
	return ErrImplementation
}

func (e *RepositoryErrorMock) Get(context.Context, string) (string, error) {
	return "", ErrImplementation
}
func (e *RepositoryErrorMock) Delete(context.Context, string) error { return ErrImplementation }
