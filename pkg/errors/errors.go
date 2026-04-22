package errors

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrURLWasDeleted = errors.New("url was deleted")
)
