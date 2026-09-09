package json

import (
	"encoding/json"
	"net/http"
)

// Return a response to the user in JSON format.
func Write(w http.ResponseWriter, status int, data any) {
	// Set headers.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Write the jobs to the user in a JSON format.
	json.NewEncoder(w).Encode(data)
}
