package categories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Category is the domain model for expense categories.
type Category struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

var (
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryHasExpenses   = errors.New("cannot delete category with expenses")
)

// Repository defines storage operations regardless of where data lives.
type Repository interface {
	List(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, category Category) (Category, error)
	Get(ctx context.Context, id int) (Category, error)
	Update(ctx context.Context, id int, category Category) (Category, error)
	Delete(ctx context.Context, id int) error
}

// InMemoryRepository is a sandbox storage implementation.
// It uses a mutex so concurrent HTTP requests cannot corrupt shared state.
type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int
	items  map[int]Category
}

// NewInMemoryRepository initializes in-memory storage for categories.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  make(map[int]Category),
	}
}

// List returns all categories currently stored in memory.
func (r *InMemoryRepository) List(ctx context.Context) ([]Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Category, 0, len(r.items))
	// Collect map values into a slice for JSON responses.
	for _, category := range r.items {
		result = append(result, category)
	}

	return result, nil
}

// Create stores a new category and assigns a server ID.
func (r *InMemoryRepository) Create(ctx context.Context, category Category) (Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.items {
		if strings.EqualFold(existing.Description, category.Description) {
			return Category{}, ErrCategoryAlreadyExists
		}
	}

	category.ID = r.nextID
	r.items[category.ID] = category
	r.nextID++

	return category, nil
}

// Get retrieves a single category by ID.
func (r *InMemoryRepository) Get(ctx context.Context, id int) (Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	category, exists := r.items[id]
	if !exists {
		return Category{}, fmt.Errorf("category not found")
	}
	return category, nil
}

// Update modifies an existing category.
func (r *InMemoryRepository) Update(ctx context.Context, id int, category Category) (Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.items[id]
	if !exists {
		return Category{}, fmt.Errorf("category not found")
	}
	for existingID, existing := range r.items {
		if existingID != id && strings.EqualFold(existing.Description, category.Description) {
			return Category{}, ErrCategoryAlreadyExists
		}
	}

	// Preserve server-assigned fields
	category.ID = existing.ID

	r.items[id] = category
	return category, nil
}

// Delete removes a category by ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return fmt.Errorf("category not found")
	}
	delete(r.items, id)
	return nil
}
