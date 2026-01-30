package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		name := os.Getenv("DB_NAME")
		password := os.Getenv("DB_PASSWORD")

		if host != "" && port != "" && user != "" && name != "" && password != "" {
			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, url.QueryEscape(password), host, port, name)
		} else {
			log.Fatal("DATABASE_URL is not set and DB_* variables are missing")
		}
	}

	// Inject password if present (from Secrets Manager)
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword != "" {
		databaseURL = strings.Replace(databaseURL, "PLACEHOLDER", url.QueryEscape(dbPassword), 1)
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Create driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Failed to create migration driver:", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Failed to create migration instance:", err)
	}

	// Get command
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [up|down|version]")
	}

	command := os.Args[1]

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to run migrations:", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to rollback migrations:", err)
		}
		fmt.Println("Migrations rolled back successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Failed to get migration version:", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	default:
		log.Fatal("Unknown command. Use: up, down, or version")
	}
}
