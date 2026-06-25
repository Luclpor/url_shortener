package storage

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

// UserRepository defines storage operations for authenticated users.
type UserRepository interface {
	// AddNewUser creates and stores a new user.
	AddNewUser(ctx context.Context) (*model.User, error)
	// GetUserByID loads a user by its identifier.
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}
