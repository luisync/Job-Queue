package jobs

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/jobqueue-grpc/pb"
	"github.com/luisync/Job-Queue/internal/json"
	"github.com/luisync/Job-Queue/internal/users"
)

// Handlers depend on the services.
type handler struct {
	client pb.JobqueueClient
}

// Constructor for creating the handlers.
func NewHandler(client pb.JobqueueClient) *handler {
	return &handler{
		client: client,
	}
}

// List all jobs belonging to the user currently logged in.
func (h *handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the context.
	userID, err := users.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Get user jobs.
	jobs, err := h.client.ListJobs(r.Context(), &pb.JobsReq{
		CreatorId: userID.String(),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, jobs)
}

// Find job by ID.
func (h *handler) FindJob(w http.ResponseWriter, r *http.Request) {
	// Get the jobs id.
	jobIDStr := chi.URLParam(r, "jobID")

	// Convert string into UUID.
	var jobID pgtype.UUID
	if err := jobID.Scan(jobIDStr); err != nil {
		log.Println(err)
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	// Get the user ID from the context.
	userID, err := users.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Find the job with the UUID that was created by the logged in user.
	job, err := h.client.FindJob(r.Context(), &pb.FindJobReq{
		Id:        jobID.String(),
		CreatorId: userID.String(),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, job)
}

// Create a new job.
func (h *handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	// Extract the new job's contents from the request.
	var newJobInput createJobReq
	if err := json.Read(r, &newJobInput); err != nil {
		log.Println(err)
		http.Error(w, "Please include the correct fields.", http.StatusBadRequest)
		return
	}

	creatorID, err := users.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	newJob := createJobParams{
		CreatorID:    creatorID,
		createJobReq: newJobInput,
	}

	// Validade payload.
	switch {
	case len(newJob.Function) == 0:
		log.Println("Missing function.")
		http.Error(w, "Please include a function.", http.StatusBadRequest)
		return

	case len(newJob.Language) == 0:
		log.Println("Missing language.")
		http.Error(w, "Please include a language.", http.StatusBadRequest)
		return
	}

	// Create the new job.
	createdJob, err := h.client.CreateJob(r.Context(), &pb.JobsReq{
		CreatorId: newJob.CreatorID.String(),
		Job: &pb.Job{
			Language:     string(newJob.Language),
			Dependencies: newJob.Dependencies,
			Function:     newJob.Function,
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, createdJob)
}

// List all results of a job that was created by the user that's currently logged in.
func (h *handler) ListJobResults(w http.ResponseWriter, r *http.Request) {
	// Get job id from the URL.
	jobID := chi.URLParam(r, "jobID")

	creatorID, err := users.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Get user jobs.
	jobs, err := h.client.ListJobResults(r.Context(), &pb.JobResultsReq{
		JobId:     jobID,
		CreatorId: creatorID.String(),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, jobs)
}

// Find the latest job result from the user that's currently logged in.
func (h *handler) FindLatestJobResult(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	creatorID, err := users.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	// Get user jobs.
	job, err := h.client.FindLatestJobResult(r.Context(), &pb.JobResultsReq{
		CreatorId: creatorID.String(),
		JobId:     jobID,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, job)
}
