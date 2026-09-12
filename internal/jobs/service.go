package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/luisync/Job-Queue/internal/adapters/postgresql/sqlc"
)

// Functions a struct must define to be considered a service.
type Service interface {
	ListJobs(ctx context.Context) ([]repo.Job, error)
	FindJob(ctx context.Context, jobID pgtype.UUID) (repo.Job, error)
}

// Services depend on the repository.
type svc struct {
	repo repo.Querier
}

// Constructor for creating the services.
func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

// List all jobs.
func (s *svc) ListJobs(ctx context.Context) ([]repo.Job, error) {
	return s.repo.ListJobs(ctx)
}

// Find a job by ID.
func (s *svc) FindJob(ctx context.Context, jobID pgtype.UUID) (repo.Job, error) {
	return s.repo.FindJobByID(ctx, jobID)
}
