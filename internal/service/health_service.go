package service

import "context"

type HealthService struct {
	checker HealthChecker
}

func NewHealthService(checker HealthChecker) *HealthService {
	return &HealthService{checker: checker}
}

func (s *HealthService) Ping(ctx context.Context) error {
	return s.checker.Ping(ctx)
}
