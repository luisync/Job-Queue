package server

import (
	"context"
	"fmt"

	repo "github.com/luisync/Job-Queue/internal/jobqueue-grpc/adapters/postgresql/sqlc"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// gRPC server.
type Server struct {
	repo repo.Querier
	pb.UnimplementedJobqueueServer
}

func NewServer(repo repo.Querier) *Server {
	return &Server{
		repo: repo,
	}
}

// Jobs server.
func (s *Server) ListJobs(ctx context.Context, req *pb.JobsReq) (*pb.ListJobsRes, error) {
	creatorID, err := toUUID(req.GetCreatorId())
	if err != nil {
		return &pb.ListJobsRes{}, fmt.Errorf("Invalid creator id, %w", err)
	}

	jobs, err := s.repo.ListJobs(ctx, creatorID)
	if err != nil {
		return &pb.ListJobsRes{}, fmt.Errorf("Error creating the job, %w", err)
	}

	// Iterate over the jobs and convert them into a job response.
	jobsRes := make([]*pb.JobsRes, 0, len(jobs))
	for _, job := range jobs {
		jobsRes = append(jobsRes, &pb.JobsRes{
			Id:           job.ID.String(),
			CreatorId:    job.CreatorID.String(),
			Language:     string(job.Language),
			Dependencies: job.Dependencies,
			Function:     job.Function,
			Status:       string(job.Status),
			UpdatedAt:    timestamppb.New(job.UpdatedAt.Time),
			CreatedAt:    timestamppb.New(job.CreatedAt.Time),
		})
	}

	return &pb.ListJobsRes{
		Jobs: jobsRes,
	}, nil
}

func (s *Server) FindJob(ctx context.Context, req *pb.FindJobReq) (*pb.JobsRes, error) {
	creatorID, err := toUUID(req.CreatorId)
	if err != nil {
		return &pb.JobsRes{}, fmt.Errorf("Invalid creator id, %w", err)
	}

	userID, err := toUUID(req.Id)
	if err != nil {
		return &pb.JobsRes{}, fmt.Errorf("Invalid user id, %w", err)
	}

	job, err := s.repo.FindJobByID(ctx, repo.FindJobByIDParams{
		ID:        userID,
		CreatorID: creatorID,
	})
	if err != nil {
		return &pb.JobsRes{}, fmt.Errorf("Error creating the job, %w", err)
	}

	return &pb.JobsRes{
		Id:           job.ID.String(),
		CreatorId:    job.CreatorID.String(),
		Language:     string(job.Language),
		Dependencies: job.Dependencies,
		Function:     job.Function,
		Status:       string(job.Status),
		UpdatedAt:    timestamppb.New(job.UpdatedAt.Time),
		CreatedAt:    timestamppb.New(job.CreatedAt.Time),
	}, nil
}

func (s *Server) CreateJob(ctx context.Context, req *pb.JobsReq) (*pb.JobsRes, error) {
	creatorID, err := toUUID(req.GetCreatorId())
	if err != nil {
		return &pb.JobsRes{}, fmt.Errorf("Invalid creator id, %w", err)
	}

	createdJob, err := s.repo.CreateJob(ctx, repo.CreateJobParams{
		CreatorID:    creatorID,
		Language:     repo.JobLanguage(req.Job.Language),
		Dependencies: req.Job.Dependencies,
		Function:     req.Job.Function,
	})
	if err != nil {
		return &pb.JobsRes{}, fmt.Errorf("Error creating the job, %w", err)
	}

	return &pb.JobsRes{
		Id:           createdJob.ID.String(),
		CreatorId:    createdJob.CreatorID.String(),
		Language:     string(createdJob.Language),
		Dependencies: createdJob.Dependencies,
		Function:     createdJob.Function,
		Status:       string(createdJob.Status),
		UpdatedAt:    timestamppb.New(createdJob.UpdatedAt.Time),
		CreatedAt:    timestamppb.New(createdJob.CreatedAt.Time),
	}, nil
}

// Job results server.
func (s *Server) ListJobResults(ctx context.Context, req *pb.JobResultsReq) (*pb.ListJobResultsRes, error) {
	creatorID, err := toUUID(req.CreatorId)
	if err != nil {
		return &pb.ListJobResultsRes{}, fmt.Errorf("Invalid creator id, %w", err)
	}

	jobID, err := toUUID(req.JobId)
	if err != nil {
		return &pb.ListJobResultsRes{}, fmt.Errorf("Invalid job id, %w", err)
	}

	jobs, err := s.repo.ListJobResults(ctx, repo.ListJobResultsParams{
		CreatorID: creatorID,
		JobID:     jobID,
	})
	if err != nil {
		return &pb.ListJobResultsRes{}, fmt.Errorf("Error creating the job, %w", err)
	}

	jobsRes := make([]*pb.JobResultsRes, 0, len(jobs))
	for _, job := range jobs {
		jobsRes = append(jobsRes, &pb.JobResultsRes{
			Output:    job.Output,
			CreatedAt: timestamppb.New(job.CreatedAt.Time),
		})
	}

	return &pb.ListJobResultsRes{
		JobResults: jobsRes,
	}, nil
}

func (s *Server) FindLatestJobResult(ctx context.Context, req *pb.JobResultsReq) (*pb.JobResultsRes, error) {
	creatorID, err := toUUID(req.CreatorId)
	if err != nil {
		return &pb.JobResultsRes{}, fmt.Errorf("Invalid creator id, %w", err)
	}

	jobID, err := toUUID(req.JobId)
	if err != nil {
		return &pb.JobResultsRes{}, fmt.Errorf("Invalid job id, %w", err)
	}

	job, err := s.repo.FindLatestJobResult(ctx, repo.FindLatestJobResultParams{
		CreatorID: creatorID,
		JobID:     jobID,
	})
	if err != nil {
		return &pb.JobResultsRes{}, fmt.Errorf("Error finding the latest job response, %w", err)
	}

	return &pb.JobResultsRes{
		Output:    job.Output,
		CreatedAt: timestamppb.New(job.CreatedAt.Time),
	}, nil
}

// Users server.
func (s *Server) FindUserByEmail(ctx context.Context, req *pb.UsersReq) (*pb.UsersRes, error) {
	userEmail := req.GetEmail()

	user, err := s.repo.FindUserByEmail(ctx, userEmail)
	if err != nil {
		return &pb.UsersRes{}, fmt.Errorf("Invalid user email, %w", err)
	}

	return &pb.UsersRes{
		Id:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		UpdatedAt: timestamppb.New(user.UpdatedAt.Time),
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}

func (s *Server) Register(ctx context.Context, req *pb.UsersReq) (*pb.UsersRes, error) {
	createdUser, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Password:  req.Username,
		Email:     req.Email,
	})
	if err != nil {
		return &pb.UsersRes{}, fmt.Errorf("Error creating the user, %w", err)
	}

	return &pb.UsersRes{
		Id:        createdUser.ID.String(),
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Username:  createdUser.Username,
		Email:     createdUser.Email,
		Password:  createdUser.Password,
		UpdatedAt: timestamppb.New(createdUser.UpdatedAt.Time),
		CreatedAt: timestamppb.New(createdUser.CreatedAt.Time),
	}, nil
}

