package expenses

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestExpensesHandler(t *testing.T) {
	t.Run("create success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		requestBody := map[string]interface{}{
			"user_id":     1,
			"category_id": 2,
			"amount":      50.75,
			"description": "Lunch",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/expenses/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleExpenses(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp Expense
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
		if resp.Date.IsZero() {
			t.Fatal("expected date to be set, got zero value")
		}
		if resp.UserID != 1 {
			t.Fatalf("expected user_id 1, got %d", resp.UserID)
		}
		if resp.Amount != 50.75 {
			t.Fatalf("expected amount 50.75, got %f", resp.Amount)
		}
		if resp.Description != "Lunch" {
			t.Fatalf("expected description 'Lunch', got %s", resp.Description)
		}
	})

	t.Run("get success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		createdExpense, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 2, Amount: 25.00, Description: "Coffee"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		req := httptest.NewRequest("GET", "/expenses/1", nil)
		w := httptest.NewRecorder()
		handler.handleExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Expense
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != createdExpense.ID {
			t.Fatalf("expected ID %d, got %d", createdExpense.ID, resp.ID)
		}
		if resp.Description != "Coffee" {
			t.Fatalf("expected description 'Coffee', got %s", resp.Description)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("GET", "/expenses/999", nil)
		w := httptest.NewRecorder()
		handler.handleExpenses(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}

		var errResp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp["error"] == "" {
			t.Fatal("expected error message in response")
		}
	})

	t.Run("list success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Lunch"})
		if err != nil {
			t.Fatalf("unexpected error creating first expense: %v", err)
		}
		_, err = svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 2, CategoryID: 2, Amount: 20.0, Description: "Dinner"})
		if err != nil {
			t.Fatalf("unexpected error creating second expense: %v", err)
		}

		req := httptest.NewRequest("GET", "/expenses/", nil)
		w := httptest.NewRecorder()
		handler.handleExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []Expense
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 expenses, got %d", len(resp))
		}
	})

	t.Run("update success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		created, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Lunch"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		requestBody := map[string]interface{}{
			"user_id":     1,
			"category_id": 1,
			"amount":      25.5,
			"description": "Brunch",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/expenses/"+strconv.Itoa(created.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.handleExpenses(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Expense
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Description != "Brunch" {
			t.Fatalf("expected description 'Brunch', got %s", resp.Description)
		}
		if resp.Amount != 25.5 {
			t.Fatalf("expected amount 25.5, got %f", resp.Amount)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.00, Description: "Test"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		req := httptest.NewRequest("DELETE", "/expenses/1", nil)
		w := httptest.NewRecorder()
		handler.handleExpenses(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", w.Code)
		}

		req2 := httptest.NewRequest("GET", "/expenses/1", nil)
		w2 := httptest.NewRecorder()
		handler.handleExpenses(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d", w2.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("POST", "/expenses/", bytes.NewReader([]byte("{ invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleExpenses(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}

		var errResp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp["error"] == "" {
			t.Fatal("expected error message for invalid JSON")
		}
	})
}
