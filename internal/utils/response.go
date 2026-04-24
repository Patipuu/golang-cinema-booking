package utils

import (
	"encoding/json"
	"net/http"
)

// Response is the standard JSON response format.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// WriteJSON sends a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// JSONSuccess sends a successful Response.
func JSONSuccess(w http.ResponseWriter, data any, message string) {
	WriteJSON(w, http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// JSONCreated sends a 201 Created Response.
func JSONCreated(w http.ResponseWriter, data any, message string) {
	WriteJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// JSONError sends error response with message and status.
func JSONError(w http.ResponseWriter, message string, status int) {
	WriteJSON(w, status, Response{Success: false, Error: message, Message: message})
}

// JSONBadRequest sends a 400 Bad Request Error.
func JSONBadRequest(w http.ResponseWriter, message string, errors any) {
	WriteJSON(w, http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Error:   message,
		Errors:  errors,
	})
}

// JSONUnauthorized sends a 401 Unauthorized Error.
func JSONUnauthorized(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusUnauthorized, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}

// JSONForbidden sends a 403 Forbidden Error.
func JSONForbidden(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusForbidden, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}

// JSONNotFound sends a 404 Not Found Error.
func JSONNotFound(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusNotFound, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}

// JSONInternal sends a 500 Internal Server Error.
func JSONInternal(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
		Error:   message,
	})
}
