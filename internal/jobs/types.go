package jobs

import (
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/luisync/Job-Queue/internal/adapters/postgresql/sqlc"
)

// Definition of the creation of a new job entity.
type createJobParams struct {
	CreatorID    pgtype.UUID      `json:"creator_id"`
	Language     repo.JobLanguage `json:"language"`
	Dependencies string           `json:"dependencies"`
	Function     string           `json:"function"`
}
