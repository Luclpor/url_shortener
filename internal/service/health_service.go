package service

import "context"

// HealthService wraps a HealthChecker for HTTP health checks.
type HealthService struct {
	checker HealthChecker
}

// NewHealthService creates a health service for the given checker.
func NewHealthService(checker HealthChecker) *HealthService {
	return &HealthService{checker: checker}
}

// Ping delegates the health check to the configured checker.
func (s *HealthService) Ping(ctx context.Context) error {
	return s.checker.Ping(ctx)
}
