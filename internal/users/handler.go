package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler is the HTTP entry point for users endpoints.
// Think of this like a controller in other frameworks.
// It should focus only on request/response concerns.
type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users/", h.handleUsers)
}

func (h *Handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Always return JSON from this endpoint.
	w.Header().Set("Content-Type", "application/json")

	// Extract the ID from the URL path if present (format: /users/123).
	// http.ServeMux matches both /users and /users/ by default.
	path := strings.TrimPrefix(r.URL.Path, "/users")
	path = strings.TrimPrefix(path, "/")

	// If path is not empty, we have an ID parameter.
	if path != "" {
		// Parse the ID from the path segment.
		id, err := strconv.Atoi(path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
			return
		}

		// Handle individual user endpoints (GET by ID, PUT, DELETE).
		h.handleUsersByID(w, r, id)
		return
	}

	// Handle collection endpoints (GET all, POST create).
	switch r.Method {
	case http.MethodGet:
		// GET /users: fetch all users via the service layer.
		users, err := h.service.ListUsers(r.Context())
		if err != nil {
			// Internal errors should not expose implementation details.
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
			return
		}
		writeJSON(w, http.StatusOK, users)

	case http.MethodPost:
		// POST /users: decode request payload into a temporary request DTO.
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}

		// If JSON is malformed or types mismatch, return 400.
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		// Pass validated/decoded input to service for business processing.
		user, err := h.service.CreateUser(r.Context(), CreateUserInput{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		// Resource created successfully.
		writeJSON(w, http.StatusCreated, user)

	default:
		// Any unsupported method for this route gets 405.
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// handleUsersByID handles read/update/delete operations for a single user resource.
func (h *Handler) handleUsersByID(w http.ResponseWriter, r *http.Request, id int) {
	switch r.Method {
	case http.MethodGet:
		// GET /users/{id}: fetch a specific user.
		user, err := h.service.GetUser(r.Context(), id)
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, user)

	case http.MethodPut:
		// PUT /users/{id}: update a specific user.
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		}

		// Decode the request body.
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		// Pass to service to update the user.
		user, err := h.service.UpdateUser(r.Context(), id, CreateUserInput{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		// Return updated user.
		writeJSON(w, http.StatusOK, user)

	case http.MethodDelete:
		// DELETE /users/{id}: remove a specific user.
		err := h.service.DeleteUser(r.Context(), id)
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		// Return 204 No Content for successful deletion.
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
	if errors.Is(err, ErrEmailAlreadyInUse) || errors.Is(err, ErrUserHasExpenses) {
		return http.StatusConflict
	}
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
