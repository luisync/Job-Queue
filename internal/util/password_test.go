package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password.",
			password: "y1^MiP'5]n",
			wantErr:  false,
		},
		{
			name:     "Empty password.",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Password exceeding 72 bytes.",
			password: string(make([]byte, 73)),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPass, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
				return
			}

			require.NoError(t, err)

			// Verify that the correct hash was created from the input.
			err = bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(tt.password))
			assert.NoError(t, err)
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		hashedPassword string
		wantErr        bool
	}{
		{
			name:           "Valid password and hash.",
			password:       "Zz2hLOOuUExDZtVi",
			hashedPassword: "$2a$10$0iZlP..8vRjBu1Id.Gn6uO/ZxKzh.tAFZMm13/UJVilDHKgErzHCG",
			wantErr:        false,
		},
		{
			name:           "Empty password and corresponding hash.",
			password:       "",
			hashedPassword: "$2y$10$dUfI/PMTJBW.icryYg4P4eZDkhdVK/Qow5dsEJpleF2XZHzhgH1N.",
			wantErr:        false,
		},
		{
			name:           "Empty hash.",
			password:       "GgDHBcfQcCKGmEmI",
			hashedPassword: "",
			wantErr:        true,
		},
		{
			name:           "Incorrect hash.",
			password:       "6LXz6eAtYaxUgVFl",
			hashedPassword: "$2y$10$kQCJkP.Ej6a93OzFr0od8eHOYzwgVEKq7LEQ.KkbXv0jWq3Lihzbu",
			wantErr:        true,
		},
		{
			name:           "Invalid hash.",
			password:       "OKERoU0ijsO8SEtY",
			hashedPassword: "invalid hash",
			wantErr:        true,
		},
		{
			name:           "Empty hash and password.",
			password:       "",
			hashedPassword: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acErr := CheckPassword(tt.password, tt.hashedPassword)
			if tt.wantErr {
				assert.Error(t, acErr)
				return
			}

			assert.NoError(t, acErr)
		})
	}
}
