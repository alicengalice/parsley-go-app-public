package categories

import (
	"context"
	"errors"
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
		cat, err := repo.Create(ctx, Category{Description: "Food"})
		if err != nil {
			t.Errorf("create failed: %v", err)
		}
		if cat.ID == 0 {
			t.Errorf("expected category id to be set, got 0")
		}

		// Get
		fetched, err := repo.Get(ctx, cat.ID)
		if err != nil {
			t.Errorf("get failed: %v", err)
		}
		if fetched.Description != "Food" {
			t.Errorf("expected description Food, got %s", fetched.Description)
		}

		// Update
		updated, err := repo.Update(ctx, cat.ID, Category{Description: "Groceries"})
		if err != nil {
			t.Errorf("update failed: %v", err)
		}
		if updated.Description != "Groceries" {
			t.Errorf("expected description Groceries, got %s", updated.Description)
		}

		// List
		all, err := repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 1 {
			t.Errorf("expected 1 category, got %d", len(all))
		}

		// Delete
		if err := repo.Delete(ctx, cat.ID); err != nil {
			t.Errorf("delete failed: %v", err)
		}
		if _, err := repo.Get(ctx, cat.ID); err == nil {
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
					_, err := repo.Update(ctx, 999, Category{Description: "test"})
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

	t.Run("case-insensitive unique description", func(t *testing.T) {
		cat, err := repo.Create(ctx, Category{Description: "Utilities"})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		_, err = repo.Create(ctx, Category{Description: "UTILITIES"})
		if !errors.Is(err, ErrCategoryAlreadyExists) {
			t.Fatalf("expected ErrCategoryAlreadyExists, got %v", err)
		}

		if err := repo.Delete(ctx, cat.ID); err != nil {
			t.Fatalf("cleanup delete failed: %v", err)
		}
	})

	t.Run("delete category with expenses", func(t *testing.T) {
		category, err := repo.Create(ctx, Category{Description: "Referenced"})
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `CREATE TABLE expenses (category_id INTEGER NOT NULL REFERENCES categories(id))`); err != nil {
			t.Fatalf("create expenses table: %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `INSERT INTO expenses (category_id) VALUES (?)`, category.ID); err != nil {
			t.Fatalf("create expense: %v", err)
		}

		if err := repo.Delete(ctx, category.ID); !errors.Is(err, ErrCategoryHasExpenses) {
			t.Fatalf("expected ErrCategoryHasExpenses, got %v", err)
		}
		if _, err := dbConn.ExecContext(ctx, `DELETE FROM expenses WHERE category_id = ?`, category.ID); err != nil {
			t.Fatalf("cleanup expense: %v", err)
		}
		if err := repo.Delete(ctx, category.ID); err != nil {
			t.Fatalf("cleanup category: %v", err)
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
		cats := []Category{
			{Description: "Food"},
			{Description: "Transport"},
			{Description: "Utilities"},
		}
		for _, c := range cats {
			repo.Create(ctx, c)
		}

		all, err = repo.List(ctx)
		if err != nil {
			t.Errorf("list failed: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 categories, got %d", len(all))
		}

		// Ordering
		for i := 0; i < len(all)-1; i++ {
			if all[i].ID > all[i+1].ID {
				t.Errorf("expected ascending order, got %d > %d", all[i].ID, all[i+1].ID)
			}
		}
	})
}
