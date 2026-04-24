package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"booking_cinema_golang/internal/domain"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils"

	"github.com/go-playground/validator/v10"
)

// AuthHandler groups HTTP handlers for authentication.
type AuthHandler struct {
	authSvc  service.AuthService
	validate *validator.Validate
}

// NewAuthHandler creates an AuthHandler with injected service.
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, validate: validator.New()}
}

// userResponse is the safe public representation of a User — no password hash or OTP fields.
type userResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone"`
	Role       string    `json:"role"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

func toUserResponse(u *domain.User) userResponse {
	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}
	return userResponse{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		FullName:   u.FullName,
		Phone:      phone,
		Role:       u.Role,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
	}
}


// Register handles POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"     validate:"required,email"`
		Password string `json:"password"  validate:"required,min=8"`
		Username string `json:"username"  validate:"required,min=3,max=50"`
		FullName string `json:"full_name" validate:"required"`
		Phone    string `json:"phone"     validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		utils.JSONBadRequest(w, err.Error(), nil)
		return
	}

	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Username, req.FullName, req.Phone)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Response{
		Success: true,
		Message: "registration successful, please check your email for the verification code",
		Data:    toUserResponse(user),
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		utils.JSONBadRequest(w, err.Error(), nil)
		return
	}

	user, token, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.JSONSuccess(w, map[string]any{
		"token": token,
		"user":  toUserResponse(user),
	}, "login successful")
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	utils.JSONSuccess(w, nil, "Đăng xuất thành công")
}


// VerifyOTP handles POST /api/auth/verify-otp
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string `json:"user_id"  validate:"required"`
		OTPCode string `json:"otp_code" validate:"required,len=6"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		utils.JSONBadRequest(w, err.Error(), nil)
		return
	}

	if err := h.authSvc.VerifyOTP(r.Context(), req.UserID, req.OTPCode); err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.JSONSuccess(w, nil, "account verified successfully")
}

// ResendVerification handles POST /api/auth/resend-verification
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		utils.JSONBadRequest(w, err.Error(), nil)
		return
	}

	if err := h.authSvc.ResendVerification(r.Context(), req.Email); err != nil {
		h.handleServiceError(w, err)
		return
	}

	utils.JSONSuccess(w, nil, "verification code resent, please check your email")
}

// handleServiceError maps known service errors to appropriate HTTP status codes.
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists),
		errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrAlreadyVerified):
		utils.JSONError(w, err.Error(), http.StatusConflict)
	case errors.Is(err, service.ErrInvalidCredentials):
		utils.JSONError(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, service.ErrAccountNotVerified):
		utils.JSONError(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidOTP),
		errors.Is(err, service.ErrExpiredOTP):
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, service.ErrUserNotFound):
		utils.JSONNotFound(w, err.Error())
	default:
		fmt.Fprintf(os.Stderr, "Unexpected auth error: %v\n", err)
		utils.JSONInternal(w, "an unexpected error occurred")
	}
}
