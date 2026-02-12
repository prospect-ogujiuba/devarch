//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstanceDuplicate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "POST",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/duplicate", instanceID),
		`{"instance_id":"copy-1"}`)
	require.Equal(t, 201, w.Code, "body: %s", w.Body.String())

	// Verify new row in DB
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND instance_id = 'copy-1' AND deleted_at IS NULL", stackID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "duplicated instance should exist in DB")

	// Original still exists
	err = testDB.QueryRow("SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND instance_id = $2 AND deleted_at IS NULL", stackID, instanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "original instance should still exist")
}

func TestInstanceRename(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/rename", instanceID),
		`{"instance_id":"new-name"}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// Old name gone
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND instance_id = $2 AND deleted_at IS NULL", stackID, instanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "old instance name should be gone")

	// New name exists
	err = testDB.QueryRow("SELECT COUNT(*) FROM service_instances WHERE stack_id = $1 AND instance_id = 'new-name' AND deleted_at IS NULL", stackID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "new instance name should exist")
}

func TestInstanceDeletePreview(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "GET",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/delete-preview", instanceID), "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")
}

func TestInstanceUpdatePorts(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/ports", instanceID),
		`{"ports":[{"host_port":"9090","container_port":"80","protocol":"tcp"}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// Verify DB
	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_ports WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 port override")
}

func TestInstanceUpdateVolumes(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/volumes", instanceID),
		`{"volumes":[{"volume_type":"bind","source":"./data","target":"/data"}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_volumes WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 volume override")
}

func TestInstanceUpdateEnvVars(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/env-vars", instanceID),
		`{"env_vars":[{"key":"FOO","value":"bar"}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_env_vars WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 env var override")
}

func TestInstanceUpdateEnvFiles(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/env-files", instanceID),
		`{"env_files":["./env/.env"]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_env_files WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 env file override")
}

func TestInstanceUpdateNetworks(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/networks", instanceID),
		`{"networks":["custom-net"]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_networks WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 network override")
}

func TestInstanceUpdateConfigMounts(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/config-mounts", instanceID),
		`{"config_mounts":[{"source_path":"./conf","target_path":"/etc/conf","readonly":true}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_config_mounts WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 config mount override")
}

func TestInstanceUpdateLabels(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/labels", instanceID),
		`{"labels":[{"key":"env","value":"test"}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_labels WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 label override")
}

func TestInstanceUpdateDomains(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/domains", instanceID),
		`{"domains":[{"domain":"test.local","proxy_port":80}]}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_domains WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 domain override")
}

func TestInstanceUpdateHealthcheck(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// Healthcheck handler decodes directly into *models.ServiceHealthcheck (not wrapped)
	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/healthcheck", instanceID),
		`{"test":"CMD curl -f http://localhost","interval_seconds":30,"timeout_seconds":10,"retries":3,"start_period_seconds":5}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_healthchecks WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 healthcheck override")
}

func TestInstanceUpdateDependencies(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")
	otherID := createInstanceViaDB(t, testDB, stackID, "redis")

	// Dependencies handler expects [{depends_on, condition}] not plain strings
	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/dependencies", instanceID),
		fmt.Sprintf(`{"dependencies":[{"depends_on":"%s"}]}`, otherID))
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var dbInstanceID int
	err := testDB.QueryRow("SELECT si.id FROM service_instances si WHERE si.stack_id = $1 AND si.instance_id = $2", stackID, instanceID).Scan(&dbInstanceID)
	require.NoError(t, err)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM instance_dependencies WHERE instance_id = $1", dbInstanceID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "should have 1 dependency override")
}

func TestInstanceResourceLimits(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// PUT resource limits
	w := doRequest(t, router, "PUT",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/resources", instanceID),
		`{"cpu_limit":"1.0","memory_limit":"512m"}`)
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	// GET resource limits and verify
	w = doRequest(t, router, "GET",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/resources", instanceID), "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")
	data := env["data"].(map[string]interface{})
	require.Equal(t, "1.0", data["cpu_limit"])
	require.Equal(t, "512m", data["memory_limit"])
}

func TestInstanceConfigFiles(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	basePath := fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/files", instanceID)

	// PUT config file (handler expects JSON with content/file_mode/is_template)
	w := doRequest(t, router, "PUT", basePath+"/app.conf",
		`{"content":"server { listen 80; }","file_mode":"0644","is_template":false}`)
	require.Equal(t, 200, w.Code, "put body: %s", w.Body.String())

	// GET list
	w = doRequest(t, router, "GET", basePath, "")
	require.Equal(t, 200, w.Code, "list body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")
	data := env["data"].([]interface{})
	require.Len(t, data, 1, "should list 1 config file")

	// DELETE config file
	w = doRequest(t, router, "DELETE", basePath+"/app.conf", "")
	require.True(t, w.Code == 200 || w.Code == 204, "delete expected 200 or 204, got %d: %s", w.Code, w.Body.String())

	// GET list again - should be empty
	w = doRequest(t, router, "GET", basePath, "")
	require.Equal(t, 200, w.Code)

	var env2 map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&env2)
	require.NoError(t, err)
	require.Contains(t, env2, "data")
	data2 := env2["data"].([]interface{})
	require.Len(t, data2, 0, "should list 0 config files after delete")
}

func TestInstanceEffectiveConfig(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// Seed a port on the template service so effective-config has something to merge
	var serviceID int
	err := testDB.QueryRow("SELECT id FROM services WHERE name = 'postgres'").Scan(&serviceID)
	require.NoError(t, err)
	_, err = testDB.Exec(
		"INSERT INTO service_ports (service_id, host_port, container_port, protocol) VALUES ($1, '5432', '5432', 'tcp')",
		serviceID)
	require.NoError(t, err)

	w := doRequest(t, router, "GET",
		fmt.Sprintf("/api/v1/stacks/test-stack/instances/%s/effective-config", instanceID), "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")
}
