package users

import (
	"context"
	"testing"
)

func TestUsersService(t *testing.T) {
	t.Run("list empty", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		users, err := svc.ListUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 0 {
			t.Fatalf("expected 0 users, got %d", len(users))
		}
	})

	t.Run("create user", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		input := CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"}
		user, err := svc.CreateUser(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
		if user.FirstName != input.FirstName {
			t.Fatalf("expected first_name %q, got %q", input.FirstName, user.FirstName)
		}
		if user.Email != input.Email {
			t.Fatalf("expected email %q, got %q", input.Email, user.Email)
		}
	})

	t.Run("create validation", func(t *testing.T) {
		tests := []struct {
			name  string
			input CreateUserInput
		}{
			{name: "missing first name", input: CreateUserInput{FirstName: "", LastName: "Smith", Email: "a@b.com"}},
			{name: "missing last name", input: CreateUserInput{FirstName: "Alice", LastName: "", Email: "a@b.com"}},
			{name: "missing email", input: CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: ""}},
			{name: "invalid email", input: CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "notanemail"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := NewInMemoryRepository()
				svc := NewService(repo)

				_, err := svc.CreateUser(context.Background(), tt.input)
				if err == nil {
					t.Fatal("expected validation error")
				}
			})
		}
	})

	t.Run("create duplicate email rejected", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating first user: %v", err)
		}

		_, err = svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alicia", LastName: "Nguyen", Email: "ALICE@example.com"})
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})

	t.Run("update to duplicate email rejected", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating first user: %v", err)
		}
		bob, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating second user: %v", err)
		}

		_, err = svc.UpdateUser(context.Background(), bob.ID, CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "alice@example.com"})
		if err == nil {
			t.Fatal("expected error for duplicate email on update")
		}

		// Updating a user to keep their own email should still succeed.
		_, err = svc.UpdateUser(context.Background(), bob.ID, CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error updating user to their own email: %v", err)
		}
	})

	t.Run("list after create", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating first user: %v", err)
		}
		_, err = svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Bob", LastName: "Jones", Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating second user: %v", err)
		}

		users, err := svc.ListUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 2 {
			t.Fatalf("expected 2 users, got %d", len(users))
		}
	})

	t.Run("get user", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		user, err := svc.GetUser(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != created.ID {
			t.Fatalf("expected ID %d, got %d", created.ID, user.ID)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.GetUser(context.Background(), 0)
		if err == nil {
			t.Fatal("expected error for id=0")
		}
	})

	t.Run("update user", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		updated, err := svc.UpdateUser(context.Background(), created.ID, CreateUserInput{FirstName: "Alicia", LastName: "Smith", Email: "alicia@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.FirstName != "Alicia" {
			t.Fatalf("expected first_name %q, got %q", "Alicia", updated.FirstName)
		}
		if updated.Email != "alicia@example.com" {
			t.Fatalf("expected email %q, got %q", "alicia@example.com", updated.Email)
		}
	})

	t.Run("update invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.UpdateUser(context.Background(), 0, CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err == nil {
			t.Fatal("expected error for id=0")
		}
	})

	t.Run("delete user", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateUser(context.Background(), CreateUserInput{FirstName: "Alice", LastName: "Smith", Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("unexpected error creating user: %v", err)
		}

		err = svc.DeleteUser(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = svc.GetUser(context.Background(), created.ID)
		if err == nil {
			t.Fatal("expected error after deletion")
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		err := svc.DeleteUser(context.Background(), -1)
		if err == nil {
			t.Fatal("expected error for negative id")
		}
	})
}
