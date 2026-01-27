package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

var migrationsDir string

func main() {
	var (
		dbURL   = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		command = flag.String("cmd", "up", "Command: up, down, status")
		migDir  = flag.String("migrations", "", "Migrations directory")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	if *migDir == "" {
		*migDir = os.Getenv("MIGRATIONS_DIR")
	}
	if *migDir == "" {
		exe, err := os.Executable()
		if err == nil {
			*migDir = filepath.Join(filepath.Dir(exe), "..", "..", "migrations")
		}
	}
	if *migDir == "" {
		*migDir = "./migrations"
	}
	migrationsDir = *migDir

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := ensureMigrationsTable(db); err != nil {
		log.Fatalf("failed to create migrations table: %v", err)
	}

	switch *command {
	case "up":
		if err := migrateUp(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
	case "status":
		if err := showStatus(db); err != nil {
			log.Fatalf("status failed: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", *command)
	}
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func getMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func migrateUp(db *sql.DB) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	files, err := getMigrationFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		log.Printf("applied: %s", version)
	}

	log.Println("migrations complete")
	return nil
}

func migrateDown(db *sql.DB) error {
	var lastVersion string
	err := db.QueryRow("SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1").Scan(&lastVersion)
	if err == sql.ErrNoRows {
		log.Println("no migrations to rollback")
		return nil
	}
	if err != nil {
		return err
	}

	downFile := lastVersion + ".down.sql"
	content, err := os.ReadFile(filepath.Join(migrationsDir, downFile))
	if err != nil {
		return fmt.Errorf("read %s: %w", downFile, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		return fmt.Errorf("execute %s: %w", downFile, err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", lastVersion); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("rolled back: %s", lastVersion)
	return nil
}

func showStatus(db *sql.DB) error {
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	files, err := getMigrationFiles()
	if err != nil {
		return err
	}

	fmt.Println("Migration Status:")
	fmt.Println("-----------------")
	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")
		status := "pending"
		if applied[version] {
			status = "applied"
		}
		fmt.Printf("[%s] %s\n", status, version)
	}
	return nil
}
