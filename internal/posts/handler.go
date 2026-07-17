package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/example/reference-app/internal/http/response"
	"github.com/example/reference-app/internal/logging"
)

type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With(slog.String("component", "posts_handler")),
	}
}

type PostResponse struct {
	ID        string `json:"id"`
	AuthorID  string `json:"author_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Published bool   `json:"published"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toPostResponse(p *Post) PostResponse {
	return PostResponse{
		ID:        p.ID.String(),
		AuthorID:  p.AuthorID.String(),
		Title:     p.Title,
		Content:   p.Content,
		Published: p.Published,
		CreatedAt: p.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: p.UpdatedAt.Format(http.TimeFormat),
	}
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreatePost creates a post for the authenticated user.
// @Summary Create a post
// @Tags posts
// @Accept json
// @Produce json
// @Security OAuth2Auth[user]
// @Param request body CreatePostRequest true "Post payload"
// @Success 201 {object} PostResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/ [post]
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}

	p, err := h.service.CreatePost(r.Context(), userID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "invalid_request", "title and content are required")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to create post", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, toPostResponse(p))
}

// GetPost returns one published post.
// @Summary Get a post
// @Tags posts
// @Produce json
// @Param postID path string true "Post UUID"
// @Success 200 {object} PostResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/{postID} [get]
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid post ID format")
		return
	}

	p, err := h.service.GetPost(r.Context(), postID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "post not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to get post", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toPostResponse(p))
}

// UpdatePost updates a post owned by the authenticated user.
// @Summary Update a post
// @Tags posts
// @Accept json
// @Produce json
// @Security OAuth2Auth[user]
// @Param postID path string true "Post UUID"
// @Param request body UpdatePostRequest true "Post payload"
// @Success 200 {object} PostResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 403 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/{postID} [put]
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	postIDStr := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid post ID format")
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "malformed JSON payload")
		return
	}

	p, err := h.service.UpdatePost(r.Context(), postID, userID, req.Title, req.Content)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "post not found")
			return
		}
		if errors.Is(err, ErrForbiddenOwner) {
			response.Error(w, http.StatusForbidden, "forbidden", "only the author can edit this post")
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "invalid_request", "title and content are required")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to update post", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toPostResponse(p))
}

// PublishPost publishes a post owned by the authenticated user.
// @Summary Publish a post
// @Tags posts
// @Produce json
// @Security OAuth2Auth[user]
// @Param postID path string true "Post UUID"
// @Success 200 {object} PostResponse
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 403 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/{postID}/publish [post]
func (h *Handler) PublishPost(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	postIDStr := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid post ID format")
		return
	}

	p, err := h.service.PublishPost(r.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "post not found")
			return
		}
		if errors.Is(err, ErrForbiddenOwner) {
			response.Error(w, http.StatusForbidden, "forbidden", "only the author can publish this post")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to publish post", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toPostResponse(p))
}

// DeletePost deletes a post owned by the authenticated user.
// @Summary Delete a post
// @Tags posts
// @Security OAuth2Auth[user]
// @Param postID path string true "Post UUID"
// @Success 204
// @Failure 400 {object} response.ErrorPayload
// @Failure 401 {object} response.ErrorPayload
// @Failure 403 {object} response.ErrorPayload
// @Failure 404 {object} response.ErrorPayload
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/{postID} [delete]
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, _ := logging.GetUserID(r.Context())

	postIDStr := chi.URLParam(r, "postID")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid post ID format")
		return
	}

	err = h.service.DeletePost(r.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "not_found", "post not found")
			return
		}
		if errors.Is(err, ErrForbiddenOwner) {
			response.Error(w, http.StatusForbidden, "forbidden", "only the author can delete this post")
			return
		}
		h.logger.ErrorContext(r.Context(), "failed to delete post", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}

// ListPosts returns published posts.
// @Summary List published posts
// @Tags posts
// @Produce json
// @Param limit query integer false "Maximum number of posts"
// @Param offset query integer false "Pagination offset"
// @Success 200 {array} PostResponse
// @Failure 500 {object} response.ErrorPayload
// @Router /posts/ [get]
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(10)
	if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
		limit = int32(l)
	}

	offset := int32(0)
	if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
		offset = int32(o)
	}

	// Always show published posts
	postsList, err := h.service.ListPosts(r.Context(), limit, offset, false)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to list posts", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "server_error", "internal server error")
		return
	}

	resp := make([]PostResponse, len(postsList))
	for i, p := range postsList {
		resp[i] = toPostResponse(p)
	}

	response.JSON(w, http.StatusOK, resp)
}
