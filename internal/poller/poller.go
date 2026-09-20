package poller

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"uuid"

	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
	"github.com/redis/go-redis/v9"
)

type poller struct {
	streamKey   string
	batchSize   int32
	redisClient *redis.Client
	repo        repo.Querier
}

func NewPoller(streamKey string, batchSize int32, redisClient *redis.Client, repo repo.Querier) *poller {
	return &poller{
		streamKey:   streamKey,
		batchSize:   batchSize,
		redisClient: redisClient,
		repo:        repo,
	}
}

// Fetch and populate redis with job data.
func (s *poller) StartPoller(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Poller started", "interval", interval, "batch_size", s.batchSize)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping poller...")
			return
		case <-ticker.C:
			count, err := s.EnqueuePendingJobs(ctx)
			if err != nil {
				slog.Error("Failed to poll pending jobs", "error", err)
			} else if count > 0 {
				slog.Info("Enqueued pending jobs into Redis", "count", count)
			}
		}
	}
}

// Query main database and attempt to insert the result into redis.
func (s *poller) EnqueuePendingJobs(ctx context.Context) (int, error) {
	jobIDs, err := s.repo.SchedulerFetchAndLockPendingJobs(ctx, s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("Failed to fetch jobs from the database, %w", err)
	}

	if len(jobIDs) == 0 {
		return 0, nil
	}

	// Insert job data into redis via a pieline.
	pipe := s.redisClient.Pipeline()

	// Insert each id into the pipeline.
	for _, id := range jobIDs {
		// Format UUID string.
		jobIDStr := uuid.UUID(id.Bytes).String()

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: s.streamKey,
			MaxLen: 10000,
			Approx: true,
			Values: map[string]any{
				"job_id": jobIDStr,
			},
		})
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("Failed to execute pipeline, %w", err)
	}

	return len(jobIDs), nil
}
