package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// SQLiteRepository persists users in a SQLite database.
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

// ensureSchema creates the users table if it does not already exist.
func (r *SQLiteRepository) ensureSchema(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create users schema: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_ci ON users (LOWER(email))`); err != nil {
		return fmt.Errorf("create users email index: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, first_name, last_name, email, created_at
FROM users
ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	result := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		result = append(result, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return result, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, user User) (User, error) {
	res, err := r.db.ExecContext(ctx, `
INSERT INTO users (first_name, last_name, email)
VALUES (?, ?, ?)`,
		user.FirstName,
		user.LastName,
		user.Email,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return User{}, ErrEmailAlreadyInUse
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("get inserted id: %w", err)
	}

	return r.Get(ctx, int(id))
}

func (r *SQLiteRepository) Get(ctx context.Context, id int) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
SELECT id, first_name, last_name, email, created_at
FROM users
WHERE id = ?`, id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, id int, user User) (User, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE users
SET first_name = ?, last_name = ?, email = ?
WHERE id = ?`,
		user.FirstName,
		user.LastName,
		user.Email,
		id,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			return User{}, ErrEmailAlreadyInUse
		}
		return User{}, fmt.Errorf("update user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return User{}, fmt.Errorf("user not found")
	}

	updated, err := r.Get(ctx, id)
	if err != nil {
		return User{}, err
	}

	return updated, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		if isForeignKeyConstraint(err) {
			return ErrUserHasExpenses
		}
		return fmt.Errorf("delete user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("user not found")
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
