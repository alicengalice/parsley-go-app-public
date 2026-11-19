package categories

import (
	"context"
	"fmt"
	"strings"
)

// Service defines the business use-cases for categories.
type Service interface {
	ListCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error)
	GetCategory(ctx context.Context, id int) (Category, error)
	UpdateCategory(ctx context.Context, id int, input CreateCategoryInput) (Category, error)
	DeleteCategory(ctx context.Context, id int) error
}

// CreateCategoryInput is the input model for create and update operations.
type CreateCategoryInput struct {
	Description string
}

// service is the concrete implementation of Service.
type service struct {
	repo Repository
}

// NewService returns a Service backed by the given repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListCategories returns all categories.
func (s *service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.List(ctx)
}

// CreateCategory creates a new category after validating input.
func (s *service) CreateCategory(ctx context.Context, input CreateCategoryInput) (Category, error) {
	if err := validateCreateOrUpdateInput(input); err != nil {
		return Category{}, err
	}

	cat := Category{
		Description: input.Description,
	}
	return s.repo.Create(ctx, cat)
}

// GetCategory retrieves a single category by ID.
func (s *service) GetCategory(ctx context.Context, id int) (Category, error) {
	if id <= 0 {
		return Category{}, fmt.Errorf("invalid category id")
	}
	return s.repo.Get(ctx, id)
}

// UpdateCategory modifies an existing category.
func (s *service) UpdateCategory(ctx context.Context, id int, input CreateCategoryInput) (Category, error) {
	if id <= 0 {
		return Category{}, fmt.Errorf("invalid category id")
	}
	if err := validateCreateOrUpdateInput(input); err != nil {
		return Category{}, err
	}

	cat := Category{
		Description: input.Description,
	}
	return s.repo.Update(ctx, id, cat)
}

// DeleteCategory removes a category.
func (s *service) DeleteCategory(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid category id")
	}
	return s.repo.Delete(ctx, id)
}

func validateCreateOrUpdateInput(input CreateCategoryInput) error {
	if strings.TrimSpace(input.Description) == "" {
		return fmt.Errorf("description is required")
	}
	return nil
}
