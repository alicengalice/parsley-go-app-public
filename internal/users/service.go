package users

import (
	"context"
	"fmt"
	"strings"
)

// Service defines the business use-cases for users.
type Service interface {
	ListUsers(ctx context.Context) ([]User, error)
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	GetUser(ctx context.Context, id int) (User, error)
	UpdateUser(ctx context.Context, id int, input CreateUserInput) (User, error)
	DeleteUser(ctx context.Context, id int) error
}

// CreateUserInput is the input model for create and update operations.
type CreateUserInput struct {
	FirstName string
	LastName  string
	Email     string
}

// service is the concrete implementation of Service.
type service struct {
	repo Repository
}

// NewService returns a Service backed by the given repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListUsers returns all users.
func (s *service) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

// CreateUser creates a new user after validating input.
func (s *service) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	if err := validateCreateOrUpdateInput(input); err != nil {
		return User{}, err
	}

	// Convert input DTO into domain model expected by repository.
	user := User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	// Persist and return the newly created user.
	return s.repo.Create(ctx, user)
}

// GetUser retrieves a single user by ID.
func (s *service) GetUser(ctx context.Context, id int) (User, error) {
	if id <= 0 {
		return User{}, fmt.Errorf("invalid user id")
	}
	return s.repo.Get(ctx, id)
}

// UpdateUser modifies an existing user.
func (s *service) UpdateUser(ctx context.Context, id int, input CreateUserInput) (User, error) {
	if id <= 0 {
		return User{}, fmt.Errorf("invalid user id")
	}
	if err := validateCreateOrUpdateInput(input); err != nil {
		return User{}, err
	}

	user := User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}
	return s.repo.Update(ctx, id, user)
}

// DeleteUser removes a user.
func (s *service) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}
	return s.repo.Delete(ctx, id)
}

func validateCreateOrUpdateInput(input CreateUserInput) error {
	if strings.TrimSpace(input.FirstName) == "" {
		return fmt.Errorf("first_name is required")
	}
	if strings.TrimSpace(input.LastName) == "" {
		return fmt.Errorf("last_name is required")
	}
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(input.Email, "@") {
		return fmt.Errorf("email is invalid")
	}
	return nil
}
