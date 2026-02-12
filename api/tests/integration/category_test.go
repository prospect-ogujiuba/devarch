//go:build integration

package integration_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func seedCategory(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO categories (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, name)
	require.NoError(t, err)
}

func seedServiceInCategory(t *testing.T, db *sql.DB, serviceName, categoryName string) {
	t.Helper()
	seedCategory(t, db, categoryName)
	_, err := db.Exec(`
		INSERT INTO services (name, category_id, image_name)
		SELECT $1, id, 'test-image:latest'
		FROM categories WHERE name = $2
		ON CONFLICT (name) DO NOTHING
	`, serviceName, categoryName)
	require.NoError(t, err)
}

func TestCategoryGet(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedServiceInCategory(t, testDB, "my-svc", "databases")

	w := doRequest(t, router, "GET", "/api/v1/categories/databases", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	dataMap, ok := env["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be object")
	require.Equal(t, "databases", dataMap["name"])
}

func TestCategoryServices(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedServiceInCategory(t, testDB, "postgres", "databases")
	seedServiceInCategory(t, testDB, "mysql", "databases")

	w := doRequest(t, router, "GET", "/api/v1/categories/databases/services", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data to be array")
	require.Equal(t, 2, len(data))
}
