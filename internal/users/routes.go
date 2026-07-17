package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, requireBearer func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(requireBearer)
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", h.GetMe)
			r.Put("/me", h.UpdateMe)
			r.Delete("/me", h.DeleteMe)
		})
	})
}
