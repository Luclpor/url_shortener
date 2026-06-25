package inmemory

import "context"

// HealthRepository reports health for the in-memory storage backend.
type HealthRepository struct{}

// NewHealthRepository creates an in-memory health checker.
func NewHealthRepository() *HealthRepository {
	return &HealthRepository{}
}

// Ping always succeeds for the in-memory storage backend.
func (r *HealthRepository) Ping(_ context.Context) error {
	return nil
}
