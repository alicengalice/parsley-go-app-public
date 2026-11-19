package users

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/alicengalice/parsley/internal/platform/db"
)

func TestSQLiteRepository(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.OpenSQLite(ctx, ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer dbConn.Close()

	repo, err := NewSQLiteRepository(ctx, dbConn)
	if err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	t.Run("CRUD operations", func(t *testing.T) {
		// Create
		user, err := repo.Create(ctx, User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		})
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		if user.ID == 0 {
			t.Errorf("expected user id to be set, got 0")
		}
		if user.CreatedAt.IsZero() {
			t.Errorf("expected created_at to be set, got zero value")
		}

		// Get
		fetched, err := repo.Get(ctx, user.ID)
		if err != nil {
			t.Errorf("get failed: %v", err)
		}
		if fetched.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", fetched.Email)
		}

		// Update
		updated, err := repo.Update(ctx, user.ID, User{
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
		})
		if err != nil {
			t.Errorf("update failed: %v", err)
		}
		if updated.FirstName != "Jane" {
			t.Errorf("expected first_name Jane, got %s", updated.FirstName)
		}

		// List
		all, err := repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 user, got %d", len(all))
		}

		// Delete
		if err := repo.Delete(ctx, user.ID); err != nil {
			t.Errorf("delete failed: %v", err)
		}
		if _, err := repo.Get(ctx, user.ID); err == nil {
			t.Errorf("expected error after delete, got nil")
		}
	})

	t.Run("error cases", func(t *testing.T) {
		tests := []struct {
			name string
			fn   func() error
		}{
			{
				name: "get nonexistent",
				fn: func() error {
					_, err := repo.Get(ctx, 999)
					return err
				},
			},
			{
				name: "update nonexistent",
				fn: func() error {
					_, err := repo.Update(ctx, 999, User{
						FirstName: "Test", LastName: "User", Email: "test@example.com",
					})
					return err
				},
			},
			{
				name: "delete nonexistent",
				fn: func() error {
					return repo.Delete(ctx, 999)
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if err := test.fn(); err == nil {
					t.Errorf("expected error, got nil")
				}
			})
		}
	})

	t.Run("case-insensitive unique email", func(t *testing.T) {
		user, err := repo.Create(ctx, User{FirstName: "Existing", LastName: "User", Email: "existing@example.com"})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		_, err = repo.Create(ctx, User{FirstName: "Duplicate", LastName: "User", Email: "EXISTING@example.com"})
		if !errors.Is(err, ErrEmailAlreadyInUse) {
			t.Fatalf("expected ErrEmailAlreadyInUse, got %v", err)
		}

		if err := repo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("cleanup delete failed: %v", err)
		}
	})

	t.Run("delete user with expenses", func(t *testing.T) {
		user, err := repo.Create(ctx, User{FirstName: "Expense", LastName: "Owner", Email: "owner@example.com"})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `CREATE TABLE expenses (user_id INTEGER NOT NULL REFERENCES users(id))`); err != nil {
			t.Fatalf("create expenses table: %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `INSERT INTO expenses (user_id) VALUES (?)`, user.ID); err != nil {
			t.Fatalf("create expense: %v", err)
		}

		if err := repo.Delete(ctx, user.ID); !errors.Is(err, ErrUserHasExpenses) {
			t.Fatalf("expected ErrUserHasExpenses, got %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `DELETE FROM expenses WHERE user_id = ?`, user.ID); err != nil {
			t.Fatalf("cleanup expense: %v", err)
		}
		if err := repo.Delete(ctx, user.ID); err != nil {
			t.Fatalf("cleanup user: %v", err)
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		// Empty list
		all, err := repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 0 {
			t.Errorf("expected empty list, got %d items", len(all))
		}

		// Multiple records
		for i := 1; i <= 3; i++ {
			_, err := repo.Create(ctx, User{FirstName: "User", LastName: "Test", Email: fmt.Sprintf("test%d@example.com", i)})
			if err != nil {
				t.Fatalf("create failed: %v", err)
			}
		}

		all, err = repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 users, got %d", len(all))
		}

		// Ordering
		for i := 0; i < len(all)-1; i++ {
			if all[i].ID > all[i+1].ID {
				t.Errorf("expected ascending order, got %d > %d", all[i].ID, all[i+1].ID)
			}
		}

		// Timestamp check
		if all[0].CreatedAt.IsZero() {
			t.Errorf("expected created_at to be set, got zero value")
		}
	})
}
