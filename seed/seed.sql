-- Parsley local seed dataset
-- Purpose: create baseline tables and load deterministic sample data.

BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    category_id INTEGER NOT NULL REFERENCES categories(id),
    amount REAL NOT NULL,
    description TEXT NOT NULL,
    date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Reset existing rows so each seed run starts from the same state.
DELETE FROM expenses;
DELETE FROM categories;
DELETE FROM users;
DELETE FROM sqlite_sequence WHERE name IN ('expenses', 'categories', 'users');

-- Users
INSERT INTO users (first_name, last_name, email) VALUES
('Alicia', 'Nguyen', 'alicia@parsley-example.com'),
('Marcus', 'Lee', 'marcus@example.com'),
('Priya', 'Shah', 'priya@example.com');

-- Categories
INSERT INTO categories (description) VALUES
('Food'),
('Transport'),
('Housing'),
('Utilities'),
('Health'),
('Leisure');

-- Expenses
INSERT INTO expenses (user_id, category_id, amount, description, date) VALUES
(1, 1, 12.50, 'Lunch bowl', '2026-08-01 12:10:00'),
(1, 2, 18.00, 'Metro pass top-up', '2026-08-01 18:00:00'),
(2, 3, 850.00, 'Monthly rent contribution', '2026-08-02 08:30:00'),
(3, 4, 42.75, 'Electricity bill share', '2026-08-03 20:15:00'),
(2, 5, 27.40, 'Pharmacy purchase', '2026-08-04 09:05:00'),
(3, 6, 34.20, 'Weekend movie and snacks', '2026-08-04 21:10:00');

COMMIT;
