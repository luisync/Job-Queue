package users

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"github.com/luisync/Job-Queue/internal/json"
	"github.com/luisync/Job-Queue/internal/token"
	"github.com/luisync/Job-Queue/internal/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handlers depend on the services.
type handler struct {
	client     pb.JobqueueClient
	TokenMaker token.TokenMaker
}

// Constructor for creating the handlers.
func NewHandlerWithKey(client pb.JobqueueClient, secretKey string) *handler {
	return &handler{
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
	}
}

func NewHandlerWithTokenMaker(client pb.JobqueueClient, tokenMaker token.TokenMaker) *handler {
	return &handler{
		client:     client,
		TokenMaker: tokenMaker,
	}
}

// Register a user.
func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	// Get the email and password entered.
	var newUser createUserParams
	if err := json.Read(r, &newUser); err != nil {
		log.Println(err)
		http.Error(w, "Please include the correct fields.", http.StatusBadRequest)
		return
	}

	// Validade input.
	if len(newUser.Email) == 0 || len(newUser.Username) == 0 || len(newUser.Password) == 0 || len(newUser.First_name) == 0 || len(newUser.Last_name) == 0 {
		log.Printf("Empty credentials.")
		http.Error(w, "Please fill in all of the required fields.", http.StatusBadRequest)
		return
	}

	if len([]rune(newUser.Username)) < 8 || len([]rune(newUser.Password)) < 8 {
		log.Printf("Credentials are too short.")
		http.Error(w, "Please include a valid username and password.", http.StatusBadRequest)
		return
	}

	if len([]rune(newUser.First_name)) < 2 || len([]rune(newUser.First_name)) > 50 ||
		len([]rune(newUser.Last_name)) < 2 || len([]rune(newUser.Last_name)) > 50 {
		log.Printf("Name is out of the expected 2-50 character range.")
		http.Error(w, "Please include a valid first and last name.", http.StatusBadRequest)
		return
	}

	address, err := mail.ParseAddress(newUser.Email)
	if err != nil || address.Address != newUser.Email {
		log.Printf("Include a valid email.")
		http.Error(w, "Please include a valid email.", http.StatusBadRequest)
		return
	}

	// Hash password.
	hashedPassword, err := util.HashPassword(newUser.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}
	newUser.Password = hashedPassword

	// Create user.
	createdUser, err := h.client.Register(r.Context(), &pb.UsersReq{
		FirstName: newUser.First_name,
		LastName:  newUser.Last_name,
		Username:  newUser.Username,
		Email:     newUser.Email,
		Password:  newUser.Password,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, createdUser)
}

// Login as a user.
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var userDetails loginUserParams
	if err := json.Read(r, &userDetails); err != nil {
		log.Println(err)
		http.Error(w, "Please include the correct fields.", http.StatusBadRequest)
		return
	}

	// Validade input.
	if len(userDetails.Email) == 0 || len(userDetails.Password) == 0 {
		log.Printf("Empty credentials.")
		http.Error(w, "Please fill in all of the required fields.", http.StatusBadRequest)
		return
	}

	// Check whether the user is in the database.
	user, err := h.client.FindUserByEmail(r.Context(), &pb.UsersReq{
		Email: userDetails.Email,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Please include a valid email and password.", http.StatusBadRequest)
		return
	}

	// Check whether the the password is in the database.
	if err := util.CheckPassword(userDetails.Password, user.Password); err != nil {
		log.Printf("Wrong password.")
		http.Error(w, "Please include a valid email and password.", http.StatusBadRequest)
		return
	}

	// Create access token.
	var userID pgtype.UUID
	if err := userID.Scan(user.Id); err != nil {
		log.Print("Error converting id into uuid", err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(userID, user.Email, 15*time.Minute)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Create refresh token.
	refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(userID, user.Email, 24*time.Hour)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Create login session.
	session, err := h.client.CreateSession(r.Context(), &pb.SessionsReq{
		Id:           refreshClaims.RegisteredClaims.ID,
		UserEmail:    user.Email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    timestamppb.New(refreshClaims.RegisteredClaims.ExpiresAt.Time),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	res := LoginRes{
		Session_ID:               session.Id,
		Access_token:             accessToken,
		Refresh_token:            refreshToken,
		Access_token_expires_at:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		Refresh_token_expires_at: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		Email:                    user.Email,
	}

	json.Write(w, http.StatusOK, res)
}

// Logout a user.
func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user session id.
	claims, ok := r.Context().Value(AuthKey{}).(*token.UserClaims)
	if !ok {
		log.Println("Failed to read the context.")
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Delete session.
	deletedSession, err := h.client.DeteleSessions(r.Context(), &pb.SessionsReq{
		UserEmail: claims.Email,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, deletedSession)
}

// Renew access token.
func (h *handler) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	// Get the refresh token.
	var refreshToken RenewAccessTokenReq
	if err := json.Read(r, &refreshToken); err != nil {
		log.Println(err)
		http.Error(w, "Please include the correct fields.", http.StatusBadRequest)
		return
	}

	if len(refreshToken.Refresh_token) == 0 {
		log.Println("Empty field.")
		http.Error(w, "Please include a refresh token.", http.StatusBadRequest)
		return
	}

	// Varify token.
	refreshClaims, err := h.TokenMaker.VerfifyToken(refreshToken.Refresh_token)
	if err != nil {
		log.Println(err)
		http.Error(w, "Please include a valid refresh token.", http.StatusBadRequest)
		return
	}

	// Get the user's current session.
	session, err := h.client.FindSessionByID(r.Context(), &pb.SessionsReq{
		RefreshToken: refreshClaims.RegisteredClaims.ID,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Compare session email with the current one.
	if session.UserEmail != refreshClaims.Email {
		log.Println(
			"The session in the database doesn't have the same " +
				" email as the one the user is logged in on.")
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Check whether the session has been revoked.
	if session.IsRevoked {
		log.Println("This session has been revoked.")
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Create a new token.
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, 15*time.Minute)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	res := RenewAccessTokenRes{
		Access_token:            accessToken,
		Access_token_expires_at: accessClaims.RegisteredClaims.ExpiresAt.Time,
	}

	json.Write(w, http.StatusOK, res)
}

// Revoke a user's sessions.
func (h *handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	// Get the session id.
	claims, ok := r.Context().Value(AuthKey{}).(*token.UserClaims)
	if !ok {
		log.Println("Failed to read the context.")
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Revoke the session.
	revokedSession, err := h.client.RevokeSessions(r.Context(), &pb.SessionsReq{
		UserEmail: claims.Email,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, revokedSession)
}

// Get the id of the user that's currently logged in.
func GetUserIDFromContext(ctx context.Context) (pgtype.UUID, error) {
	// Get id from the context.
	claims, ok := ctx.Value(AuthKey{}).(*token.UserClaims)
	if !ok {
		return pgtype.UUID{}, fmt.Errorf("Failed to get the user id from the context.")
	}

	userID := claims.ID

	return userID, nil
}
