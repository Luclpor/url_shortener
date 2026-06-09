package auth

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
)

type contextKey string

const userContextKey contextKey = "user"

// SetUserOnContext stores an authenticated user in the request context.
func (ua *UserAuth) SetUserOnContext(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext returns the authenticated user from the request context.
func (ua *UserAuth) GetUserFromContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(userContextKey).(*model.User)
	if !ok {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}
