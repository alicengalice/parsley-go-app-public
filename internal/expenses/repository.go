package expenses

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Expense is a sample domain model local to the expenses package.
// Later, you can move this to a shared domain package if needed.
type Expense struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CategoryID  int       `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// Repository defines storage operations regardless of where data lives
// (in-memory map, SQL database, NoSQL, etc.).
type Repository interface {
	List(ctx context.Context) ([]Expense, error)
	Create(ctx context.Context, expense Expense) (Expense, error)
	Get(ctx context.Context, id int) (Expense, error)
	Update(ctx context.Context, id int, expense Expense) (Expense, error)
	Delete(ctx context.Context, id int) error
}

// InMemoryRepository is a sandbox storage implementation.
// It uses a mutex so concurrent HTTP requests cannot corrupt shared state.
type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int
	items  map[int]Expense
}

// LEARNING: Constructor initializes in-memory state explicitly.
// This gives predictable defaults and avoids nil-map panics.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  make(map[int]Expense),
	}
}

// LEARNING: Repository methods return data plus error as explicit values.
// Go favors returned errors over exceptions for normal control flow.
func (r *InMemoryRepository) List(ctx context.Context) ([]Expense, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Expense, 0, len(r.items))
	// Collect map values into a slice for JSON responses.
	for _, expense := range r.items {
		result = append(result, expense)
	}

	return result, nil
}

// LEARNING: defer guarantees lock release even when returning early.
// This pattern reduces mutex-leak bugs in multi-branch methods.
func (r *InMemoryRepository) Create(ctx context.Context, expense Expense) (Expense, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	expense.ID = r.nextID
	expense.Date = time.Now()
	r.items[expense.ID] = expense
	r.nextID++

	return expense, nil
}

// Get retrieves a single expense by ID.
func (r *InMemoryRepository) Get(ctx context.Context, id int) (Expense, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	expense, exists := r.items[id]
	if !exists {
		return Expense{}, fmt.Errorf("expense not found")
	}
	return expense, nil
}

// Update modifies an existing expense (preserves ID and Date).
func (r *InMemoryRepository) Update(ctx context.Context, id int, expense Expense) (Expense, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.items[id]
	if !exists {
		return Expense{}, fmt.Errorf("expense not found")
	}

	// Preserve server-assigned fields
	expense.ID = existing.ID
	expense.Date = existing.Date

	r.items[id] = expense
	return expense, nil
}

// Delete removes an expense by ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return fmt.Errorf("expense not found")
	}
	delete(r.items, id)
	return nil
}
