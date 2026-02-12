//go:build integration

package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceCreate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack via DB
	stackID := createStackViaDB(t, testDB, "test-stack")
	seedServiceTemplate(t, testDB, "postgres")

	// Get service ID for template_service_id
	var serviceID int
	err := testDB.QueryRow("SELECT id FROM services WHERE name = $1", "postgres").Scan(&serviceID)
	require.NoError(t, err)

	// POST create instance
	payload := map[string]interface{}{
		"instance_id":         "postgres-1",
		"template_service_id": serviceID,
		"description":         "Test postgres instance",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stacks/test-stack/instances", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 201 Created
	require.Equal(t, http.StatusCreated, rec.Code, "response body: %s", rec.Body.String())

	// Assert envelope has data
	var envelope map[string]interface{}
	err = json.NewDecoder(rec.Body).Decode(&envelope)
	require.NoError(t, err)
	require.Contains(t, envelope, "data")
	data := envelope["data"].(map[string]interface{})
	require.Equal(t, "postgres-1", data["instance_id"])

	// Assert DB row exists
	var instanceID string
	err = testDB.QueryRow("SELECT instance_id FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID).Scan(&instanceID)
	require.NoError(t, err)
	require.Equal(t, "postgres-1", instanceID)
}

func TestInstanceList(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack + 2 instances via DB
	stackID := createStackViaDB(t, testDB, "test-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")
	createInstanceViaDB(t, testDB, stackID, "redis")

	// GET list instances
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/test-stack/instances", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 200 OK
	require.Equal(t, http.StatusOK, rec.Code, "response body: %s", rec.Body.String())

	// Assert envelope data is array with 2 items
	var envelope map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&envelope)
	require.NoError(t, err)
	require.Contains(t, envelope, "data")
	data := envelope["data"].([]interface{})
	require.Len(t, data, 2)
}

func TestInstanceGet(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack + instance via DB
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// GET instance
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s", instanceID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 200 OK
	require.Equal(t, http.StatusOK, rec.Code, "response body: %s", rec.Body.String())

	// Assert envelope data has correct fields
	var envelope map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&envelope)
	require.NoError(t, err)
	require.Contains(t, envelope, "data")
	data := envelope["data"].(map[string]interface{})
	require.Equal(t, instanceID, data["instance_id"])
}

func TestInstanceGetNotFound(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack (no instances)
	createStackViaDB(t, testDB, "test-stack")

	// GET nonexistent instance
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/test-stack/instances/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 404 Not Found
	require.Equal(t, http.StatusNotFound, rec.Code, "response body: %s", rec.Body.String())

	// Assert error envelope
	var envelope map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&envelope)
	require.NoError(t, err)
	require.Contains(t, envelope, "error")
	errorObj := envelope["error"].(map[string]interface{})
	require.Contains(t, errorObj, "code")
	require.Contains(t, errorObj, "message")
}

func TestInstanceDelete(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack + instance
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// DELETE instance
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s", instanceID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 200 or 204
	require.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNoContent, "expected 200 or 204, got %d", rec.Code)

	// Assert DB row has deleted_at set
	var deletedAt sql.NullTime
	err := testDB.QueryRow("SELECT deleted_at FROM service_instances WHERE stack_id = $1 AND instance_id = $2", stackID, instanceID).Scan(&deletedAt)
	require.NoError(t, err)
	require.True(t, deletedAt.Valid, "deleted_at should be set")
}

func TestInstanceDeletedExcludedFromList(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Create stack + instance
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// Delete instance via DB (soft-delete)
	_, err := testDB.Exec("UPDATE service_instances SET deleted_at = NOW() WHERE stack_id = $1 AND instance_id = $2", stackID, instanceID)
	require.NoError(t, err)

	// GET instances list
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stacks/test-stack/instances", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Assert 200 OK
	require.Equal(t, http.StatusOK, rec.Code, "response body: %s", rec.Body.String())

	// Assert empty or not containing deleted instance
	var envelope map[string]interface{}
	err = json.NewDecoder(rec.Body).Decode(&envelope)
	require.NoError(t, err)
	require.Contains(t, envelope, "data")
	data := envelope["data"].([]interface{})
	require.Empty(t, data, "deleted instance should not be in list")
}