// Sessions server.
func (s *Server) FindSessionByID(ctx context.Context, req *pb.SessionsReq) (*pb.SessionsRes, error) {
	session, err := s.repo.FindSessionByID(ctx, req.GetId())
	if err != nil {
		return &pb.SessionsRes{}, fmt.Errorf("Error getting the session, %w", err)
	}

	return &pb.SessionsRes{
		Id:           session.ID,
		UserEmail:    session.UserEmail,
		RefreshToken: session.RefreshToken,
		IsRevoked:    session.IsRevoked,
		CreatedAt:    timestamppb.New(session.CreatedAt),
		ExpiresAt:    timestamppb.New(session.ExpiresAt),
	}, nil
}

func (s *Server) FindSessionByEmail(ctx context.Context, req *pb.SessionsReq) (*pb.ListSessionsRes, error) {
	sessions, err := s.repo.FindSessionsByEmail(ctx, req.GetUserEmail())
	if err != nil {
		return &pb.ListSessionsRes{}, fmt.Errorf("Error getting the sessions, %w", err)
	}

	// Iterate through all the sessions and convert them into session responses.
	sessionsRes := make([]*pb.SessionsRes, 0, len(sessions))
	for _, session := range sessions {
		sessionsRes = append(sessionsRes, &pb.SessionsRes{
			Id:           session.ID,
			UserEmail:    session.UserEmail,
			RefreshToken: session.RefreshToken,
			IsRevoked:    session.IsRevoked,
			CreatedAt:    timestamppb.New(session.CreatedAt),
			ExpiresAt:    timestamppb.New(session.ExpiresAt),
		})
	}

	return &pb.ListSessionsRes{
		Sessions: sessionsRes,
	}, nil
}

