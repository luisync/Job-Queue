package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"github.com/luisync/Job-Queue/internal/mocks"
	"github.com/luisync/Job-Queue/internal/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRegister(t *testing.T) {
	secretKey := "8/#?K}RAy$R|#($&1QLVZCWpK8*az_Jrh(EKUQv8rHq"

	tests := []struct {
		name           string
		body           createUserParams
		grpcMock       func(m *mocks.MockJobqueueClient)
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "Ok - Valid user",
			body: createUserParams{
				First_name: "John",
				Last_name:  "Robert",
				Username:   "johnrobert",
				Password:   "johnpassword",
				Email:      "jrobert@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&pb.UsersRes{
						Id:        "63b9337c-44ef-43ac-b21e-3db7d2cb07c1",
						FirstName: "John",
						LastName:  "Robert",
						Username:  "johnrobert",
						Password:  "johnpassword",
						Email:     "jrobert@gmail.com",
					}, nil).
					Times(1)

			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "Bad Request - Empty username",
			body: createUserParams{
				First_name: "Afonso",
				Last_name:  "Aiges",
				Username:   "",
				Password:   "aigespassword",
				Email:      "aigesemail@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				// gRPC isn't called when there is a validation error.
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Bad Request - Invalid last name",
			body: createUserParams{
				First_name: "Don",
				Last_name:  "R",
				Username:   "donaccount",
				Password:   "donpassword",
				Email:      "donRob@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Bad Request - Missing field",
			body: createUserParams{
				First_name: "Gomes",
				Username:   "gomesRaf",
				Password:   "gomespassword",
				Email:      "gomesrafa@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Bad Request - Invalid invalid email",
			body: createUserParams{
				First_name: "Ana",
				Last_name:  "Lopez",
				Username:   "analopez",
				Password:   "anapassword",
				Email:      "invalid mail",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Internal Server Error - Data mismatch",
			body: createUserParams{
				First_name: "Alex",
				Last_name:  "Luigi",
				Username:   "alexluigi",
				Password:   "alexpassword",
				Email:      "gomesrafa@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(&pb.UsersRes{}, errors.New("Error creating the user"))
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start go mock.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockClient := mocks.NewMockJobqueueClient(mockCtrl)
			tt.grpcMock(mockClient)

			h := NewHandlerWithKey(mockClient, secretKey)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "account/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()

			h.Register(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
			if tt.wantErr {
				assert.NotNil(t, res.Body)
				return
			}

			var jsonRes pb.UsersRes
			err = json.Unmarshal(res.Body.Bytes(), &jsonRes)
			require.NoError(t, err)

			assert.NotEmpty(t, jsonRes.Id)
			assert.NotEmpty(t, jsonRes.FirstName)
			assert.NotEmpty(t, jsonRes.LastName)
			assert.NotEmpty(t, jsonRes.Username)
			assert.NotEmpty(t, jsonRes.Email)
			assert.NotEmpty(t, jsonRes.Password)
			assert.NotEmpty(t, jsonRes.UpdatedAt)
			assert.NotEmpty(t, jsonRes.CreatedAt)
		})
	}
}

func TestLogin(t *testing.T) {
	secretKey := "HL6WXp*L.=^UghKIK},3Hcu)tpAW%&0dncfa)r46B|a"

	tests := []struct {
		name           string
		body           loginUserParams
		grpcMock       func(m *mocks.MockJobqueueClient)
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "Ok - Valid existing user",
			body: loginUserParams{
				Password: "williampass",
				Email:    "williamemail@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					FindUserByEmail(gomock.Any(), gomock.Any()).
					Return(&pb.UsersRes{
						Id:        "b3efd1ca-35e4-43f1-991c-87f3422dec5b",
						FirstName: "william",
						LastName:  "Flores",
						Username:  "williamofficial",
						Email:     "williamemail@gmail.com",
						Password:  "$2a$10$m0jMQv56c9JC8CsmMytH6eyRKnYhMPoCPi1j4oqvLqYt9S1O6R0gK",
						UpdatedAt: timestamppb.Now(),
						CreatedAt: timestamppb.Now(),
					}, nil).
					Times(1)
				m.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{
						Id:           "e9abe675-20be-4609-9886-1bfb98b4516d",
						UserEmail:    "williamemail@gmail.com",
						RefreshToken: "refresh token",
						IsRevoked:    false,
						CreatedAt:    timestamppb.Now(),
						ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "Bad Request - Missing fields",
			body: loginUserParams{
				Password: "williampass",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					FindUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				m.EXPECT().CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Bad Request - Invalid password",
			body: loginUserParams{
				Password: "not william's password",
				Email:    "williamemail@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					FindUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				m.EXPECT().CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Server Error - Error creating the session",
			body: loginUserParams{
				Password: "not william's password",
				Email:    "williamemail@gmail.com",
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					FindUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)
				m.EXPECT().CreateSession(gomock.Any(), gomock.Any()).
					Return(&pb.UsersRes{}, errors.New("Error creating session")).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start go mock.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockClient := mocks.NewMockJobqueueClient(mockCtrl)
			tt.grpcMock(mockClient)

			h := NewHandlerWithKey(mockClient, secretKey)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "account/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()

			h.Register(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
			if tt.wantErr {
				assert.Contains(t, res.Body.String(), "Server error, please try again later.")
				return
			}

			var jsonRes LoginRes
			err = json.Unmarshal(res.Body.Bytes(), &jsonRes)
			require.NoError(t, err)

			assert.NotEmpty(t, jsonRes.Session_ID)
			assert.NotEmpty(t, jsonRes.Access_token)
			assert.NotEmpty(t, jsonRes.Refresh_token)
			assert.NotEmpty(t, jsonRes.Access_token_expires_at)
			assert.NotEmpty(t, jsonRes.Refresh_token_expires_at)
			assert.NotEmpty(t, jsonRes.Email)
		})
	}
}

func TestLogout(t *testing.T) {
	secretKey := "Eh>Du1hi<B8.1(V/Uv/y:_;2||m6U8-{d>?sreFZ*1G"

	tests := []struct {
		name           string
		expectedClaims func() context.Context
		grpcMock       func(m *mocks.MockJobqueueClient)
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "Ok - Valid claim",
			expectedClaims: func() context.Context {
				var userID pgtype.UUID
				err := userID.Scan("b3efd1ca-35e4-43f1-991c-87f3422dec5b")
				require.NoError(t, err)

				claims, err := token.NewUserClaims(userID, "williamemail@gmail.com", 24*time.Hour)
				require.NoError(t, err)

				return context.WithValue(context.Background(), AuthKey{}, claims)
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					DeteleSessions(gomock.Any(), gomock.Any()).
					Return(&pb.ListSessionsRes{
						Sessions: []*pb.SessionsRes{
							&pb.SessionsRes{
								Id:           "17d84354-5f2d-48a1-9cbb-ae1e87ec1500",
								UserEmail:    "williamemail@gmail.com",
								RefreshToken: "refresh token",
								IsRevoked:    false,
								CreatedAt:    timestamppb.Now(),
								ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
							},
							&pb.SessionsRes{
								Id:           "953fcd83-6d37-4785-aeff-e94c44b3050b",
								UserEmail:    "williamemail@gmail.com",
								RefreshToken: "refresh token",
								IsRevoked:    true,
								CreatedAt:    timestamppb.Now(),
								ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
							},
						},
					}, nil).
					Times(1)
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "Server Error - Incorrect context",
			expectedClaims: func() context.Context {
				return context.WithValue(context.Background(), AuthKey{}, nil)
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					DeteleSessions(gomock.Any(), gomock.Any()).
					Return(&pb.ListSessionsRes{}, nil).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name: "Server Error - Invalid token",
			expectedClaims: func() context.Context {
				return context.WithValue(context.Background(), AuthKey{}, "")
			},
			grpcMock: func(m *mocks.MockJobqueueClient) {
				m.EXPECT().
					DeteleSessions(gomock.Any(), gomock.Any()).
					Return(&pb.ListSessionsRes{}, errors.New("Error finding token"))
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockClient := mocks.NewMockJobqueueClient(ctrl)
			handler := NewHandlerWithKey(mockClient, secretKey)

			request := httptest.NewRequestWithContext(tt.expectedClaims(), http.MethodPost, "account/logout", nil)
			response := httptest.NewRecorder()

			handler.Logout(response, request)

			assert.Equal(t, tt.expectedStatus, response.Code)
			if tt.wantErr {
				assert.Contains(t, response.Body.String(), "Server error, please try again later.")
				return
			}

			var jsonRes *pb.ListSessionsRes
			err := json.Unmarshal(response.Body.Bytes(), &jsonRes)

			require.NoError(t, err)
			for _, session := range jsonRes.Sessions {
				assert.NotEmpty(t, session.Id)
				assert.NotEmpty(t, session.UserEmail)
				assert.NotEmpty(t, session.RefreshToken)
				assert.NotEmpty(t, session.IsRevoked)
				assert.NotEmpty(t, session.CreatedAt)
				assert.NotEmpty(t, session.ExpiresAt)
			}
		})
	}
}

func TestRewnewAccessToken(t *testing.T) {
	// Create a valid refresh token to pass as the payload.
	var userID pgtype.UUID
	err := userID.Scan("fa90c969-2da0-4bd2-85fb-8dd589e1d017")
	require.NoError(t, err)

	secretKey := "xH^hmlsIR>=2Yt+016RT!!cXVz%8W4jr!D:R+?*qEZ{"
	tokenMaker := token.NewJWTMaker(secretKey)
	refreshToken, actualClaims, err := tokenMaker.CreateToken(userID, "joshemail@gmail.com", 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name             string
		body             RenewAccessTokenReq
		dependenciesMock func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker)
		expectedStatus   int
		wantErr          bool
	}{
		{
			name: "Ok - Valid token",
			body: RenewAccessTokenReq{
				Refresh_token: refreshToken,
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{
						Id:           userID.String(),
						UserEmail:    "joshemail@gmail.com",
						RefreshToken: refreshToken,
						IsRevoked:    false,
						CreatedAt:    timestamppb.Now(),
						ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
					}, nil).
					Times(1)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(actualClaims, nil).
					Times(1)
				t.EXPECT().
					CreateToken(userID, "joshemail@gmail.com", 15*time.Minute).
					Times(1)

			},
			expectedStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name: "Bad Request - Missing refresh token",
			body: RenewAccessTokenReq{},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Times(0)

				t.EXPECT().
					VerfifyToken(gomock.Any()).
					Times(0)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Bad Request - Empty refresh token",
			body: RenewAccessTokenReq{
				Refresh_token: "",
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Times(0)

				t.EXPECT().
					VerfifyToken(gomock.Any()).
					Times(0)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "Server Error - Server couldn't verify the token",
			body: RenewAccessTokenReq{
				Refresh_token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30",
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Times(0)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(&token.UserClaims{}, errors.New("Error verfiying token")).
					Times(1)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name: "Server Error - Error finding session",
			body: RenewAccessTokenReq{
				Refresh_token: refreshToken,
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{}, errors.New("Error fetching session")).
					Times(1)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(actualClaims, nil).
					Times(1)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name: "Server Error - Refresh token and session email mismatch",
			body: RenewAccessTokenReq{
				Refresh_token: refreshToken,
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{
						Id:           userID.String(),
						UserEmail:    "mismatch@gmail.com",
						RefreshToken: refreshToken,
						IsRevoked:    false,
						CreatedAt:    timestamppb.Now(),
						ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
					}, nil).
					Times(1)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(actualClaims, nil).
					Times(1)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name: "Server Error - Session revoked",
			body: RenewAccessTokenReq{
				Refresh_token: refreshToken,
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{
						Id:           userID.String(),
						UserEmail:    "joshemail@gmail.com",
						RefreshToken: refreshToken,
						IsRevoked:    true,
						CreatedAt:    timestamppb.Now(),
						ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
					}, nil)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(actualClaims, nil).
					Times(1)
				t.EXPECT().
					CreateToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name: "Server Error - Error creating a new token",
			body: RenewAccessTokenReq{
				Refresh_token: refreshToken,
			},
			dependenciesMock: func(m *mocks.MockJobqueueClient, t *mocks.MockTokenMaker) {
				m.EXPECT().
					FindSessionByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&pb.SessionsRes{
						Id:           userID.String(),
						UserEmail:    "joshemail@gmail.com",
						RefreshToken: refreshToken,
						IsRevoked:    true,
						CreatedAt:    timestamppb.Now(),
						ExpiresAt:    timestamppb.New(time.Now().Add(24 * time.Hour)),
					}, nil).
					Times(1)

				t.EXPECT().
					VerfifyToken(secretKey).
					Return(actualClaims, nil).
					Times(1)
				t.EXPECT().
					CreateToken(userID, "joshemail@gmail.com", 15*time.Minute).
					Return("", &token.UserClaims{}, errors.New("Error creating token")).
					Times(1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientController := gomock.NewController(t)
			tokenMakerController := gomock.NewController(t)

			clientMock := mocks.NewMockJobqueueClient(clientController)
			tokenMakerMock := mocks.NewMockTokenMaker(tokenMakerController)

			test.dependenciesMock(clientMock, tokenMakerMock)

			handler := NewHandlerWithTokenMaker(clientMock, tokenMakerMock)

			bodyBytes, err := json.Marshal(test.body)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "account/renew", bytes.NewBuffer(bodyBytes))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.RenewAccessToken(response, request)
			assert.Equal(t, test.expectedStatus, response.Code)
			if test.wantErr {
				assert.NotNil(t, response.Body)
				return
			}

			var responseJSON RenewAccessTokenRes
			err = json.Unmarshal(response.Body.Bytes(), &responseJSON)
			require.NoError(t, err)

			assert.NotEmpty(t, responseJSON.Access_token)
			assert.NotEmpty(t, responseJSON.Access_token_expires_at)
		})
	}

}
