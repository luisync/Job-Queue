package jobs

import (
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/luisync/Job-Queue/internal/adapters/postgresql/sqlc"
)

type createJobReq struct {
	Language     repo.JobLanguage `json:"language"`
	Dependencies string           `json:"dependencies"`
	Function     string           `json:"function"`
}

type createJobParams struct {
	CreatorID pgtype.UUID `json:"creator_id"`
	createJobReq
}

type findJobByIDParams struct {
	ID         pgtype.UUID `json:"id"`
	Creator_id pgtype.UUID `json:"creator_id"`
}
