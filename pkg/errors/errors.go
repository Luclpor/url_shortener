package errors

import "errors"

var (
	// ErrUserNotFound indicates that an authenticated user is missing.
	ErrUserNotFound = errors.New("user not found")
	// ErrAlreadyExists indicates that a URL mapping already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound indicates that a requested record does not exist.
	ErrNotFound = errors.New("not found")
	// ErrURLWasDeleted indicates that a URL mapping was previously deleted.
	ErrURLWasDeleted = errors.New("url was deleted")
)
