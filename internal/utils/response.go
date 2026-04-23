package utils

import (
	"encoding/json"
	"net/http"
)

// JSON response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteJSON writes status and JSON body to w.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// JSONSuccess sends success response with data.
func JSONSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, Response{Success: true, Data: data})
}

// JSONError sends error response with message and status.
func JSONError(w http.ResponseWriter, message string, status int) {
	WriteJSON(w, status, Response{Success: false, Error: message})
}

// JSONBadRequest sends 400.
func JSONBadRequest(w http.ResponseWriter, message string) {
	JSONError(w, message, http.StatusBadRequest)
}

// JSONUnauthorized sends 401.
func JSONUnauthorized(w http.ResponseWriter, message string) {
	JSONError(w, message, http.StatusUnauthorized)
}

// JSONNotFound sends 404.
func JSONNotFound(w http.ResponseWriter, message string) {
	JSONError(w, message, http.StatusNotFound)
}

// JSONForbidden sends 403.
func JSONForbidden(w http.ResponseWriter, message string) {
	JSONError(w, message, http.StatusForbidden)
}

// JSONInternal sends 500.
func JSONInternal(w http.ResponseWriter, message string) {
	JSONError(w, message, http.StatusInternalServerError)
}
