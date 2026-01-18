package integration

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	// Use test database
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://madabank:dev_password_change_in_prod@localhost:5432/madabank_test?sslmode=disable"
	}

	testDB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	defer func() {
		_ = testDB.Close()
	}()

	os.Exit(code)
}

// setupTestDB cleans all tables before each test
func setupTestDB(t *testing.T) {
	// Clean all tables before each test
	tables := []string{"audit_logs", "transactions", "cards", "accounts", "users"}
	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

var _ = setupTestDB // Suppress unused warning
