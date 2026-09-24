package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TokenMaker interface {
	CreateToken(id pgtype.UUID, email string, duration time.Duration) (string, *UserClaims, error)
	VerfifyToken(tokenStr string) (*UserClaims, error)
}

// Generator for creating JWTs.
type JWTMaker struct {
	secretKey string
}

// Create a new JWT generator.
func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{secretKey}
}

// Generate a new JWT with the generator.
func (maker *JWTMaker) CreateToken(id pgtype.UUID, email string, duration time.Duration) (string, *UserClaims, error) {
	// Create user login claims.
	claims, err := NewUserClaims(id, email, duration)
	if err != nil {
		return "", nil, err
	}

	// Create the token with the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token.
	tokenStr, err := token.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, err
	}

	return tokenStr, claims, nil
}

// Verify whether a token was signed by this maker.
func (maker *JWTMaker) VerfifyToken(tokenStr string) (*UserClaims, error) {
	// Get the information inside of the token.
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method.
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Invalid token signing method.")
		}

		return []byte(maker.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Check the token's claims.
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("Invalid token claims.")
	}

	return claims, nil
}
