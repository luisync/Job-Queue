package users

import (
	"context"

	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
)

// Functions a struct must define to be considered a service.
type Service interface {
	FindUserByEmail(ctx context.Context, email string) (repo.User, error)
	Register(ctx context.Context, newUser createUserParams) (repo.User, error)
	FindSessionByID(ctx context.Context, id string) (repo.Session, error)
	FindSessionsByEmail(ctx context.Context, email string) ([]repo.Session, error)
	RevokeSessions(ctx context.Context, email string) ([]repo.Session, error)
	DeteleSessions(ctx context.Context, email string) ([]repo.Session, error)
	CreateSession(ctx context.Context, newSession createSessionParams) (repo.Session, error)
}

// Services depend on the repository.
type svc struct {
	repo repo.Querier
}

// Constructor for creating the services.
func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

// Find a user by email.
func (svc *svc) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return svc.repo.FindUserByEmail(ctx, email)
}

// Register a user.
func (svc *svc) Register(ctx context.Context, newUser createUserParams) (repo.User, error) {
	return svc.repo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: newUser.First_name,
		LastName:  newUser.Last_name,
		Username:  newUser.Username,
		Password:  newUser.Password,
		Email:     newUser.Email,
	})
}

// User sessions.
func (s *svc) FindSessionByID(ctx context.Context, id string) (repo.Session, error) {
	return s.repo.FindSessionByID(ctx, id)
}

func (s *svc) FindSessionsByEmail(ctx context.Context, email string) ([]repo.Session, error) {
	return s.repo.FindSessionsByEmail(ctx, email)
}

func (s *svc) RevokeSessions(ctx context.Context, email string) ([]repo.Session, error) {
	return s.repo.RevokeSessions(ctx, email)
}

func (s *svc) DeteleSessions(ctx context.Context, email string) ([]repo.Session, error) {
	return s.repo.DeteleSessions(ctx, email)
}

func (s *svc) CreateSession(ctx context.Context, newSession createSessionParams) (repo.Session, error) {
	return s.repo.CreateSession(ctx, repo.CreateSessionParams{
		ID:           newSession.ID,
		UserEmail:    newSession.User_email,
		RefreshToken: newSession.Refresh_token,
		IsRevoked:    newSession.Is_revoked,
		ExpiresAt:    newSession.Expires_at,
	})
}
