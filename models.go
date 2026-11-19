package main

import "time"

// ============================================================================
// LEARNING: Go Slices vs Java Lists
// ============================================================================
//
// Java (statically typed ArrayList):
//   List<String> list = new ArrayList<>();
//   list.add("item");
//   list.add("another");
//
// Go (dynamic slice - similar to ArrayList):
//   var list []string                    // declare empty slice of strings
//   list = append(list, "item")          // append adds element, returns new slice
//   list = append(list, "another")
//
// Key differences:
//   - Go slices are more lightweight than Java's ArrayList
//   - append() returns a new slice, so you assign back to the variable
//   - Go handles growing the slice internally (similar to ArrayList.ensureCapacity())
//   - Slices are views into an underlying array; safe for concurrent reads
//
// Example in Parsley context:
//   var expenses []Expense                    // empty slice
//   expenses = append(expenses, newExpense)   // add a new expense
//
// ============================================================================

// User represents an application user account.
// JSON tags define the external API field names.
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Category groups expenses into reporting buckets (food, transport, etc.).
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Expense is the core transaction model for this API.
// Date is server-assigned when the record is created.
type Expense struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CategoryID  int       `json:"category_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}
