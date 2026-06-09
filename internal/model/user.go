package model

import "github.com/google/uuid"

// User represents an authenticated service user.
type User struct {
	// ID is the unique user identifier stored in the authentication cookie.
	ID uuid.UUID
}
