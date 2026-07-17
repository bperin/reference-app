package content

import (
	"encoding/json"
	"net/http"

	"github.com/example/reference-app/internal/http/response"
)

type Verifier interface {
	Verify(r *http.Request) error
}

type EventarcHandler struct {
	service  *Service
	verifier Verifier
}

func NewEventarcHandler(service *Service, verifier Verifier) *EventarcHandler {
	return &EventarcHandler{service: service, verifier: verifier}
}

type storageEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func (h *EventarcHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := h.verifier.Verify(r); err != nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	var event struct {
		Data storageEvent `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid event", err.Error())
		return
	}

	if _, err := h.service.Complete(r.Context(), event.Data.Bucket, event.Data.Name); err != nil {
		response.Error(w, http.StatusInternalServerError, "internal error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
