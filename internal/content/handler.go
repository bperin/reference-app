package content

import (
	"encoding/json"
	"net/http"

	"github.com/example/reference-app/internal/http/response"
	"github.com/example/reference-app/internal/logging"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type reserveRequest struct {
	ParentID *uuid.UUID `json:"parent_id"`
	MimeType string     `json:"mime_type"`
	Filename string     `json:"filename"`
}

// Reserve reserves a new content upload.
// @Summary Reserve content upload
// @Description Reserves a new content upload with MIME type and hierarchy.
// @Tags content
// @Accept json
// @Produce json
// @Param request body reserveRequest true "Reservation request"
// @Success 201 {object} map[string]interface{} "Reservation details"
// @Failure 400 {object} response.ErrorPayload "Invalid request"
// @Failure 401 {object} response.ErrorPayload "Unauthorized"
// @Failure 500 {object} response.ErrorPayload "Internal server error"
// @Router /contents [post]
func (h *Handler) Reserve(w http.ResponseWriter, r *http.Request) {
	userID, ok := logging.GetUserID(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id not in context")
		return
	}

	var req reserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request", err.Error())
		return
	}

	content, signedURL, err := h.service.Reserve(r.Context(), userID, req.ParentID, req.MimeType, req.Filename)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"content":    content,
		"signed_url": signedURL,
	})
}
