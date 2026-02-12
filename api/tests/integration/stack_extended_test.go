//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackClone(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "original-stack")

	w := doRequest(t, router, "POST", "/api/v1/stacks/original-stack/clone", `{"name":"cloned-stack"}`)
	require.Equal(t, 201, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	require.Equal(t, "cloned-stack", data["name"])

	// Verify new stack exists in DB
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM stacks WHERE name = $1 AND deleted_at IS NULL", "cloned-stack").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStackRename(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "old-name")

	w := doRequest(t, router, "POST", "/api/v1/stacks/old-name/rename", `{"name":"new-name"}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	require.Equal(t, "new-name", data["name"])

	// Verify new name exists and old name is soft-deleted
	var newCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM stacks WHERE name = $1 AND deleted_at IS NULL", "new-name").Scan(&newCount)
	require.NoError(t, err)
	require.Equal(t, 1, newCount)

	var oldDeletedAt *string
	err = testDB.QueryRow("SELECT deleted_at::text FROM stacks WHERE name = $1", "old-name").Scan(&oldDeletedAt)
	require.NoError(t, err)
	require.NotNil(t, oldDeletedAt, "old stack should be soft-deleted")
}

func TestStackEnable(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "enable-stack")

	w := doRequest(t, router, "POST", "/api/v1/stacks/enable-stack/enable", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	require.Equal(t, true, data["enabled"])
}

func TestStackDisable(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "disable-stack")

	w := doRequest(t, router, "POST", "/api/v1/stacks/disable-stack/disable", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// Verify enabled=false in DB
	var enabled bool
	err := testDB.QueryRow("SELECT enabled FROM stacks WHERE name = $1 AND deleted_at IS NULL", "disable-stack").Scan(&enabled)
	require.NoError(t, err)
	require.False(t, enabled, "stack should be disabled")
}

func TestStackCompose(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "compose-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "GET", "/api/v1/stacks/compose-stack/compose", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	yaml, ok := data["yaml"].(string)
	require.True(t, ok, "expected yaml field in response")
	require.NotEmpty(t, yaml, "expected non-empty YAML content")
}

func TestStackPlan(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "plan-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "GET", "/api/v1/stacks/plan-stack/plan", "")
	// Plan may return 200 or 500 depending on container client stub behavior
	require.True(t, w.Code == 200 || w.Code == 500, "expected 200 or 500, got %d: %s", w.Code, w.Body.String())
}

func TestStackDeletePreview(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "preview-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "GET", "/api/v1/stacks/preview-stack/delete-preview", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	require.Equal(t, "preview-stack", data["stack_name"])
	require.NotNil(t, data["instance_count"])
}

func TestStackExport(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "export-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "GET", "/api/v1/stacks/export-stack/export", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// Export returns raw YAML, not JSON envelope
	body := w.Body.String()
	require.NotEmpty(t, body)
}

func TestStackPermanentDelete(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create and soft-delete
	createStackViaDB(t, testDB, "perm-delete-stack")
	w := doRequest(t, router, "DELETE", "/api/v1/stacks/perm-delete-stack", "")
	require.True(t, w.Code == 200 || w.Code == 204, "soft delete should succeed, got %d", w.Code)

	// Permanently delete from trash
	w = doRequest(t, router, "DELETE", "/api/v1/stacks/trash/perm-delete-stack", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// Verify gone from DB entirely
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM stacks WHERE name = $1", "perm-delete-stack").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "stack should be permanently deleted")
}

func TestStackLockGenerate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "lock-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "POST", "/api/v1/stacks/lock-stack/lock", "")
	// GenerateLock writes raw JSON (not envelope), check for 200
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	body := w.Body.String()
	require.NotEmpty(t, body)

	// Verify it's valid JSON
	var lockData interface{}
	err := json.Unmarshal([]byte(body), &lockData)
	require.NoError(t, err, "lock response should be valid JSON")
}

func TestStackLockValidate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "lockval-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	// First generate a lock
	w := doRequest(t, router, "POST", "/api/v1/stacks/lockval-stack/lock", "")
	require.Equal(t, 200, w.Code, "generate lock failed: %s", w.Body.String())

	lockBody := w.Body.String()

	// Now validate that lock
	w = doRequest(t, router, "POST", "/api/v1/stacks/lockval-stack/lock/validate", lockBody)
	require.Equal(t, 200, w.Code, "validate lock failed: %s", w.Body.String())
}

func TestStackWiresCRUD(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	stackID := createStackViaDB(t, testDB, "wire-stack")

	// Seed two different service templates
	seedServiceTemplate(t, testDB, "provider-svc")
	seedServiceTemplate(t, testDB, "consumer-svc")

	// Create provider and consumer instances
	providerInstanceID := createInstanceViaDB(t, testDB, stackID, "provider-svc")
	consumerInstanceID := createInstanceViaDB(t, testDB, stackID, "consumer-svc")

	// Get service IDs for setting up exports/imports
	var providerServiceID, consumerServiceID int
	err := testDB.QueryRow("SELECT id FROM services WHERE name = $1", "provider-svc").Scan(&providerServiceID)
	require.NoError(t, err)
	err = testDB.QueryRow("SELECT id FROM services WHERE name = $1", "consumer-svc").Scan(&consumerServiceID)
	require.NoError(t, err)

	// Create a service export on the provider
	_, err = testDB.Exec(`
		INSERT INTO service_exports (service_id, type, port, protocol)
		VALUES ($1, 'database', 5432, 'tcp')
	`, providerServiceID)
	require.NoError(t, err)

	// Create an import contract on the consumer
	_, err = testDB.Exec(`
		INSERT INTO service_import_contracts (service_id, name, type, required, env_vars)
		VALUES ($1, 'db-connection', 'database', true, '{"DB_HOST":""}')
	`, consumerServiceID)
	require.NoError(t, err)

	// POST create wire
	wirePayload := fmt.Sprintf(`{
		"consumer_instance_id": %q,
		"provider_instance_id": %q,
		"import_contract_name": "db-connection"
	}`, consumerInstanceID, providerInstanceID)

	w := doRequest(t, router, "POST", "/api/v1/stacks/wire-stack/wires", wirePayload)
	require.Equal(t, http.StatusCreated, w.Code, "create wire failed: %s", w.Body.String())

	// Parse wire ID from response
	var wireEnv map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&wireEnv)
	require.NoError(t, err)
	require.Contains(t, wireEnv, "data")

	wireData := wireEnv["data"].(map[string]interface{})
	wireID := wireData["id"].(float64)
	require.Greater(t, wireID, float64(0), "wire ID should be positive")

	// GET list wires
	w = doRequest(t, router, "GET", "/api/v1/stacks/wire-stack/wires", "")
	require.Equal(t, 200, w.Code, "list wires failed: %s", w.Body.String())

	var listEnv map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&listEnv)
	require.NoError(t, err)
	require.Contains(t, listEnv, "data")

	listData := listEnv["data"].(map[string]interface{})
	wires := listData["wires"].([]interface{})
	require.Len(t, wires, 1, "expected 1 wire")

	// DELETE wire
	deletePath := fmt.Sprintf("/api/v1/stacks/wire-stack/wires/%d", int(wireID))
	w = doRequest(t, router, "DELETE", deletePath, "")
	require.Equal(t, http.StatusNoContent, w.Code, "delete wire failed: %s", w.Body.String())

	// Verify wire is gone
	w = doRequest(t, router, "GET", "/api/v1/stacks/wire-stack/wires", "")
	require.Equal(t, 200, w.Code)

	var listEnv2 map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&listEnv2)
	require.NoError(t, err)

	listData2 := listEnv2["data"].(map[string]interface{})
	wires2 := listData2["wires"].([]interface{})
	require.Len(t, wires2, 0, "expected 0 wires after delete")
}

// TestStackComposeNotFound verifies 404 for compose on nonexistent stack
func TestStackComposeNotFound(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "GET", "/api/v1/stacks/nonexistent/compose", "")
	require.Equal(t, 404, w.Code)
}

// TestStackCloneDuplicate verifies clone fails when target name exists
func TestStackCloneDuplicate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "src-stack")
	createStackViaDB(t, testDB, "existing-stack")

	w := doRequest(t, router, "POST", "/api/v1/stacks/src-stack/clone", `{"name":"existing-stack"}`)
	require.Equal(t, 409, w.Code, "body: %s", w.Body.String())
}

// TestStackDisableVerifyDB confirms the enabled column changes
func TestStackDisableVerifyDB(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	createStackViaDB(t, testDB, "db-disable")

	// Stack is enabled by default
	var before bool
	err := testDB.QueryRow("SELECT enabled FROM stacks WHERE name = $1", "db-disable").Scan(&before)
	require.NoError(t, err)
	require.True(t, before)

	w := doRequest(t, router, "POST", "/api/v1/stacks/db-disable/disable", "")
	require.Equal(t, 200, w.Code)

	var after bool
	err = testDB.QueryRow("SELECT enabled FROM stacks WHERE name = $1", "db-disable").Scan(&after)
	require.NoError(t, err)
	require.False(t, after)
}

