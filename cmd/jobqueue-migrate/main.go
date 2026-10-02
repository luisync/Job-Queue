package main

import (
	"database/sql"
	"log"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/luisync/Job-Queue/internal/env"
	"github.com/pressly/goose/v3"
)

const (
	migrationsDir = "/migrations"
	maxRetries    = 30
	retryDelay    = 2 * time.Second
)

func main() {
	dbString := env.GetString(
		"GOOSE_DBSTRING",
		"host=127.0.0.1 user=postgres password=postgres dbname=jobs-queue sslmode=disable",
	)

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = sql.Open("pgx", dbString)
		if err != nil {
			log.Printf(
				"failed to open database connection (attempt %d/%d): %v",
				attempt,
				maxRetries,
				err,
			)

			time.Sleep(retryDelay)
			continue
		}

		err = db.Ping()
		if err == nil {
			break
		}

		log.Printf(
			"database not ready (attempt %d/%d): %v",
			attempt,
			maxRetries,
			err,
		)

		db.Close()
		db = nil

		time.Sleep(retryDelay)
	}

	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("Migration failed to set dialect", "error", err)
	}

	slog.Info("Starting database migration")

	if err := goose.Up(db, migrationsDir); err != nil {
		slog.Error("Migration failed to set migration directory", "error", err)
	}

	slog.Info("Database migration completed successfully")
}
