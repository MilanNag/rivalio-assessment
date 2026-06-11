package httpapi

import (
	"encoding/json"
	"net/http"
)

// errorBody is the consistent error envelope returned by every endpoint:
//
//	{"error": {"code": "validation_error", "message": "...", "fields": {"title": "..."}}}
type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorBody{Error: apiError{Code: code, Message: message}})
}

func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: apiError{
		Code:    "validation_error",
		Message: "One or more fields are invalid.",
		Fields:  fields,
	}})
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal_error", "Something went wrong. Please try again.")
}
