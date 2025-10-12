package repository

import "errors"

var ErrKeyNotFound = errors.New("error: key not found")

type Repository interface {
	Add(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}
