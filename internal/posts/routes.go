package posts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler, requireBearer func(http.Handler) http.Handler) {
	r.Route("/posts", func(r chi.Router) {
		// Public endpoints.
		r.Get("/", h.ListPosts)
		r.Get("/{postID}", h.GetPost)

		r.Group(func(r chi.Router) {
			r.Use(requireBearer)
			r.Post("/", h.CreatePost)
			r.Put("/{postID}", h.UpdatePost)
			r.Delete("/{postID}", h.DeletePost)
			r.Post("/{postID}/publish", h.PublishPost)
		})
	})
}
