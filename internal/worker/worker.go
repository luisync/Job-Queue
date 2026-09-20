package worker

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
	"github.com/redis/go-redis/v9"
)

type worker struct {
	id          string
	config      config
	redisClient *redis.Client
	repo        repo.Querier
}

type config struct {
	groupName     string
	streamKey     string
	minIdleTime   time.Duration
	claimInterval time.Duration
}

var ErrNoMessages = errors.New("No messages.")

func NewWorkerConfig(groupName string, streamKey string, minIdleTime time.Duration, claimInterval time.Duration) config {
	return config{
		groupName:     groupName,
		streamKey:     streamKey,
		minIdleTime:   minIdleTime,
		claimInterval: claimInterval,
	}
}

func NewWorker(config config, redisClient *redis.Client, repo repo.Querier) (*worker, error) {
	workerID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("Failed to generate worker id, %w", err)
	}

	return &worker{
		id:          workerID.String(),
		config:      config,
		redisClient: redisClient,
		repo:        repo,
	}, nil
}

func (w *worker) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start recovery routine for abandoned jobs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerName := fmt.Sprintf("%s-recovery", w.id)
		w.StartRecoveryLoop(ctx, workerName)
	}()

	// Create 2 workers.
	for i := range 2 {
		wg.Add(1)
		workerName := fmt.Sprintf("%s-%d", w.id, i)
		go workerUnit(ctx, w, &wg, workerName)
	}

	// Stop the workers when the program terminates.
	<-ctx.Done()
	slog.Info("Shutting down workers...")

	wg.Wait()
	slog.Info("All workers have stopped.")

	return nil
}

// Create a worker.
func workerUnit(ctx context.Context, w *worker, wg *sync.WaitGroup, workerName string) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker exiting", "worker_name", workerName)
			return
		default:
			// Extract job from redis.
			jobID, messageID, err := w.extract(ctx, workerName)
			if err != nil {
				// Check whether the error was caused by a an empty response.
				if errors.Is(err, ErrNoMessages) {
					continue
				}

				slog.Error("Failed to read messages", "worker_name", workerName, "error", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// Query the job details from the main database.
			job, err := w.query(ctx, jobID)
			if err != nil {
				// Job wasn't able to be read.
				slog.Error("Job query failed", "job_id", jobID, "error", err)
				continue
			}

			// Execute the job and call the repository.
			if err := w.execute(ctx, job, jobID, messageID); err != nil {
				slog.Error("Job execution failed", "job_id", jobID, "error", err)
				continue
			}
		}
	}
}

// Extract a job from redis storage.
func (w *worker) extract(ctx context.Context, workerName string) (string, string, error) {
	// Fecth messages from the stream.
	stream, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.config.groupName,
		Consumer: workerName,
		Streams:  []string{w.config.streamKey, ">"},
		Count:    1,
		Block:    3 * time.Second,
	}).Result()

	// Check whether the error is caused by an empty return.
	if errors.Is(err, redis.Nil) || (err == nil && (len(stream) == 0 || len(stream[0].Messages) == 0)) {
		return "", "", ErrNoMessages
	}
	if err != nil {
		return "", "", fmt.Errorf("Failed to fetch messages from the stream, %w", err)
	}

	message := stream[0].Messages[0]
	jobID, ok := message.Values["job_id"].(string)
	if !ok {
		// Remove this job from the consumer group.
		_ = w.redisClient.XAck(ctx, w.config.streamKey, w.config.groupName, message.ID).Err()
		return "", "", fmt.Errorf("Failed to convert the job id to a string in message %s", message.ID)
	}

	return jobID, message.ID, nil
}

func (w *worker) query(ctx context.Context, jobIDStr string) (*repo.WorkerFindJobByIDRow, error) {
	var jobID pgtype.UUID
	if err := jobID.Scan(jobIDStr); err != nil {
		return &repo.WorkerFindJobByIDRow{}, fmt.Errorf("Invalid job id, %w", err)
	}

	job, err := w.repo.WorkerFindJobByID(ctx, jobID)
	if err != nil {
		return &repo.WorkerFindJobByIDRow{}, fmt.Errorf("Error fetching job from the server, %w", err)
	}

	return &job, nil
}

