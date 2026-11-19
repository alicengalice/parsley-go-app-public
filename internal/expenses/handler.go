package expenses

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Handler is the HTTP entry point for expenses endpoints.
// Think of this like a controller in other frameworks.
// It should focus only on request/response concerns.
type Handler struct {
	service Service
}

// LEARNING: Constructor + dependency injection.
// We inject the Service interface so the handler is decoupled from storage details
// and easier to test with alternate service implementations.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// LEARNING: Receiver methods are Go's instance methods.
// The receiver (h *Handler) is similar to 'this' and keeps related route logic grouped.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/expenses/", h.handleExpenses)
}

// LEARNING: Lowercase methods are package-private in Go.
// This keeps handler internals hidden while exposing only NewHandler/RegisterRoutes.
func (h *Handler) handleExpenses(w http.ResponseWriter, r *http.Request) {
	// Always return JSON from this endpoint.
	w.Header().Set("Content-Type", "application/json")

	// Extract the ID from the URL path if present (format: /expenses/123).
	// http.ServeMux matches both /expenses and /expenses/ by default.
	path := strings.TrimPrefix(r.URL.Path, "/expenses")
	path = strings.TrimPrefix(path, "/")

	// If path is not empty, we have an ID parameter.
	if path != "" {
		// Parse the ID from the path segment.
		id, err := strconv.Atoi(path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid expense id"})
			return
		}

		// Handle individual expense endpoints (GET by ID, PUT, DELETE).
		h.handleExpenseByID(w, r, id)
		return
	}

	// Handle collection endpoints (GET all, POST create).
	switch r.Method {
	case http.MethodGet:
		// GET /expenses: fetch all expenses via the service layer.
		expenses, err := h.service.ListExpenses(r.Context())
		if err != nil {
			// Internal errors should not expose implementation details.
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list expenses"})
			return
		}
		writeJSON(w, http.StatusOK, expenses)

	case http.MethodPost:
		// POST /expenses: decode request payload into a temporary request DTO.
		var req struct {
			UserID      int     `json:"user_id"`
			CategoryID  int     `json:"category_id"`
			Amount      float64 `json:"amount"`
			Description string  `json:"description"`
		}

		// If JSON is malformed or types mismatch, return 400.
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		// Pass validated/decoded input to service for business processing.
		expense, err := h.service.CreateExpense(r.Context(), CreateExpenseInput{
			UserID:      req.UserID,
			CategoryID:  req.CategoryID,
			Amount:      req.Amount,
			Description: req.Description,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		// Resource created successfully.
		writeJSON(w, http.StatusCreated, expense)

	default:
		// Any unsupported method for this route gets 405.
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

// LEARNING: REST-style routing is done manually with net/http.
// First branch by path (collection vs id route), then branch by HTTP method.
func (h *Handler) handleExpenseByID(w http.ResponseWriter, r *http.Request, id int) {
	switch r.Method {
	case http.MethodGet:
		// GET /expenses/{id}: fetch a specific expense.
		expense, err := h.service.GetExpense(r.Context(), id)
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, expense)

	case http.MethodPut:
		// PUT /expenses/{id}: update a specific expense.
		var req struct {
			UserID      int     `json:"user_id"`
			CategoryID  int     `json:"category_id"`
			Amount      float64 `json:"amount"`
			Description string  `json:"description"`
		}

		// Decode the request body.
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		// Pass to service to update the expense.
		expense, err := h.service.UpdateExpense(r.Context(), id, CreateExpenseInput{
			UserID:      req.UserID,
			CategoryID:  req.CategoryID,
			Amount:      req.Amount,
			Description: req.Description,
		})
		if err != nil {
			writeJSON(w, statusFromError(err), map[string]string{"error": err.Error()})
			return
		}

		// Return updated expense.
		writeJSON(w, http.StatusOK, expense)

	case http.MethodDelete:
		// DELETE /expenses/{id}: remove a specific expense.
		err := h.service.DeleteExpense(r.Context(), id)
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

// LEARNING: Single JSON writer avoids repeated response boilerplate.
// Keeping this centralized also makes later response-format changes safer.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// LEARNING: Lightweight Error-to-HTTP Mapping
// The handler translates domain/storage errors into HTTP status codes.
// - "...not found" -> 404
// - everything else here -> 400 (validation/business rule failure)
func statusFromError(err error) int {
	if errors.Is(err, ErrInvalidExpenseReference) {
		return http.StatusUnprocessableEntity
	}
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
