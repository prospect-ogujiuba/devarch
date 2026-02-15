//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/priz/devarch-api/internal/api"
	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/security"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB             *sql.DB
	postgresContainer  testcontainers.Container
	testConnStr        string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start Postgres 16-alpine via testcontainers
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("devarch_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start postgres container: %v", err)
	}
	postgresContainer = container

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to get connection string: %v", err)
	}
	testConnStr = connStr

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	testDB = db

	// Run migrations
	migrateUp(testDB, "../../migrations")

	// Cleanup on exit
	defer func() {
		if err := testDB.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}()

	// Run tests
	os.Exit(m.Run())
}

func migrateUp(db *sql.DB, dir string) {
	// Create schema_migrations table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Find migration files
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		log.Fatalf("Failed to glob migration files: %v", err)
	}

	// Sort files
	sort.Strings(files)

	// Apply each migration
	for _, file := range files {
		// Extract version from filename (e.g., "001_name.up.sql" -> 1)
		base := filepath.Base(file)
		parts := strings.Split(base, "_")
		if len(parts) == 0 {
			log.Fatalf("Invalid migration filename: %s", file)
		}

		var version int
		_, err := fmt.Sscanf(parts[0], "%d", &version)
		if err != nil {
			log.Fatalf("Failed to parse version from %s: %v", file, err)
		}

		// Check if already applied
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check migration status for version %d: %v", version, err)
		}

		if exists {
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Execute migration
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Failed to execute migration %s: %v", file, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			log.Fatalf("Failed to record migration %d: %v", version, err)
		}

		log.Printf("Applied migration %s", base)
	}
}

func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	tables := []string{
		// Instance overrides and config (children first)
		"instance_ports",
		"instance_volumes",
		"instance_env_vars",
		"instance_env_files",
		"instance_labels",
		"instance_domains",
		"instance_healthchecks",
		"instance_config_files",
		"instance_dependencies",
		"instance_resource_limits",
		"instance_networks",
		"instance_config_mounts",

		// Service instance wires
		"service_instance_wires",

		// Service instances
		"service_instances",

		// Stacks
		"stacks",

		// Projects
		"projects",

		// Service relationships
		"service_config_mounts",
		"service_config_versions",
		"service_config_files",
		"service_exports",
		"service_import_contracts",
		"service_ports",
		"service_volumes",
		"service_env_vars",
		"service_env_files",
		"service_networks",
		"service_dependencies",
		"service_labels",
		"service_domains",
		"service_healthchecks",

		// Services
		"services",

		// Categories
		"categories",

		// Container state tracking
		"container_metrics",
		"container_states",
		"sync_state",
		"sync_jobs",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(t, err, "failed to truncate %s", table)
	}
}

func setupRouter(t *testing.T) http.Handler {
	t.Helper()
	stubClient := &container.Client{}
	return api.NewRouter(testDB, stubClient, nil, nil, nil, nil, nil, nil, nil, security.ModeDevOpen, slog.Default())
}

func createStackViaDB(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	var id int
	err := db.QueryRow(`
		INSERT INTO stacks (name, description, network_name, enabled)
		VALUES ($1, $2, $3, true)
		RETURNING id
	`, name, "Test stack: "+name, "devarch-"+name+"-net").Scan(&id)
	require.NoError(t, err)
	return id
}

func seedServiceTemplate(t *testing.T, db *sql.DB, serviceName string) {
	t.Helper()
	// Ensure category exists
	_, err := db.Exec(`INSERT INTO categories (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, "test-cat")
	require.NoError(t, err)
	// Ensure service exists
	_, err = db.Exec(`
		INSERT INTO services (name, category_id, image_name)
		SELECT $1, id, 'test-image:latest'
		FROM categories WHERE name = 'test-cat'
		ON CONFLICT (name) DO NOTHING
	`, serviceName)
	require.NoError(t, err)
}

func createInstanceViaDB(t *testing.T, db *sql.DB, stackID int, serviceName string) string {
	t.Helper()
	seedServiceTemplate(t, db, serviceName)

	// Get the service ID for template_service_id
	var serviceID int
	err := db.QueryRow("SELECT id FROM services WHERE name = $1", serviceName).Scan(&serviceID)
	require.NoError(t, err)

	instanceID := fmt.Sprintf("%s-%s-1", serviceName, "test")
	_, err = db.Exec(`
		INSERT INTO service_instances (stack_id, instance_id, template_service_id)
		VALUES ($1, $2, $3)
	`, stackID, instanceID, serviceID)
	require.NoError(t, err)
	return instanceID
}
