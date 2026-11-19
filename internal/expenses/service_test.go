package expenses

import (
	"context"
	"testing"
)

func TestExpensesService(t *testing.T) {
	t.Run("list empty", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		expenses, err := svc.ListExpenses(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(expenses) != 0 {
			t.Fatalf("expected 0 expenses, got %d", len(expenses))
		}
	})

	t.Run("create expense", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		input := CreateExpenseInput{UserID: 1, CategoryID: 2, Amount: 50.75, Description: "Lunch"}
		expense, err := svc.CreateExpense(context.Background(), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expense.ID == 0 {
			t.Fatal("expected non-zero ID")
		}
		if expense.Date.IsZero() {
			t.Fatal("expected non-zero date")
		}
		if expense.UserID != input.UserID {
			t.Fatalf("expected user_id %d, got %d", input.UserID, expense.UserID)
		}
		if expense.Amount != input.Amount {
			t.Fatalf("expected amount %f, got %f", input.Amount, expense.Amount)
		}
	})

	t.Run("create and list", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		input1 := CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Item 1"}
		input2 := CreateExpenseInput{UserID: 2, CategoryID: 2, Amount: 20.0, Description: "Item 2"}

		_, err := svc.CreateExpense(context.Background(), input1)
		if err != nil {
			t.Fatalf("unexpected error creating first expense: %v", err)
		}
		_, err = svc.CreateExpense(context.Background(), input2)
		if err != nil {
			t.Fatalf("unexpected error creating second expense: %v", err)
		}

		expenses, err := svc.ListExpenses(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(expenses) != 2 {
			t.Fatalf("expected 2 expenses, got %d", len(expenses))
		}
	})

	t.Run("create id increments", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		input := CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 5.0, Description: "Test"}
		exp1, _ := svc.CreateExpense(context.Background(), input)
		exp2, _ := svc.CreateExpense(context.Background(), input)
		exp3, _ := svc.CreateExpense(context.Background(), input)

		if exp1.ID != 1 || exp2.ID != 2 || exp3.ID != 3 {
			t.Fatalf("expected IDs 1,2,3 got %d,%d,%d", exp1.ID, exp2.ID, exp3.ID)
		}
	})

	t.Run("create validation", func(t *testing.T) {
		tests := []struct {
			name  string
			input CreateExpenseInput
		}{
			{name: "missing user_id", input: CreateExpenseInput{UserID: 0, CategoryID: 1, Amount: 10.0, Description: "Test"}},
			{name: "missing category_id", input: CreateExpenseInput{UserID: 1, CategoryID: 0, Amount: 10.0, Description: "Test"}},
			{name: "zero amount", input: CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 0, Description: "Test"}},
			{name: "missing description", input: CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: ""}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := NewInMemoryRepository()
				svc := NewService(repo)

				_, err := svc.CreateExpense(context.Background(), tt.input)
				if err == nil {
					t.Fatal("expected validation error")
				}
			})
		}
	})

	t.Run("get expense", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Test"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		expense, err := svc.GetExpense(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expense.ID != created.ID {
			t.Fatalf("expected ID %d, got %d", created.ID, expense.ID)
		}
	})

	t.Run("get invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		_, err := svc.GetExpense(context.Background(), 0)
		if err == nil {
			t.Fatal("expected error for id=0")
		}
	})

	t.Run("update expense", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Lunch"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		updated, err := svc.UpdateExpense(context.Background(), created.ID, CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 20.0, Description: "Dinner"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Description != "Dinner" {
			t.Fatalf("expected description %q, got %q", "Dinner", updated.Description)
		}
		if updated.Amount != 20.0 {
			t.Fatalf("expected amount 20.0, got %f", updated.Amount)
		}
	})

	t.Run("delete expense", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		created, err := svc.CreateExpense(context.Background(), CreateExpenseInput{UserID: 1, CategoryID: 1, Amount: 10.0, Description: "Test"})
		if err != nil {
			t.Fatalf("unexpected error creating expense: %v", err)
		}

		err = svc.DeleteExpense(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = svc.GetExpense(context.Background(), created.ID)
		if err == nil {
			t.Fatal("expected error after deletion")
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		repo := NewInMemoryRepository()
		svc := NewService(repo)

		err := svc.DeleteExpense(context.Background(), -1)
		if err == nil {
			t.Fatal("expected error for negative id")
		}
	})
}
