package jobs

import "context"

// Functions a struct must define to be considered a service.
type Service interface {
	ListJobs(ctx context.Context) error
}

// Services depend on the repository.
type svc struct {
}

// Constructor for creating the services.
func NewService() Service {
	return &svc{}
}

// List all jobs.
func (s *svc) ListJobs(ctx context.Context) error {
	return nil
}
