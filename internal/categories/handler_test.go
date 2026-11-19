package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestCategoriesHandler(t *testing.T) {
	t.Run("create success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		requestBody := map[string]interface{}{"description": "Food"}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/categories/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleCategories(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp Category
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
		if resp.Description != "Food" {
			t.Fatalf("expected description 'Food', got %s", resp.Description)
		}
	})

	t.Run("create duplicate category", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		req := httptest.NewRequest("POST", "/categories/", bytes.NewBufferString(`{"description":"food"}`))
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("get success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		createdCategory, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Transport"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		req := httptest.NewRequest("GET", "/categories/1", nil)
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Category
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != createdCategory.ID {
			t.Fatalf("expected ID %d, got %d", createdCategory.ID, resp.ID)
		}
		if resp.Description != "Transport" {
			t.Fatalf("expected description 'Transport', got %s", resp.Description)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("GET", "/categories/999", nil)
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

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

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating first category: %v", err)
		}
		_, err = svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Transport"})
		if err != nil {
			t.Fatalf("unexpected error creating second category: %v", err)
		}

		req := httptest.NewRequest("GET", "/categories/", nil)
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []Category
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(resp))
		}
	})

	t.Run("update success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		created, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		requestBody := map[string]interface{}{"description": "Groceries"}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/categories/"+strconv.Itoa(created.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp Category
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Description != "Groceries" {
			t.Fatalf("expected description 'Groceries', got %s", resp.Description)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Health"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		req := httptest.NewRequest("DELETE", "/categories/1", nil)
		w := httptest.NewRecorder()
		handler.handleCategories(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", w.Code)
		}

		req2 := httptest.NewRequest("GET", "/categories/1", nil)
		w2 := httptest.NewRecorder()
		handler.handleCategories(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d", w2.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("POST", "/categories/", bytes.NewReader([]byte("{ invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleCategories(w, req)
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
