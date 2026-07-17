package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/reference-app/internal/http/response"
	"github.com/example/reference-app/internal/users"
)

const maxRequestBodyBytes = 1 << 20

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With(slog.String("component", "auth_handler")),
	}
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type RegisterResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type RevokeRequest struct {
	Token string `json:"token"`
}

// Register creates a user account.
// @Summary Register a user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration payload"
// @Success 201 {object} RegisterResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 409 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var request RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}

	user, err := h.service.RegisterUser(
		r.Context(),
		strings.TrimSpace(request.Email),
		request.Password,
		strings.TrimSpace(request.DisplayName),
	)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrInvalidInput):
			response.Error(w, http.StatusBadRequest, "invalid_request", "email, password, and display_name are required")
		case errors.Is(err, users.ErrConflictEmail):
			response.Error(w, http.StatusConflict, "email_conflict", "email is already registered")
		default:
			h.logger.ErrorContext(r.Context(), "register user", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusCreated, RegisterResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
	})
}

// Token exchanges password or refresh-token grants for OAuth2 tokens.
// @Summary Exchange an OAuth2 token grant
// @Description Supports password (username, password, optional scope) and refresh_token (refresh_token) grants.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body TokenRequest true "OAuth2 grant payload"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /auth/oauth/token [post]
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var request TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}

	var (
		tokenResponse *TokenResponse
		err           error
	)
	switch request.GrantType {
	case "password":
		tokenResponse, err = h.service.ExchangePassword(
			r.Context(),
			strings.TrimSpace(request.Username),
			request.Password,
			strings.Fields(request.Scope),
		)
	case "refresh_token":
		tokenResponse, err = h.service.Refresh(r.Context(), request.RefreshToken)
	default:
		response.Error(w, http.StatusBadRequest, "unsupported_grant_type", "grant_type must be password or refresh_token")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid_grant", "invalid username or password")
		case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrRefreshTokenRevoked), errors.Is(err, ErrRefreshTokenExpired):
			response.Error(w, http.StatusUnauthorized, "invalid_grant", "invalid refresh token")
		default:
			h.logger.ErrorContext(r.Context(), "exchange OAuth token", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		}
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	response.JSON(w, http.StatusOK, tokenResponse)
}

// Revoke invalidates a refresh token when it exists.
// @Summary Revoke a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RevokeRequest true "Refresh token payload"
// @Success 204
// @Failure 400 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /auth/oauth/revoke [post]
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	defer r.Body.Close()

	var request RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}
	if err := h.service.Revoke(r.Context(), request.Token); err != nil {
		if !errors.Is(err, ErrInvalidRefreshToken) {
			h.logger.ErrorContext(r.Context(), "revoke OAuth token", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