func (s *Server) RevokeSessions(ctx context.Context, req *pb.SessionsReq) (*pb.ListSessionsRes, error) {
	revokedSessions, err := s.repo.RevokeSessions(ctx, req.GetUserEmail())
	if err != nil {
		return &pb.ListSessionsRes{}, fmt.Errorf("Error revoking the sessions, %w", err)
	}

	// Iterate through all the sessions and convert them into session responses.
	revokedSessionsRes := make([]*pb.SessionsRes, 0, len(revokedSessions))
	for _, session := range revokedSessions {
		revokedSessionsRes = append(revokedSessionsRes, &pb.SessionsRes{
			Id:           session.ID,
			UserEmail:    session.UserEmail,
			RefreshToken: session.RefreshToken,
			IsRevoked:    session.IsRevoked,
			CreatedAt:    timestamppb.New(session.CreatedAt),
			ExpiresAt:    timestamppb.New(session.ExpiresAt),
		})
	}

	return &pb.ListSessionsRes{
		Sessions: revokedSessionsRes,
	}, nil
}

func (s *Server) DeleteSessions(ctx context.Context, req *pb.SessionsReq) (*pb.ListSessionsRes, error) {
	deletedSessions, err := s.repo.DeteleSessions(ctx, req.GetUserEmail())
	if err != nil {
		return &pb.ListSessionsRes{}, fmt.Errorf("Error deleting sessions, %w", err)
	}

	// Iterate through all the sessions and convert them into session responses.
	deletedSessionsRes := make([]*pb.SessionsRes, 0, len(deletedSessions))
	for _, session := range deletedSessions {
		deletedSessionsRes = append(deletedSessionsRes, &pb.SessionsRes{
			Id:           session.ID,
			UserEmail:    session.UserEmail,
			RefreshToken: session.RefreshToken,
			IsRevoked:    session.IsRevoked,
			CreatedAt:    timestamppb.New(session.CreatedAt),
			ExpiresAt:    timestamppb.New(session.ExpiresAt),
		})
	}

	return &pb.ListSessionsRes{
		Sessions: deletedSessionsRes,
	}, nil
}

func (s *Server) CreateSession(ctx context.Context, req *pb.SessionsReq) (*pb.SessionsRes, error) {
	createdSession, err := s.repo.CreateSession(ctx, repo.CreateSessionParams{
		ID:           req.Id,
		UserEmail:    req.UserEmail,
		RefreshToken: req.RefreshToken,
		IsRevoked:    req.IsRevoked,
		ExpiresAt:    req.ExpiresAt.AsTime(),
	})
	if err != nil {
		return &pb.SessionsRes{}, fmt.Errorf("Error creating the session, %w", err)
	}

	return &pb.SessionsRes{
		Id:           createdSession.ID,
		UserEmail:    createdSession.UserEmail,
		RefreshToken: createdSession.RefreshToken,
		IsRevoked:    createdSession.IsRevoked,
		CreatedAt:    timestamppb.New(createdSession.CreatedAt),
		ExpiresAt:    timestamppb.New(createdSession.ExpiresAt),
	}, nil
}
