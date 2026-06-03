package postgres

import (
	"context"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

func (r *URLRepository) AddNewUser(ctx context.Context) (*model.User, error) {
	const query = `
		INSERT INTO service_user DEFAULT VALUES
		RETURNING id
	`

	var u = new(model.User)
	err := r.pool.QueryRow(ctx, query).Scan(&u.ID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *URLRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	const query = `
		SELECT id
		FROM service_user
		WHERE id = $1
	`

	var u model.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(&u.ID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
