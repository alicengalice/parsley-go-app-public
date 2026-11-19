package categories

import (
	"context"
	"testing"
)

func TestCategoriesService(t *testing.T) {
	t.Run("list empty", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		categories, err := svc.ListCategories(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(categories) != 0 {
			t.Fatalf("expected 0 categories, got %d", len(categories))
		}
	})

	t.Run("create category", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		cat, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cat.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
		if cat.Description != "Food" {
			t.Fatalf("expected description %q, got %q", "Food", cat.Description)
		}
	})

	t.Run("create duplicate description rejected", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating first category: %v", err)
		}

		_, err = svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "food"})
		if err == nil {
			t.Fatal("expected error for duplicate description")
		}
	})

	t.Run("update to duplicate description rejected", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating first category: %v", err)
		}
		transport, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Transport"})
		if err != nil {
			t.Fatalf("unexpected error creating second category: %v", err)
		}

		_, err = svc.UpdateCategory(context.Background(), transport.ID, CreateCategoryInput{Description: "Food"})
		if err == nil {
			t.Fatal("expected error for duplicate description on update")
		}

		// Updating a category to keep its own description should still succeed.
		_, err = svc.UpdateCategory(context.Background(), transport.ID, CreateCategoryInput{Description: "Transport"})
		if err != nil {
			t.Fatalf("unexpected error updating category to its own description: %v", err)
		}
	})

	t.Run("create validation", func(t *testing.T) {
		tests := []struct {
			name  string
			input CreateCategoryInput
		}{
			{name: "blank description", input: CreateCategoryInput{Description: "   "}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := NewInMemoryRepository()
				svc := NewService(repo)

				_, err := svc.CreateCategory(context.Background(), tt.input)
				if err == nil {
					t.Fatal("expected validation error")
				}
			})
		}
	})

	t.Run("list after create", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating first category: %v", err)
		}
		_, err = svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Transport"})
		if err != nil {
			t.Fatalf("unexpected error creating second category: %v", err)
		}

		categories, err := svc.ListCategories(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(categories) != 2 {
			t.Fatalf("expected 2 categories, got %d", len(categories))
		}
	})

	t.Run("get category", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		cat, err := svc.GetCategory(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cat.ID != created.ID {
			t.Fatalf("expected ID %d, got %d", created.ID, cat.ID)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.GetCategory(context.Background(), 0)
		if err == nil {
			t.Fatal("expected error for id=0")
		}
	})

	t.Run("update category", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		updated, err := svc.UpdateCategory(context.Background(), created.ID, CreateCategoryInput{Description: "Groceries"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Description != "Groceries" {
			t.Fatalf("expected description %q, got %q", "Groceries", updated.Description)
		}
	})

	t.Run("update validation", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		_, err = svc.UpdateCategory(context.Background(), created.ID, CreateCategoryInput{Description: ""})
		if err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("delete category", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateCategory(context.Background(), CreateCategoryInput{Description: "Food"})
		if err != nil {
			t.Fatalf("unexpected error creating category: %v", err)
		}

		err = svc.DeleteCategory(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = svc.GetCategory(context.Background(), created.ID)
		if err == nil {
			t.Fatal("expected error after deletion")
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		err := svc.DeleteCategory(context.Background(), -1)
		if err == nil {
			t.Fatal("expected error for negative id")
		}
	})
}
