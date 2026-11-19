package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// User is a sample domain model local to the users package.
// Later, you can move this to a shared domain package if needed.
type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrEmailAlreadyInUse = errors.New("email already in use")
	ErrUserHasExpenses   = errors.New("cannot delete user with expenses")
)

// Repository defines storage operations regardless of where data lives
// (in-memory map, SQL database, NoSQL, etc.).
type Repository interface {
	List(ctx context.Context) ([]User, error)
	Create(ctx context.Context, user User) (User, error)
	Get(ctx context.Context, id int) (User, error)
	Update(ctx context.Context, id int, user User) (User, error)
	Delete(ctx context.Context, id int) error
}

// InMemoryRepository is a sandbox storage implementation.
// It uses a mutex so concurrent HTTP requests cannot corrupt shared state.
type InMemoryRepository struct {
	mu     sync.Mutex
	nextID int
	items  map[int]User
}

// NewInMemoryRepository initializes in-memory storage for users.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		nextID: 1,
		items:  make(map[int]User),
	}
}

// List returns all users currently stored in memory.
func (r *InMemoryRepository) List(ctx context.Context) ([]User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]User, 0, len(r.items))
	// Collect map values into a slice for JSON responses.
	for _, user := range r.items {
		result = append(result, user)
	}

	return result, nil
}

// Create stores a new user and assigns server-managed fields.
func (r *InMemoryRepository) Create(ctx context.Context, user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.items {
		if strings.EqualFold(existing.Email, user.Email) {
			return User{}, ErrEmailAlreadyInUse
		}
	}

	user.ID = r.nextID
	user.CreatedAt = time.Now()
	r.items[user.ID] = user
	r.nextID++

	return user, nil
}

// Get retrieves a single user by ID.
func (r *InMemoryRepository) Get(ctx context.Context, id int) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.items[id]
	if !exists {
		return User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

// Update modifies an existing user (preserves ID and CreatedAt).
func (r *InMemoryRepository) Update(ctx context.Context, id int, user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.items[id]
	if !exists {
		return User{}, fmt.Errorf("user not found")
	}
	for existingID, existing := range r.items {
		if existingID != id && strings.EqualFold(existing.Email, user.Email) {
			return User{}, ErrEmailAlreadyInUse
		}
	}

	// Preserve server-assigned fields
	user.ID = existing.ID
	user.CreatedAt = existing.CreatedAt

	r.items[id] = user
	return user, nil
}

// Delete removes a user by ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return fmt.Errorf("user not found")
	}
	delete(r.items, id)
	return nil
}
