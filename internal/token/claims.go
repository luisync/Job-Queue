package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Custom JWT claim for logging into an account.
type UserClaims struct {
	ID    pgtype.UUID `json:"id"`
	Email string      `json:"email"`
	jwt.RegisteredClaims
}

// Instantiate a new set of login claims.
func NewUserClaims(id pgtype.UUID, email string, duration time.Duration) (*UserClaims, error) {
	// Generate a token ID.
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &UserClaims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}, nil
}
