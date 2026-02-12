//go:build integration

package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func doRequest(t *testing.T, router http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestStackCreate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "POST", "/api/v1/stacks", `{"name":"test-stack","description":"A test"}`)
	require.Equal(t, 201, w.Code)

	// Verify envelope
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	// Verify DB row exists
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM stacks WHERE name = $1 AND deleted_at IS NULL", "test-stack").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStackCreateDuplicate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create first stack
	w := doRequest(t, router, "POST", "/api/v1/stacks", `{"name":"test-stack","description":"First"}`)
	require.Equal(t, 201, w.Code)

	// Try to create duplicate
	w = doRequest(t, router, "POST", "/api/v1/stacks", `{"name":"test-stack","description":"Second"}`)
	require.True(t, w.Code == 409 || w.Code == 400 || w.Code == 500, "Expected error status for duplicate, got %d", w.Code)

	// Verify envelope has error
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "error")
}

func TestStackList(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create 2 stacks via DB
	createStackViaDB(t, testDB, "stack-one")
	createStackViaDB(t, testDB, "stack-two")

	w := doRequest(t, router, "GET", "/api/v1/stacks", "")
	require.Equal(t, 200, w.Code)

	// Verify envelope
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	// Verify data is array with 2 items
	data, ok := env["data"].([]interface{})
	require.True(t, ok, "Expected data to be array")
	require.Equal(t, 2, len(data))
}

func TestStackGet(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack via DB
	createStackViaDB(t, testDB, "my-stack")

	w := doRequest(t, router, "GET", "/api/v1/stacks/my-stack", "")
	require.Equal(t, 200, w.Code)

	// Verify envelope
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	// Verify name in response
	dataMap, ok := env["data"].(map[string]interface{})
	require.True(t, ok, "Expected data to be object")
	require.Equal(t, "my-stack", dataMap["name"])
}

func TestStackGetNotFound(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "GET", "/api/v1/stacks/nonexistent", "")
	require.Equal(t, 404, w.Code)

	// Verify envelope has error
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "error")
	errObj := env["error"].(map[string]interface{})
	require.Contains(t, errObj, "code")
	require.Contains(t, errObj, "message")
}

func TestStackUpdate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack
	createStackViaDB(t, testDB, "update-stack")

	// Update description
	w := doRequest(t, router, "PUT", "/api/v1/stacks/update-stack", `{"description":"Updated description"}`)
	require.Equal(t, 200, w.Code)

	// Verify DB row reflects change
	var desc string
	err := testDB.QueryRow("SELECT description FROM stacks WHERE name = $1", "update-stack").Scan(&desc)
	require.NoError(t, err)
	require.Equal(t, "Updated description", desc)
}

func TestStackDelete(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack
	createStackViaDB(t, testDB, "delete-stack")

	// Delete it
	w := doRequest(t, router, "DELETE", "/api/v1/stacks/delete-stack", "")
	require.True(t, w.Code == 200 || w.Code == 204, "Expected 200 or 204 for delete, got %d", w.Code)

	// Verify DB row has non-null deleted_at
	var deletedAt *string
	err := testDB.QueryRow("SELECT deleted_at::text FROM stacks WHERE name = $1", "delete-stack").Scan(&deletedAt)
	require.NoError(t, err)
	require.NotNil(t, deletedAt, "Expected deleted_at to be non-null after soft delete")
}

func TestStackSoftDeleteExcludedFromList(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create and delete a stack
	createStackViaDB(t, testDB, "deleted-stack")
	w := doRequest(t, router, "DELETE", "/api/v1/stacks/deleted-stack", "")
	require.True(t, w.Code == 200 || w.Code == 204)

	// List stacks
	w = doRequest(t, router, "GET", "/api/v1/stacks", "")
	require.Equal(t, 200, w.Code)

	// Verify empty list (or list without deleted stack)
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data, ok := env["data"].([]interface{})
	require.True(t, ok, "Expected data to be array")
	require.Equal(t, 0, len(data), "Expected empty list after deleting only stack")
}

func TestStackSoftDeleteInTrash(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create and delete a stack
	createStackViaDB(t, testDB, "trashed-stack")
	w := doRequest(t, router, "DELETE", "/api/v1/stacks/trashed-stack", "")
	require.True(t, w.Code == 200 || w.Code == 204)

	// Get trash list
	w = doRequest(t, router, "GET", "/api/v1/stacks/trash", "")
	require.Equal(t, 200, w.Code)

	// Verify trash contains the deleted stack
	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data, ok := env["data"].([]interface{})
	require.True(t, ok, "Expected data to be array")
	require.Equal(t, 1, len(data), "Expected trash to contain 1 stack")

	// Verify it's the right stack
	stackMap := data[0].(map[string]interface{})
	require.Equal(t, "trashed-stack", stackMap["name"])
}

func TestStackRestore(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create and delete a stack
	createStackViaDB(t, testDB, "restore-stack")
	w := doRequest(t, router, "DELETE", "/api/v1/stacks/restore-stack", "")
	require.True(t, w.Code == 200 || w.Code == 204)

	// Restore it
	w = doRequest(t, router, "POST", "/api/v1/stacks/trash/restore-stack/restore", "")
	require.Equal(t, 200, w.Code)

	// Verify it's back in the list
	w = doRequest(t, router, "GET", "/api/v1/stacks", "")
	require.Equal(t, 200, w.Code)

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data, ok := env["data"].([]interface{})
	require.True(t, ok, "Expected data to be array")
	require.Equal(t, 1, len(data), "Expected restored stack to be in list")

	// Verify it's the right stack
	stackMap := data[0].(map[string]interface{})
	require.Equal(t, "restore-stack", stackMap["name"])

	// Verify DB deleted_at is NULL
	var deletedAt *string
	err = testDB.QueryRow("SELECT deleted_at::text FROM stacks WHERE name = $1", "restore-stack").Scan(&deletedAt)
	require.NoError(t, err)
	require.Nil(t, deletedAt, "Expected deleted_at to be NULL after restore")
}
