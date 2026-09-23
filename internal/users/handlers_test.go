package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	mock_pb "github.com/luisync/Job-Queue/internal/mocks"
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
		grpcMock       func(m *mock_pb.MockJobqueueClient)
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			expectedStatus: http.StatusAccepted,
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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

			mockClient := mock_pb.NewMockJobqueueClient(mockCtrl)
			tt.grpcMock(mockClient)

			h := NewHandler(mockClient, secretKey)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "account/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()

			h.Register(res, req)

			assert.Equal(t, tt.expectedStatus, res.Code)
			if tt.wantErr {
				assert.Contains(t, res.Body.String(), "Server error, please try again later.")
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
		grpcMock       func(m *mock_pb.MockJobqueueClient)
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "Ok - Valid existing user",
			body: loginUserParams{
				Password: "williampass",
				Email:    "williamemail@gmail.com",
			},
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			expectedStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name: "Bad Request - Missing fields",
			body: loginUserParams{
				Password: "williampass",
			},
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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
			grpcMock: func(m *mock_pb.MockJobqueueClient) {
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

			mockClient := mock_pb.NewMockJobqueueClient(mockCtrl)
			tt.grpcMock(mockClient)

			h := NewHandler(mockClient, secretKey)

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
