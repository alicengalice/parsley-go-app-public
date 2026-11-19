package expenses

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var ErrInvalidExpenseReference = errors.New("user_id or category_id does not reference an existing record")

// SQLiteRepository persists expenses in a SQLite database.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository prepares the schema and returns a SQLite-backed repository.
func NewSQLiteRepository(ctx context.Context, db *sql.DB) (*SQLiteRepository, error) {
	r := &SQLiteRepository{db: db}
	if err := r.ensureSchema(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

// LEARNING: Schema-First Repository Startup
// The repository ensures the required table exists before serving requests.
// This keeps local/dev startup simple because baseline schema is bootstrapped lazily.
func (r *SQLiteRepository) ensureSchema(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS expenses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id),
	category_id INTEGER NOT NULL REFERENCES categories(id),
	amount REAL NOT NULL,
	description TEXT NOT NULL,
	date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create expenses schema: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]Expense, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, category_id, amount, description, date
FROM expenses
ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	result := make([]Expense, 0)
	// LEARNING: Row Scanning Pattern
	// Iterate through SQL rows and scan each row into a domain model,
	// then append it to a slice returned to service/handler layers.
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.ID, &expense.UserID, &expense.CategoryID, &expense.Amount, &expense.Description, &expense.Date); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		result = append(result, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}

	return result, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, expense Expense) (Expense, error) {
	res, err := r.db.ExecContext(ctx, `
INSERT INTO expenses (user_id, category_id, amount, description)
VALUES (?, ?, ?, ?)`,
		expense.UserID,
		expense.CategoryID,
		expense.Amount,
		expense.Description,
	)
	if err != nil {
		return Expense{}, fmt.Errorf("create expense: %w", translateConstraintError(err))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Expense{}, fmt.Errorf("get inserted id: %w", err)
	}

	// LEARNING: Create-then-read keeps response shape consistent with Get/List.
	// It avoids duplicating scan/transform logic after insert operations.
	return r.Get(ctx, int(id))
}

func (r *SQLiteRepository) Get(ctx context.Context, id int) (Expense, error) {
	var expense Expense
	err := r.db.QueryRowContext(ctx, `
SELECT id, user_id, category_id, amount, description, date
FROM expenses
WHERE id = ?`, id).Scan(&expense.ID, &expense.UserID, &expense.CategoryID, &expense.Amount, &expense.Description, &expense.Date)
	if err != nil {
		if err == sql.ErrNoRows {
			return Expense{}, fmt.Errorf("expense not found")
		}
		return Expense{}, fmt.Errorf("get expense: %w", err)
	}
	return expense, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, id int, expense Expense) (Expense, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE expenses
SET user_id = ?, category_id = ?, amount = ?, description = ?
WHERE id = ?`,
		expense.UserID,
		expense.CategoryID,
		expense.Amount,
		expense.Description,
		id,
	)
	if err != nil {
		return Expense{}, fmt.Errorf("update expense: %w", translateConstraintError(err))
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Expense{}, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return Expense{}, fmt.Errorf("expense not found")
	}

	updated, err := r.Get(ctx, id)
	if err != nil {
		return Expense{}, err
	}

	if updated.Date.IsZero() {
		updated.Date = time.Now()
	}

	return updated, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

func translateConstraintError(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY {
		return ErrInvalidExpenseReference
	}
	return err
}
