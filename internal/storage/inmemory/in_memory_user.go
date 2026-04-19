package inmemory

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

func (db *InMemoryDB) AddNewUser(_ context.Context) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	user := model.User{
		ID: uuid.New(),
	}
	db.users = append(db.users, user)
	return &user, nil
}

func (db *InMemoryDB) GetUserByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, user := range db.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, nil
}
