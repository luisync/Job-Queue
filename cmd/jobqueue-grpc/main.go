package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luisync/Job-Queue/internal/env"
	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/server"
	"google.golang.org/grpc"
)

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}

// Start the gRPC server.
func main() {
	// Instantiate the server.
	ctx := context.Background()

	// Define config.
	cfg := config{
		addr: env.GetString("GRPC_SERVER_ADDR", "0.0.0.0:9091"),
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=postgres dbname=jobs-queue sslmode=disable"),
		},
	}

	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to the database.
	conn, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	slog.Info("Connected to database", "dsn", cfg.db.dsn)

	repo := repo.New(conn)
	srv := server.NewServer(repo)

	// Register the server with the grpc server.
	grpcSrv := grpc.NewServer()
	pb.RegisterJobqueueServer(grpcSrv, srv)

	listener, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}

	slog.Info("Server listening", "addr", cfg.addr)

	if err := grpcSrv.Serve(listener); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
