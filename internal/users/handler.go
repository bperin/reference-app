package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/example/reference-app/internal/http/response"
	"github.com/example/reference-app/internal/logging"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With(slog.String("component", "users_handler")),
	}
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:   u.UpdatedAt.Format(http.TimeFormat),
	}
}

// GetMe returns the authenticated user's profile.
// @Summary Get the current user
// @Tags users
// @Produce json
// @Security OAuth2Auth[user]
// @Success 200 {object} UserResponse
// @Failure 401 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	u, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "user profile not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to get profile", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toUserResponse(u))
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
}

// UpdateMe updates the authenticated user's profile.
// @Summary Update the current user
// @Tags users
// @Accept json
// @Produce json
// @Security OAuth2Auth[user]
// @Param request body UpdateProfileRequest true "Profile update payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/me [put]
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}

	u, err := h.service.UpdateProfile(r.Context(), userID, req.DisplayName)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "invalid_request", "display_name cannot be empty")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to update profile", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toUserResponse(u))
}

// DeleteMe deletes the authenticated user's account.
// @Summary Delete the current user
// @Tags users
// @Security OAuth2Auth[user]
// @Success 204
// @Failure 401 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /users/me [delete]
func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	if err := h.service.DeleteAccount(r.Context(), userID); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to delete account", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}
