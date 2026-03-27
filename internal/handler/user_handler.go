package handler

import (
	"encoding/json"
	"net/http"

	"booking_cinema_golang/internal/middleware"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils"
)

type UserHandler struct {
	userSvc service.UserService
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// GetProfile handles GET /api/v1/users/me
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		utils.JSONUnauthorized(w, "unauthorized")
		return
	}

	user, err := h.userSvc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		if err == service.ErrUserNotFound {
			utils.JSONNotFound(w, "user not found")
			return
		}
		utils.JSONInternal(w, "failed to get profile")
		return
	}

	// Remove sensitive info
	user.PasswordHash = ""
	user.OTPCode = ""
	user.OTPExpiry = nil

	utils.JSONSuccess(w, user)
}

// UpdateProfileRequest defines the allowed fields to update.
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// UpdateProfile handles PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		utils.JSONUnauthorized(w, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body")
		return
	}

	user, err := h.userSvc.UpdateProfile(r.Context(), claims.UserID, req.FullName, req.Phone)
	if err != nil {
		if err == service.ErrUserNotFound {
			utils.JSONNotFound(w, "user not found")
			return
		}
		utils.JSONInternal(w, "failed to update profile")
		return
	}

	user.PasswordHash = ""
	user.OTPCode = ""
	user.OTPExpiry = nil

	utils.JSONSuccess(w, user)
}

// ChangePasswordRequest payload for changing password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword handles PUT /api/v1/users/me/password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.UserID == "" {
		utils.JSONUnauthorized(w, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONBadRequest(w, "invalid request body")
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		utils.JSONBadRequest(w, "old_password and new_password are required")
		return
	}

	err := h.userSvc.ChangePassword(r.Context(), claims.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "incorrect old password" {
			utils.JSONBadRequest(w, err.Error())
			return
		}
		if err == service.ErrUserNotFound {
			utils.JSONNotFound(w, "user not found")
			return
		}
		utils.JSONInternal(w, "failed to change password")
		return
	}

	utils.JSONSuccess(w, map[string]string{"message": "password updated successfully"})
}
