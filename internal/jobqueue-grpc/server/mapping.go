package server

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// Helper function that attempts to convert strings into UUIDs.
func toUUID(uuidStr string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(uuidStr); err != nil {
		return pgtype.UUID{}, fmt.Errorf("Invalid id, %w", err)
	}

	return uuid, nil
}
