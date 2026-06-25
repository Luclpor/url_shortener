package service

import "context"

// HealthChecker defines a dependency that can report its liveness.
type HealthChecker interface {
	// Ping verifies that the dependency is available.
	Ping(ctx context.Context) error
}
