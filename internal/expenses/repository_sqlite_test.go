package expenses

import (
	"context"
	"errors"
	"testing"

	"github.com/alicengalice/parsley/internal/categories"
	"github.com/alicengalice/parsley/internal/platform/db"
	"github.com/alicengalice/parsley/internal/users"
)

func TestSQLiteRepository(t *testing.T) {
	ctx := context.Background()
	dbConn, err := db.OpenSQLite(ctx, ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer dbConn.Close()

	// expenses.user_id/category_id are foreign keys, so the referenced tables must exist first.
	usersRepo, err := users.NewSQLiteRepository(ctx, dbConn)
	if err != nil {
		t.Fatalf("failed to initialize users repository: %v", err)
	}
	categoriesRepo, err := categories.NewSQLiteRepository(ctx, dbConn)
	if err != nil {
		t.Fatalf("failed to initialize categories repository: %v", err)
	}
	if _, err := usersRepo.Create(ctx, users.User{FirstName: "Test", LastName: "User", Email: "test@example.com"}); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	if _, err := usersRepo.Create(ctx, users.User{FirstName: "Second", LastName: "User", Email: "second@example.com"}); err != nil {
		t.Fatalf("failed to seed second user: %v", err)
	}
	if _, err := categoriesRepo.Create(ctx, categories.Category{Description: "General"}); err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}
	if _, err := categoriesRepo.Create(ctx, categories.Category{Description: "Other"}); err != nil {
		t.Fatalf("failed to seed second category: %v", err)
	}

	repo, err := NewSQLiteRepository(ctx, dbConn)
	if err != nil {
		t.Fatalf("failed to initialize repository: %v", err)
	}

	t.Run("CRUD operations", func(t *testing.T) {
		// Create
		exp, err := repo.Create(ctx, Expense{
			UserID:      1,
			CategoryID:  2,
			Amount:      15.50,
			Description: "Lunch at office",
		})
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		if exp.ID == 0 {
			t.Errorf("expected expense id to be set, got 0")
		}
		if exp.Date.IsZero() {
			t.Errorf("expected date to be set, got zero value")
		}

		// Get
		fetched, err := repo.Get(ctx, exp.ID)
		if err != nil {
			t.Errorf("get failed: %v", err)
		}
		if fetched.Amount != 15.50 {
			t.Errorf("expected amount 15.50, got %f", fetched.Amount)
		}

		// Update
		updated, err := repo.Update(ctx, exp.ID, Expense{
			UserID:      1,
			CategoryID:  2,
			Amount:      18.75,
			Description: "Lunch and coffee",
		})
		if err != nil {
			t.Errorf("update failed: %v", err)
		}
		if updated.Amount != 18.75 {
			t.Errorf("expected amount 18.75, got %f", updated.Amount)
		}

		// List
		all, err := repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 expense, got %d", len(all))
		}

		// Delete
		if err := repo.Delete(ctx, exp.ID); err != nil {
			t.Errorf("delete failed: %v", err)
		}
		if _, err := repo.Get(ctx, exp.ID); err == nil {
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
					_, err := repo.Update(ctx, 999, Expense{
						UserID: 1, CategoryID: 1, Amount: 10, Description: "test",
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
			repo.Create(ctx, Expense{
				UserID: 1, CategoryID: 1, Amount: float64(i * 10), Description: "test",
			})
		}

		all, err = repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 expenses, got %d", len(all))
		}

		// Ordering
		for i := 0; i < len(all)-1; i++ {
			if all[i].ID > all[i+1].ID {
				t.Errorf("expected ascending order, got %d > %d", all[i].ID, all[i+1].ID)
			}
		}

		// Different users
		exp1, _ := repo.Create(ctx, Expense{UserID: 1, CategoryID: 1, Amount: 10, Description: "user1"})
		exp2, _ := repo.Create(ctx, Expense{UserID: 2, CategoryID: 1, Amount: 20, Description: "user2"})
		if exp1.UserID == exp2.UserID {
			t.Errorf("expected different user ids")
		}
	})

	t.Run("foreign key enforcement", func(t *testing.T) {
		if _, err := repo.Create(ctx, Expense{UserID: 9999, CategoryID: 1, Amount: 10, Description: "bad user"}); !errors.Is(err, ErrInvalidExpenseReference) {
			t.Errorf("expected ErrInvalidExpenseReference for nonexistent user_id, got %v", err)
		}
		if _, err := repo.Create(ctx, Expense{UserID: 1, CategoryID: 9999, Amount: 10, Description: "bad category"}); !errors.Is(err, ErrInvalidExpenseReference) {
			t.Errorf("expected ErrInvalidExpenseReference for nonexistent category_id, got %v", err)
		}

		valid, err := repo.Create(ctx, Expense{UserID: 1, CategoryID: 1, Amount: 10, Description: "valid"})
		if err != nil {
			t.Fatalf("unexpected error creating valid expense: %v", err)
		}
		if _, err := repo.Update(ctx, valid.ID, Expense{UserID: 9999, CategoryID: 1, Amount: 10, Description: "valid"}); !errors.Is(err, ErrInvalidExpenseReference) {
			t.Errorf("expected ErrInvalidExpenseReference updating nonexistent user_id, got %v", err)
		}
	})
}
