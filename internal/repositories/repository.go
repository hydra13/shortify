package repository

import "errors"

var ErrKeyNotFound = errors.New("error: key not found")

type Repository interface {
	Add(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type RepositoryErrorMock struct{}

func (e *RepositoryErrorMock) Add(string, string) error   { return errors.New("test error") }
func (e *RepositoryErrorMock) Get(string) (string, error) { return "", errors.New("test error") }
func (e *RepositoryErrorMock) Delete(string) error        { return errors.New("test error") }
