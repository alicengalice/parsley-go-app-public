# Parsley

Parsley is a local-first personal finance app focused on simple expense tracking and calm money habits. The project currently contains a verified Go REST backend, SQLite persistence, and a prototype web dashboard in the `web/` folder.

## Current status

- Backend API is implemented and working
- SQLite database is in use for persisted local data
- Tables for users, categories, and expenses are seeded automatically on first run
- Case-insensitive user emails and category descriptions are unique
- Expenses must reference an existing user and category
- All Go tests currently pass with `go test ./...`
- A first-pass web UI prototype has been created in `web/`

## Tech stack

- Go
- SQLite
- Standard library HTTP server
- Static HTML/CSS/JS web prototype

## Quick start

```bash
# Run the API server
go run .
```

The backend serves on:

- http://localhost:8081

### Health check

```bash
curl http://localhost:8081/health
```

## API overview

### Users
- `GET /users` — list users
- `POST /users` — create a user
- `GET /users/{id}` — get user by ID
- `PUT /users/{id}` — update user
- `DELETE /users/{id}` — delete user

### Categories
- `GET /categories` — list categories
- `POST /categories` — create category
- `GET /categories/{id}` — get category by ID
- `PUT /categories/{id}` — update category
- `DELETE /categories/{id}` — delete category

### Expenses
- `GET /expenses` — list expenses
- `POST /expenses` — create expense
- `GET /expenses/{id}` — get expense by ID
- `PUT /expenses/{id}` — update expense
- `DELETE /expenses/{id}` — delete expense

### Error responses

- `400 Bad Request` — malformed input or failed validation
- `404 Not Found` — requested record does not exist
- `409 Conflict` — duplicate email/category description, or deleting a user/category that still has expenses
- `422 Unprocessable Entity` — an expense refers to a user or category that does not exist

## Database

The app uses SQLite for local development and stores data in `parsley.db` at the project root.

On startup, the app:
1. opens the database
2. creates tables if needed
3. runs the seed script if the database is empty

SQLite enforces case-insensitive uniqueness for user emails and category descriptions. It also enforces foreign-key references from expenses to users and categories. A user or category with existing expenses cannot be deleted.

The sample seed data lives in:

- `seed/seed.sql`

## Web prototype

A lightweight dashboard prototype is available in:

- `web/index.html`
- `web/styles.css`
- `web/script.js`

This version is intentionally a static, first-pass UI that matches the product direction: soft pastel styling, rounded cards, and a friendly money dashboard feel.

The Go server does not serve the `web/` directory yet, so it is not wired to the API.

## Project structure

```text
.
├── internal/
│   ├── categories/
│   ├── expenses/
│   ├── platform/
│   └── users/
├── seed/
│   └── seed.sql
├── web/
│   ├── index.html
│   ├── styles.css
│   └── script.js
├── main.go
├── models.go
├── README.md
├── parsley-api-test.http
├── parsley.db
├── go.mod
└── go.sum
```

## Testing

```bash
# run all tests
go test ./...

# run tests with coverage
go test -cover ./...
```

Currently passing packages include:
- `internal/categories`
- `internal/expenses`
- `internal/users`

## Development notes

The backend follows a clean layered structure:

- Handler → Service → Repository
- SQLite repository implementations live under the internal packages
- Services validate request input; repositories enforce storage invariants atomically
- The API is designed to be easy to extend for future UI integration

## Next direction

The project is currently positioned as a web-first application with a responsive UI. The recommended next iteration is to serve the dashboard from the Go server, load real API data, and add the expense-entry flow.