// Execute the code in a job.
func (w *worker) execute(ctx context.Context, job *repo.WorkerFindJobByIDRow, jobIDStr string, messageID string) error {
	var jobID pgtype.UUID
	if err := jobID.Scan(jobIDStr); err != nil {
		return fmt.Errorf("Invalid job id, %w", err)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Set up Docker container.
	scriptCode := decodeFunction(job.Function)

	// Temporary file containing the job's code to prevent escape characters from bugging the program.
	tmpFile, err := os.CreateTemp("", "job-*.py")
	if err != nil {
		return fmt.Errorf("Failed to create temporary fiule, %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scriptCode); err != nil {
		return fmt.Errorf("Failed to write the scirpt, %w", err)
	}
	tmpFile.Close()

	config := DockerConfig()
	args, err := config.BuildDockerArgs(Language(job.Language), tmpFile.Name(), job.Dependencies)
	if err != nil {

		//  Update redis and jobs table to indicate that the job didn't run.
		_, _ = w.repo.WorkerUpdateStatus(ctx, repo.WorkerUpdateStatusParams{
			Status: repo.JobStatusFailed,
			ID:     jobID,
		})

		_ = w.redisClient.XAck(ctx, w.config.streamKey, w.config.groupName, messageID).Err()
		return fmt.Errorf("Failed to build Docker args, %w", err)
	}

	// Run the container.
	cmd := exec.CommandContext(execCtx, "docker", args...)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	finalStatus := repo.JobStatusCompleted

	if err != nil {
		finalStatus = repo.JobStatusFailed

		// Execution timed out.
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			outputStr += "\nExecution timed out."
		} else {
			slog.Warn("Job executed with an error", "job_id", jobIDStr, "error", err, "output", outputStr)
		}
	}

	// Create a new entry in the job results table.
	_, err = w.repo.WorkerCreateJobResult(ctx, repo.WorkerCreateJobResultParams{
		JobID:  jobID,
		Output: outputStr,
	})
	if err != nil {
		return fmt.Errorf("Failed to update job results, %w", err)
	}

	// Update jobs table status.
	_, err = w.repo.WorkerUpdateStatus(ctx, repo.WorkerUpdateStatusParams{
		Status: finalStatus,
		ID:     jobID,
	})
	if err != nil {
		return fmt.Errorf("Failed to update jobs, %w", err)
	}

	// Remove the job from the PEL.
	if err := w.redisClient.XAck(ctx, w.config.streamKey, w.config.groupName, messageID).Err(); err != nil {
		return fmt.Errorf("Failed to acknowledge the message, %w", err)
	}

	return nil
}

// Checks for abandoned jobs in the redis PEL.
func (w *worker) StartRecoveryLoop(ctx context.Context, workerName string) {
	ticker := time.NewTicker(w.config.claimInterval)
	defer ticker.Stop()

	startID := "0-0"

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping the recovery loop.")
			return
		case <-ticker.C:
			var err error
			startID, err = w.claimAbandonedJobs(ctx, workerName, startID)
			if err != nil {
				slog.Error("Failed to claim abandoned jobs", "error", err)
			}
		}
	}
}

// Gathers and runs abandoned jobs.
func (w *worker) claimAbandonedJobs(ctx context.Context, workerName string, startID string) (string, error) {
	messages, nextStartID, err := w.redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   w.config.streamKey,
		Group:    w.config.groupName,
		Consumer: workerName,
		MinIdle:  w.config.minIdleTime,
		Start:    startID,
		Count:    10,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return startID, fmt.Errorf("Falied to auto claim jobs, %w", err)
	}

	if len(messages) == 0 {
		return nextStartID, nil
	}

	// Process each message.
	for _, msg := range messages {
		jobIDStr, ok := msg.Values["job_id"].(string)
		if !ok {
			_ = w.redisClient.XAck(ctx, w.config.streamKey, w.config.groupName, msg.ID).Err()
			continue
		}

		// Fetch job details from database.
		job, err := w.query(ctx, jobIDStr)
		if err != nil {
			slog.Error("Failed to fetch jobs from the server", "error", err)
			continue
		}

		// Execute the job.
		if err := w.execute(ctx, job, jobIDStr, msg.ID); err != nil {
			slog.Error("Failed to execute the job", "job_id", jobIDStr, "error", err)
		}
	}

	return nextStartID, nil
}

// Decodes job code from plain text to base64.
func decodeFunction(rawFunction string) string {
	decodeBytes, err := base64.StdEncoding.DecodeString(rawFunction)
	if err != nil {
		return rawFunction
	}
	return string(decodeBytes)
}
