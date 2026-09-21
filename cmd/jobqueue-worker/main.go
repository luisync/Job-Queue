package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luisync/Job-Queue/internal/env"
	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
	"github.com/luisync/Job-Queue/internal/poller"
	"github.com/luisync/Job-Queue/internal/worker"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Cancel context on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		addr      = env.GetString("REDIS_SERVER_ADDR", "localhost:6379")
		password  = env.GetString("REDIS_PASSWORD", "")
		dsn       = env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=jobs-queue sslmode=disable")
		streamKey = env.GetString("STREAM_KEY", "jobs:stream")
		groupName = env.GetString("GROUP_NAME", "worker-group")
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to reids.
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	defer client.Close()

	// Connect to the repository.
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	slog.Info("Connected to database", "dsn", dsn)

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Worker failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Redis", "addr", addr)

	// Create consumer group on Redis.
	if err := worker.CreateConsumerGroup(ctx, client, streamKey, groupName); err != nil {
		slog.Error("Failed to create consumer group", "error", err)
		os.Exit(1)
	}

	// Instantiate the worker.
	workerConfig := worker.NewWorkerConfig(groupName, streamKey, 10*time.Minute, 1*time.Minute)
	repo := repo.New(conn)
	worker, err := worker.NewWorker(workerConfig, client, repo)
	if err != nil {
		slog.Error("Failed to initiate worker", "error", err)
		os.Exit(1)
	}

	// Instatiate pending job scheduled poller.
	poller := poller.NewPoller(streamKey, 50, client, repo)

	var wg sync.WaitGroup

	// Start the worker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := worker.Run(ctx); err != nil {
			slog.Error("Worker exited with error", "error", err)
		}
	}()

	// Start the poller.
	wg.Add(1)
	go func() {
		defer wg.Done()
		poller.StartPoller(ctx, 5*time.Second)
	}()

	slog.Info("Worker and poller running...")

	// Graceful shutdown.
	<-ctx.Done()
	slog.Info("Shutdown signal received. Cleaning up...")

	wg.Wait()
	slog.Info("Worker and poller stopped.")
}
