package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

const contentTypeJSON = "application/json"

// ErrInternalServerError is the message returned to clients for any 500; the
// real error should be logged server-side instead, since it may contain
// details (e.g. raw DB error text) that shouldn't be exposed externally.
var ErrInternalServerError = errors.New("internal server error")

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func OKResponse(w http.ResponseWriter, data any) {
	jsonResponse(w, http.StatusOK, data)
}

func CreatedResponse(w http.ResponseWriter, data any) {
	jsonResponse(w, http.StatusCreated, data)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, struct {
		Error string `json:"error"`
	}{Error: message})
}
