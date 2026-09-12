package jobs

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luisync/Job-Queue/internal/json"
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

// List all jobs.
func (h *handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Get all jobs.
	jobs, err := h.service.ListJobs(r.Context())
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
	}

	// Find the job with the UUID.
	job, err := h.service.FindJob(r.Context(), jobID)
	if err != nil {
		log.Println(err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	json.Write(w, http.StatusOK, job)
}
