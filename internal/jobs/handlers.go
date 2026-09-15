package jobs

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/json"
	"github.com/luisync/Job-Queue/internal/users"
)

// Handlers depend on the services.
type handler struct {
	service Service
}

// Constructor for creating the handlers.
func NewHandler(service Service) *handler {
	return &handler{
		service: service,
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
	jobs, err := h.service.ListJobs(r.Context(), userID)
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
	ids := findJobByIDParams{
		ID:         jobID,
		Creator_id: userID,
	}

	job, err := h.service.FindJob(r.Context(), ids)
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
	createdJob, err := h.service.CreateJob(r.Context(), newJob)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server error, please try again later.", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, createdJob)
}
