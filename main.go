package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alicengalice/parsley/internal/categories"
	"github.com/alicengalice/parsley/internal/expenses"
	"github.com/alicengalice/parsley/internal/platform/db"
	"github.com/alicengalice/parsley/internal/users"
)

//go:embed seed/seed.sql
var seedSQL string

func main() {
	// SUMMARY: What this file does
	// main.go is the composition root — the single place where the whole app is wired together.
	// It owns three responsibilities and nothing else:
	//   1. Configuration  — read DSN from env, fall back to a local default
	//   2. Wiring         — construct each layer (DB → repo → service → handler) and connect them
	//   3. Startup        — seed an empty database, register routes, start the HTTP server
	//
	// No business logic lives here. Every real concern is delegated to its own package:
	//   internal/platform/db   — opens and configures the SQLite connection
	//   internal/*/repository  — reads and writes data
	//   internal/*/service     — enforces business rules
	//   internal/*/handler     — translates HTTP requests into service calls
	//   seed/seed.sql          — deterministic local dataset, embedded at compile time

	// LEARNING: Dependency Injection Chain
	// This is the classic layering pattern:
	//   Repository (data layer) → Service (business logic) → Handler (HTTP layer)
	//
	// Java equivalent would use a DI framework like Spring:
	//   @Bean Repository repo = new InMemoryRepository();
	//   @Bean Service svc = new ServiceImpl(repo);
	//   @Bean Handler handler = new Handler(svc);
	//
	// Go does it explicitly with function calls:
	// 1. Create the repository (storage layer)
	// 2. Create the service (business logic layer) and inject the repository
	// 3. Create the handler (HTTP layer) and inject the service
	// 4. Wire up routes with the handler & register other endpoints (e.g., health check)
	// 5. Start the server

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dsn := os.Getenv("PARSLEY_DB_DSN")
	if dsn == "" {
		dsn = "parsley.db"
	}

	sqliteDB, err := db.OpenSQLite(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer sqliteDB.Close()

	if err := seedIfEmpty(ctx, sqliteDB); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	expensesRepo, err := expenses.NewSQLiteRepository(ctx, sqliteDB)
	if err != nil {
		log.Fatalf("failed to initialize expenses repository: %v", err)
	}
	expensesSvc := expenses.NewService(expensesRepo)
	expensesHandler := expenses.NewHandler(expensesSvc)

	usersRepo, err := users.NewSQLiteRepository(ctx, sqliteDB)
	if err != nil {
		log.Fatalf("failed to initialize users repository: %v", err)
	}
	usersSvc := users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersSvc)

	categoriesRepo, err := categories.NewSQLiteRepository(ctx, sqliteDB)
	if err != nil {
		log.Fatalf("failed to initialize categories repository: %v", err)
	}
	categoriesSvc := categories.NewService(categoriesRepo)
	categoriesHandler := categories.NewHandler(categoriesSvc)

	mux := http.NewServeMux()
	expensesHandler.RegisterRoutes(mux)
	usersHandler.RegisterRoutes(mux)
	categoriesHandler.RegisterRoutes(mux)
	mux.HandleFunc("/health", healthHandler)

	port := ":8081"
	fmt.Printf("🌱 Parsley server is warming up on http://localhost%s...\n", port)

	err = http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatalf("Critical system failure: %v", err)
	}
}

// seedIfEmpty runs seed.sql only when the users table has no rows.
func seedIfEmpty(ctx context.Context, sqliteDB *sql.DB) error {
	var count int
	err := sqliteDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	// table may not exist yet on a brand-new database — seed in that case too
	if err != nil || count == 0 {
		if _, err := sqliteDB.ExecContext(ctx, seedSQL); err != nil {
			return err
		}
		log.Println("database seeded")
	}
	return nil
}

// healthHandler for diagnotic endpoint health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// using inline map since this endpoint is simple
	response := map[string]string{
		"status":      "UP",
		"application": "Parsley Backend",
	}
	json.NewEncoder(w).Encode(response)
}
