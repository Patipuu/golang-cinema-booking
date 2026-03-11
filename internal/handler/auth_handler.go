package handler

import (
	"net/http"
	"encoding/json"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils"
	"github.com/go-playground/validator/v10"
)

// AuthHandler groups HTTP handlers for authentication.
type AuthHandler struct{
	authService service.AuthService
	validator   *validator.Validate
}

// NewAuthHandler khởi tạo handler với service + validator.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   utils.NewValidator(),
	}
}

// Request DTOs

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}


// Register handles user registration requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid JSON body")
		return
	}

	// Validate input
	if err := utils.ValidateStruct(h.validator, &req); err != nil {
		utils.JSONBadRequest(w, err.Error())
		return
	}

	// Gọi service 
	user, err := h.authService.Register(req.Email, req.Password, req.FullName, req.Username)
	if err != nil {
		utils.JSONInternal(w, err.Error())
		return
	}

	// Trả về JSON theo envelope chuẩn
	utils.JSONSuccess(w, map[string]interface{}{
		"user": user,
	})
}

// Login handles user login requests.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	// Parse JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid JSON body")
		return
	}

	// Validate input
	if err := utils.ValidateStruct(h.validator, &req); err != nil {
		utils.JSONBadRequest(w, err.Error())
		return
	}

	// Gọi service: trả về user + token
	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		utils.JSONUnauthorized(w, err.Error())
		return
	}

	// Trả về user + token JWT
	utils.JSONSuccess(w, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

