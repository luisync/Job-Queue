package jobs

import (
	"log"
	"net/http"

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
	// Call the corresponding service function.
	err := h.service.ListJobs(r.Context())
	if err != nil {
		// Log error to console.
		log.Println(err)

		// Return a error response to the user.
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	jobs := []string{"Job1", "Job2"}

	// Send a response.
	json.Write(w, http.StatusOK, jobs)
}
