package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
)

// Functions a struct must define to be considered a service.
type Service interface {
	ListJobs(ctx context.Context, id pgtype.UUID) ([]repo.Job, error)
	FindJob(ctx context.Context, ids findJobByIDParams) (repo.Job, error)
	CreateJob(ctx context.Context, newJob createJobParams) (repo.Job, error)
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
func (s *svc) ListJobs(ctx context.Context, id pgtype.UUID) ([]repo.Job, error) {
	return s.repo.ListJobs(ctx, id)
}

// Find a job by ID.
func (s *svc) FindJob(ctx context.Context, ids findJobByIDParams) (repo.Job, error) {
	return s.repo.FindJobByID(ctx, repo.FindJobByIDParams{
		ID:        ids.ID,
		CreatorID: ids.Creator_id,
	})
}

// Create a job.
func (s *svc) CreateJob(ctx context.Context, newJob createJobParams) (repo.Job, error) {
	return s.repo.CreateJob(ctx, repo.CreateJobParams{
		CreatorID:    newJob.CreatorID,
		Language:     newJob.Language,
		Dependencies: newJob.Dependencies,
		Function:     newJob.Function,
	})
}
