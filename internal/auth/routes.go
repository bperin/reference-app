package auth

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/oauth/token", h.Token)
	r.Post("/auth/oauth/revoke", h.Revoke)
}
