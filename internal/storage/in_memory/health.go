package in_memory

import "context"

type HealthRepository struct{}

func NewHealthRepository() *HealthRepository {
	return &HealthRepository{}
}

func (r *HealthRepository) Ping(_ context.Context) error {
	return nil
}
