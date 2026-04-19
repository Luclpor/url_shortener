package errors

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")
