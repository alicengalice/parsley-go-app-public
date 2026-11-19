package categories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// SQLiteRepository persists categories in a SQLite database.
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

// ensureSchema creates the categories table if it does not already exist.
func (r *SQLiteRepository) ensureSchema(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	description TEXT NOT NULL
);`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create categories schema: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS categories_description_unique_ci ON categories (LOWER(description))`); err != nil {
		return fmt.Errorf("create categories description index: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, description
FROM categories
ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	result := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Description); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return result, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, category Category) (Category, error) {
	res, err := r.db.ExecContext(ctx, `
INSERT INTO categories (description)
VALUES (?)`, category.Description)
	if err != nil {
		if isUniqueConstraint(err) {
			return Category{}, ErrCategoryAlreadyExists
		}
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Category{}, fmt.Errorf("get inserted id: %w", err)
	}

	return r.Get(ctx, int(id))
}

func (r *SQLiteRepository) Get(ctx context.Context, id int) (Category, error) {
	var category Category
	err := r.db.QueryRowContext(ctx, `
SELECT id, description
FROM categories
WHERE id = ?`, id).Scan(&category.ID, &category.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return Category{}, fmt.Errorf("category not found")
		}
		return Category{}, fmt.Errorf("get category: %w", err)
	}
	return category, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, id int, category Category) (Category, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE categories
SET description = ?
WHERE id = ?`, category.Description, id)
	if err != nil {
		if isUniqueConstraint(err) {
			return Category{}, ErrCategoryAlreadyExists
		}
		return Category{}, fmt.Errorf("update category: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return Category{}, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return Category{}, fmt.Errorf("category not found")
	}

	return r.Get(ctx, id)
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		if isForeignKeyConstraint(err) {
			return ErrCategoryHasExpenses
		}
		return fmt.Errorf("delete category: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
func isUniqueConstraint(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

func isForeignKeyConstraint(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY
}
