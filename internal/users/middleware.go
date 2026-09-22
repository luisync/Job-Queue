package users

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/luisync/Job-Queue/internal/token"
)

type AuthKey struct{}

// Allow for a token maker to be passed whilist maintaining the correct signature of a middleware function.
func GetAuthMiddlewareFunc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	// Return the middleware function.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the token and get it's claims.
			claims, err := verifyClaimsFromAuthHeader(r, tokenMaker)
			if err != nil {
				http.Error(w, "Error verifying token.", http.StatusUnauthorized)
				return
			}

			// Store the claims in the context.
			ctx := context.WithValue(r.Context(), AuthKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Verify that the token was signed by the server and return its claims.
func verifyClaimsFromAuthHeader(r *http.Request, tokenMaker *token.JWTMaker) (*token.UserClaims, error) {
	// Get the the token from the authorisation header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("Authorisation header is missing.")
	}

	// Check that the there are two fields "Bearer [token]"
	fields := strings.Fields(authHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, fmt.Errorf("Invalid authorisation header.")
	}

	// Verfiy the token.
	token := fields[1]
	claims, err := tokenMaker.VerfifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("Invalid token: %w", err)
	}

	// Return its claims.
	return claims, nil
}
