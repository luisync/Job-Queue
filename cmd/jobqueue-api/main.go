package main

import (
	"log/slog"
	"os"

	"github.com/luisync/Job-Queue/internal/env"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		grpcServerAddr = env.GetString("GRPC_SERVER_ADDR", "0.0.0.0:9091")
		httpServerAddr = env.GetString("HTTP_SERVER_ADDR", ":8080")
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create gRPC client.
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(grpcServerAddr, opts...)
	if err != nil {
		slog.Error("Server failed to connect to the grpc server", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	slog.Info("Connected to the grpc server", "addr", grpcServerAddr)

	client := pb.NewJobqueueClient(conn)

	// Start the HTTP server.
	api := application{
		addr: httpServerAddr,
		db:   client,
	}

	slog.Info("Starting the HTTP API gateway", "addr", httpServerAddr)
	if err := api.run(api.mount()); err != nil {
		// Terminate the program if an error occurs.
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
