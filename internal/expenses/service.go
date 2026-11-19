package expenses

import (
	"context"
	"fmt"
	"strings"
)

// Service defines the business use-cases for expenses.
// Handlers call this interface instead of talking to storage directly.
type Service interface {
	ListExpenses(ctx context.Context) ([]Expense, error)
	CreateExpense(ctx context.Context, input CreateExpenseInput) (Expense, error)
	GetExpense(ctx context.Context, id int) (Expense, error)
	UpdateExpense(ctx context.Context, id int, input CreateExpenseInput) (Expense, error)
	DeleteExpense(ctx context.Context, id int) error
}

// CreateExpenseInput is an application-level input model for creating expenses.
// It keeps the service contract explicit and independent from HTTP request structs.
type CreateExpenseInput struct {
	UserID      int
	CategoryID  int
	Amount      float64
	Description string
}

// service is the concrete implementation of Service.
// It depends on a Repository for data persistence.
type service struct {
	repo Repository
}

// LEARNING: Constructor returns an interface.
// This keeps callers focused on behavior and makes swapping implementations easier
// (for example, in-memory for tests vs SQLite for runtime).
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// LEARNING: Interfaces are satisfied implicitly in Go.
// No explicit 'implements' keyword is needed; matching method signatures is enough.
func (s *service) ListExpenses(ctx context.Context) ([]Expense, error) {
	return s.repo.List(ctx)
}

// LEARNING: Service input DTO is translated into a domain model.
// This lets HTTP payload shape evolve independently from repository contracts.
func (s *service) CreateExpense(ctx context.Context, input CreateExpenseInput) (Expense, error) {
	if err := validateCreateOrUpdateInput(input); err != nil {
		return Expense{}, err
	}

	// Convert input DTO into domain model expected by repository.
	expense := Expense{
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Description: input.Description,
	}

	// Persist and return the newly created expense.
	return s.repo.Create(ctx, expense)
}

// GetExpense retrieves a single expense by ID.
func (s *service) GetExpense(ctx context.Context, id int) (Expense, error) {
	if id <= 0 {
		return Expense{}, fmt.Errorf("invalid expense id")
	}
	return s.repo.Get(ctx, id)
}

// UpdateExpense modifies an existing expense.
func (s *service) UpdateExpense(ctx context.Context, id int, input CreateExpenseInput) (Expense, error) {
	if id <= 0 {
		return Expense{}, fmt.Errorf("invalid expense id")
	}
	if err := validateCreateOrUpdateInput(input); err != nil {
		return Expense{}, err
	}

	expense := Expense{
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Description: input.Description,
	}
	return s.repo.Update(ctx, id, expense)
}

// DeleteExpense removes an expense.
func (s *service) DeleteExpense(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid expense id")
	}
	return s.repo.Delete(ctx, id)
}

// LEARNING: Validation belongs in the service layer.
// Handlers parse transport input; services enforce domain/business rules.
// That keeps rule behavior consistent no matter who calls the service.
func validateCreateOrUpdateInput(input CreateExpenseInput) error {
	if input.UserID <= 0 {
		return fmt.Errorf("user_id must be greater than zero")
	}
	if input.CategoryID <= 0 {
		return fmt.Errorf("category_id must be greater than zero")
	}
	if input.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if strings.TrimSpace(input.Description) == "" {
		return fmt.Errorf("description is required")
	}
	return nil
}
