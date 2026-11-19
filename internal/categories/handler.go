package categories

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler is the HTTP entry point for categories endpoints.
// Think of this like a controller in other frameworks.
// It should focus only on request/response concerns.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/categories/", h.handleCategories)
}

func (h *Handler) handleCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract the ID from the URL path if present (format: /categories/123).
	path := strings.TrimPrefix(r.URL.Path, "/categories")
	path = strings.TrimPrefix(path, "/")

	if path != "" {
		id, err := strconv.Atoi(path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid category id"})
			return
		}
		h.handleCategoriesByID(w, r, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		categories, err := h.service.ListCategories(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
			return
		}
		writeJSON(w, http.StatusOK, categories)

	case http.MethodPost:
		var req struct {
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		category, err := h.service.CreateCategory(r.Context(), CreateCategoryInput{
			Description: req.Description,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, category)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// handleCategoriesByID handles read/update/delete operations for a single category resource.
func (h *Handler) handleCategoriesByID(w http.ResponseWriter, r *http.Request, id int) {
	switch r.Method {
	case http.MethodGet:
		category, err := h.service.GetCategory(r.Context(), id)
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, category)

	case http.MethodPut:
		var req struct {
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		category, err := h.service.UpdateCategory(r.Context(), id, CreateCategoryInput{
			Description: req.Description,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, category)

	case http.MethodDelete:
		err := h.service.DeleteCategory(r.Context(), id)
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// writeJSON serializes payloads and writes the HTTP status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// statusFromError maps domain/repository errors into appropriate HTTP status codes.
func statusFromError(err error) int {
	if errors.Is(err, ErrCategoryAlreadyExists) || errors.Is(err, ErrCategoryHasExpenses) {
		return http.StatusConflict
	}
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
