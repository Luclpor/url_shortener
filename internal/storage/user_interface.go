package storage

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddNewUser(ctx context.Context) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type UserAuthentication interface {
	CreateEncryptedUser() (*model.User, string, error)
	DecryptUser(value string) (*model.User, error)
	GetUserFromContext(ctx context.Context) (*model.User, error)
	SetUserOnContext(ctx context.Context, user *model.User) context.Context
}
