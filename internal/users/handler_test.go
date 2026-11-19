package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestUsersHandler(t *testing.T) {
	t.Run("create success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		requestBody := map[string]interface{}{
			"first_name": "Alice",
			"last_name":  "Smith",
			"email":      "alice@example.com",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/users/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleUsers(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp User
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != 1 {
			t.Fatalf("expected ID 1, got %d", resp.ID)
		}
		if resp.FirstName != "Alice" {
			t.Fatalf("expected first_name 'Alice', got %s", resp.FirstName)
		}
		if resp.LastName != "Smith" {
			t.Fatalf("expected last_name 'Smith', got %s", resp.LastName)
		}
		if resp.Email != "alice@example.com" {
			t.Fatalf("expected email 'alice@example.com', got %s", resp.Email)
		}
	})

	t.Run("create duplicate email", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		req := httptest.NewRequest("POST", "/users/", bytes.NewBufferString(`{"first_name":"Alicia","last_name":"Smith","email":"ALICE@example.com"}`))
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("get success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		createdUser, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		req := httptest.NewRequest("GET", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp User
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != createdUser.ID {
			t.Fatalf("expected ID %d, got %d", createdUser.ID, resp.ID)
		}
		if resp.FirstName != "Bob" {
			t.Fatalf("expected first_name 'Bob', got %s", resp.FirstName)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("GET", "/users/999", nil)
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

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

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating first user: %v", err)
		}
		_, err = svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating second user: %v", err)
		}

		req := httptest.NewRequest("GET", "/users/", nil)
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp []User
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp) != 2 {
			t.Fatalf("expected 2 users, got %d", len(resp))
		}
	})

	t.Run("update success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		created, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		requestBody := map[string]interface{}{
			"first_name": "Alicia",
			"last_name":  "Smith",
			"email":      "alicia@example.com",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/users/"+strconv.Itoa(created.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp User
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.FirstName != "Alicia" {
			t.Fatalf("expected first_name 'Alicia', got %s", resp.FirstName)
		}
		if resp.Email != "alicia@example.com" {
			t.Fatalf("expected email 'alicia@example.com', got %s", resp.Email)
		}
	})

	t.Run("delete success", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Charlie", LastName: "Brown", Email: "charlie@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		req := httptest.NewRequest("DELETE", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.handleUsers(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", w.Code)
		}

		req2 := httptest.NewRequest("GET", "/users/1", nil)
		w2 := httptest.NewRecorder()
		handler.handleUsers(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d", w2.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)
		handler := NewHandler(svc)

		req := httptest.NewRequest("POST", "/users/", bytes.NewReader([]byte("{ invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.handleUsers(w, req)
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
