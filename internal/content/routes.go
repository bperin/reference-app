package content

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, eh *EventarcHandler, requireBearer func(http.Handler) http.Handler) {
	r.Route("/contents", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(requireBearer)
			r.Post("/", h.Reserve)
		})
	})
    r.Post("/events/storage", eh.Handle)
}
