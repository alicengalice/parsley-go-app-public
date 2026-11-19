# Parsley Getting Started 🌿

This note is a short, repeatable guide for getting the Parsley backend running locally and checking the main API endpoints.

## 1) What is running today 🚀

The app serves a small expense API on port 8081 with these routes:
- `GET /health`
- `GET /users/`
- `POST /users/`
- `GET /users/{id}`
- `PUT /users/{id}`
- `DELETE /users/{id}`
- `GET /categories/`
- `POST /categories/`
- `GET /categories/{id}`
- `PUT /categories/{id}`
- `DELETE /categories/{id}`
- `GET /expenses/`
- `POST /expenses/`
- `GET /expenses/{id}`
- `PUT /expenses/{id}`
- `DELETE /expenses/{id}`

The server is wired in [main.go](main.go) and uses SQLite by default. The database file is [parsley.db](parsley.db) in the project root unless you override it with `PARSLEY_DB_DSN`.

The embedded seed script, [seed/seed.sql](seed/seed.sql), initializes an empty database with sample users, categories, and expenses.

## 2) Start the server ⚡

### Prerequisites
- Go 1.25.0 or later
- PowerShell or another terminal

### Run it
```powershell
cd c:\dev\Projects\parsley
go run .
```

A startup line similar to the following appears:
```text
🌱 Parsley server is warming up on http://localhost:8081...
```

## 3) Quick API check ✅

Only the health check is needed for a fast sanity check. The other curl examples are optional and can be skipped.

### Health check
```powershell
curl.exe http://localhost:8081/health
```

Expected response:
```json
{
  "status": "UP",
  "application": "Parsley Backend"
}
```

### Optional: inspect the database with sqlite3
Use the SQLite file for direct inspection with sqlite3 instead of the API calls.

List the tables:
```powershell
sqlite3 parsley.db ".tables"
```

Count rows in each table:
```powershell
sqlite3 parsley.db "SELECT 'expenses' AS table_name, COUNT(*) AS row_count FROM expenses UNION ALL SELECT 'categories', COUNT(*) FROM categories UNION ALL SELECT 'users', COUNT(*) FROM users;"
```

## 4) Run the tests 🧪

```powershell
go test ./...
```

Focused tests for the expenses package:
```powershell
go test ./internal/expenses/...
```

Verbose mode:
```powershell
go test -v ./internal/expenses/...
```

Coverage:
```powershell
go test -cover ./internal/expenses/...
```

## 5) Runtime config ⚙️

The app uses `parsley.db` in the current working directory by default. `PARSLEY_DB_DSN` lets you choose a different SQLite database file or connection string at startup.

If no value is set, the app uses `parsley.db` in the current working directory. In the normal workflow, that means the file at [parsley.db](parsley.db) in the project root.

Set a custom database file with:
```powershell
$env:PARSLEY_DB_DSN = "parsley-dev.db"
go run .
```

## 6) Project structure 📦

```text
parsley/
├── main.go
├── README.md
├── parsley-start.md
├── seed/
│   └── seed.sql
└── internal/
  ├── users/
  ├── categories/
  ├── expenses/
  └── platform/
    └── db/
```

## 7) Quick troubleshooting 🛠️

### Port 8081 is already in use
```powershell
netstat -ano | findstr :8081
taskkill /PID <PID> /F
```

### Connection refused
- The server must be running.
- The request must target `http://localhost:8081`.

### PowerShell `curl` issues
Use `curl.exe` for more predictable behavior in Windows PowerShell.

### Invalid request body
- The JSON must be valid.
- `Content-Type: application/json` must be included.

If something fails unexpectedly, start by running:
```powershell
go test ./...
```

---

Happy testing! 🌱

For the full API reference and current project status, see [README.md](README.md).