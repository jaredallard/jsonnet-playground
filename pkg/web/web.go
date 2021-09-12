// Package web stores the types for the web requests
package web

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// SaveCodeRequest is the request sent when saving code
type SaveCodeRequest struct {
	// Contents is the raw contents of the code to save
	Contents string `json:"code"`
}

// SaveCodeResponse is the response sent when saving code
type SaveCodeResponse struct {
	// ID is the ID of the code now that it's been saved.
	ID string `json:"id"`
}

// GetCodeResponse is the response sent when getting code
type GetCodeResponse struct {
	// Contents is the raw contents of the code stored at this ID
	Contents string `json:"contents"`
}

// ErrorResponse is the response sent when an error occurs
type ErrorResponse struct {
	// Message is the error message, aka reason
	Message string `json:"message"`
}

// SendErrorResponse sends an error over the wire
func SendErrorResponse(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	//nolint:errcheck // Why: err is only over-the-wire
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: err.Error(),
	})
}

// SendResponse sends a JSON response over the wire
func SendResponse(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		SendErrorResponse(w, errors.Wrap(err, "failed to write resp"), http.StatusInternalServerError)
		return
	}
}
