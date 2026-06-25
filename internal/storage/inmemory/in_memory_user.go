package inmemory

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

// AddNewUser creates and stores a new in-memory user.
func (db *Repository) AddNewUser(_ context.Context) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	user := model.User{
		ID: uuid.New(),
	}
	db.users = append(db.users, user)
	return &user, nil
}

// GetUserByID returns an in-memory user by ID.
func (db *Repository) GetUserByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, user := range db.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, nil
}
