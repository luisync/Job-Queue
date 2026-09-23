package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthMiddlewareFunc(t *testing.T) {
	// Create a token maker and generate a token.
	tokenMaker := token.NewJWTMaker("1933^{NJjI[*U2X)LNGeNN?{F7bRY(*wHO&c2kYg]Yc")

	var userID pgtype.UUID
	err := userID.Scan("b3efd1ca-35e4-43f1-991c-87f3422dec5b")
	require.NoError(t, err)

	token, acClaims, err := tokenMaker.CreateToken(userID, "williamemail@gmail.com", 15*time.Minute)
	require.NoError(t, err)

	// Get the authentication function.
	authFunc := GetAuthMiddlewareFunc(tokenMaker)

	tests := []struct {
		name       string
		body       func(r *http.Request)
		exptStatus int
		exptCall   bool
		wantErr    bool
	}{
		{
			name: "OK - Valid token",
			body: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+token)
			},
			exptStatus: http.StatusOK,
			exptCall:   true,
			wantErr:    false,
		},
		{
			name: "Unauthorized - Invalid token",
			body: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+"12345")
			},
			exptStatus: http.StatusUnauthorized,
			exptCall:   false,
			wantErr:    true,
		},
		{
			name:       "Unauthorized - Missing token",
			body:       func(r *http.Request) {},
			exptStatus: http.StatusUnauthorized,
			exptCall:   false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCall := false

			// Create the next handler the middlware will call.
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualCall = true
				claims := r.Context().Value(AuthKey{})

				// Verify that the middleware stored the correct values.
				assert.Equal(t, acClaims, claims)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("User authenticated"))
			})

			// Create the middleware.
			authHandler := authFunc(nextHandler)

			req := httptest.NewRequest(http.MethodPost, "/account/login", nil)
			res := httptest.NewRecorder()
			tt.body(req)

			authHandler.ServeHTTP(res, req)

			assert.Equal(t, tt.exptStatus, res.Code)
			assert.Equal(t, tt.exptCall, actualCall)

			if tt.wantErr {
				assert.Contains(t, res.Body.String(), "Error verifying token.")
			}
		})
	}
}
