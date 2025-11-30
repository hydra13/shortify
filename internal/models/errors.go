package models

import "errors"

var (
	ErrValidation    = errors.New("validation error")
	ErrInternal      = errors.New("internal error")
	ErrConflict      = errors.New("conflict error")
	ErrTokenNotValid = errors.New("token expired error")
)
