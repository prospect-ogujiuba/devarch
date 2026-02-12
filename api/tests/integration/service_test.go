//go:build integration

package integration_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// seedFullService creates a category + service with ports, volumes, env_vars.
// Returns the category ID.
func seedFullService(t *testing.T, db *sql.DB, name string) int {
	t.Helper()

	_, err := db.Exec(`INSERT INTO categories (name) VALUES ('test-cat') ON CONFLICT (name) DO NOTHING`)
	require.NoError(t, err)

	var catID int
	err = db.QueryRow(`SELECT id FROM categories WHERE name = 'test-cat'`).Scan(&catID)
	require.NoError(t, err)

	var serviceID int
	err = db.QueryRow(`
		INSERT INTO services (name, category_id, image_name, image_tag)
		VALUES ($1, $2, 'test-image', 'latest')
		RETURNING id
	`, name, catID).Scan(&serviceID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO service_ports (service_id, host_port, container_port, protocol) VALUES ($1, '8080', '80', 'tcp')`, serviceID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO service_volumes (service_id, volume_type, source, target) VALUES ($1, 'bind', './data', '/data')`, serviceID)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO service_env_vars (service_id, key, value) VALUES ($1, 'APP_ENV', 'test')`, serviceID)
	require.NoError(t, err)

	return catID
}

// decodeEnvelope decodes JSON response body into a generic map.
func decodeEnvelope(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var env map[string]interface{}
	err := json.Unmarshal(body, &env)
	require.NoError(t, err, "failed to decode response: %s", string(body))
	return env
}

func TestServiceCreate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Seed category
	_, err := testDB.Exec(`INSERT INTO categories (name) VALUES ('test-cat') ON CONFLICT (name) DO NOTHING`)
	require.NoError(t, err)
	var catID int
	err = testDB.QueryRow(`SELECT id FROM categories WHERE name = 'test-cat'`).Scan(&catID)
	require.NoError(t, err)

	body := fmt.Sprintf(`{
		"name":"my-service",
		"category_id":%d,
		"image_name":"nginx",
		"image_tag":"alpine",
		"ports":[{"host_port":"9090","container_port":"80","protocol":"tcp"}],
		"volumes":[{"volume_type":"bind","source":"./html","target":"/usr/share/nginx/html"}],
		"env_vars":[{"key":"MODE","value":"prod"}]
	}`, catID)

	w := doRequest(t, router, "POST", "/api/v1/services", body)
	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "data")

	// Verify DB row
	var count int
	err = testDB.QueryRow(`SELECT COUNT(*) FROM services WHERE name = 'my-service'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Verify port was created
	err = testDB.QueryRow(`SELECT COUNT(*) FROM service_ports sp JOIN services s ON sp.service_id = s.id WHERE s.name = 'my-service'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestServiceCreateDuplicate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	_, err := testDB.Exec(`INSERT INTO categories (name) VALUES ('test-cat') ON CONFLICT (name) DO NOTHING`)
	require.NoError(t, err)
	var catID int
	err = testDB.QueryRow(`SELECT id FROM categories WHERE name = 'test-cat'`).Scan(&catID)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"name":"dup-svc","category_id":%d,"image_name":"nginx"}`, catID)

	w := doRequest(t, router, "POST", "/api/v1/services", body)
	require.Equal(t, http.StatusCreated, w.Code)

	// Second create with same name
	w = doRequest(t, router, "POST", "/api/v1/services", body)
	require.True(t, w.Code == 409 || w.Code == 400 || w.Code == 500, "expected error for duplicate, got %d: %s", w.Code, w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "error")
}

func TestServiceList(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "svc-alpha")
	seedFullService(t, testDB, "svc-beta")

	w := doRequest(t, router, "GET", "/api/v1/services", "")
	require.Equal(t, http.StatusOK, w.Code)

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "data")

	data, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data to be array")
	require.Equal(t, 2, len(data))
}

func TestServiceGet(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "get-svc")

	w := doRequest(t, router, "GET", "/api/v1/services/get-svc", "")
	require.Equal(t, http.StatusOK, w.Code)

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "data")

	dataMap, ok := env["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be object")
	require.Equal(t, "get-svc", dataMap["name"])
}

func TestServiceGetNotFound(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "GET", "/api/v1/services/nonexistent", "")
	require.Equal(t, http.StatusNotFound, w.Code)

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "error")
	errObj := env["error"].(map[string]interface{})
	require.Contains(t, errObj, "code")
	require.Contains(t, errObj, "message")
}

func TestServiceUpdate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "update-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/update-svc", `{"image_tag":"v2.0"}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Verify DB reflects change
	var tag string
	err := testDB.QueryRow(`SELECT image_tag FROM services WHERE name = 'update-svc'`).Scan(&tag)
	require.NoError(t, err)
	require.Equal(t, "v2.0", tag)
}

func TestServiceDelete(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "delete-svc")

	w := doRequest(t, router, "DELETE", "/api/v1/services/delete-svc", "")
	require.True(t, w.Code == 200 || w.Code == 204, "expected 200 or 204, got %d: %s", w.Code, w.Body.String())

	// Verify row is gone (hard delete)
	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM services WHERE name = 'delete-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestServiceCompose(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "compose-svc")

	w := doRequest(t, router, "GET", "/api/v1/services/compose-svc/compose", "")
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Compose endpoint returns YAML as text
	body := w.Body.String()
	require.True(t, strings.Contains(body, "compose-svc") || strings.Contains(body, "test-image"),
		"expected YAML to contain service name or image, got: %s", body)
}

func TestServiceVersions(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "version-svc")

	w := doRequest(t, router, "GET", "/api/v1/services/version-svc/versions", "")
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "data")

	// Data is an array (may be empty if no snapshots yet)
	_, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data to be array, got %T", env["data"])
}

func TestServiceValidate(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "validate-svc")

	w := doRequest(t, router, "POST", "/api/v1/services/validate-svc/validate", "")
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	require.Contains(t, env, "data")

	dataMap, ok := env["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be object")
	// Validation result should have a "valid" field
	require.Contains(t, dataMap, "valid")
}

func TestServiceUpdatePorts(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "ports-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/ports-svc/ports",
		`{"ports":[{"host_port":"9090","container_port":"80","protocol":"tcp"},{"host_port":"9091","container_port":"443","protocol":"tcp"}]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Verify DB has 2 ports
	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM service_ports sp JOIN services s ON sp.service_id = s.id WHERE s.name = 'ports-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestServiceUpdateVolumes(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "volumes-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/volumes-svc/volumes",
		`{"volumes":[{"volume_type":"bind","source":"./config","target":"/etc/app"}]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	// Verify DB replaced volumes
	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM service_volumes sv JOIN services s ON sv.service_id = s.id WHERE s.name = 'volumes-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	var target string
	err = testDB.QueryRow(`SELECT sv.target FROM service_volumes sv JOIN services s ON sv.service_id = s.id WHERE s.name = 'volumes-svc'`).Scan(&target)
	require.NoError(t, err)
	require.Equal(t, "/etc/app", target)
}

func TestServiceUpdateEnvVars(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "env-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/env-svc/env-vars",
		`{"env_vars":[{"key":"FOO","value":"bar"},{"key":"BAZ","value":"qux"}]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM service_env_vars ev JOIN services s ON ev.service_id = s.id WHERE s.name = 'env-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestServiceUpdateLabels(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "labels-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/labels-svc/labels",
		`{"labels":[{"key":"app","value":"test"},{"key":"env","value":"dev"}]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM service_labels sl JOIN services s ON sl.service_id = s.id WHERE s.name = 'labels-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestServiceUpdateDomains(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "domains-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/domains-svc/domains",
		`{"domains":[{"domain":"test.local","proxy_port":80}]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var count int
	err := testDB.QueryRow(`SELECT COUNT(*) FROM service_domains sd JOIN services s ON sd.service_id = s.id WHERE s.name = 'domains-svc'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	var domain string
	err = testDB.QueryRow(`SELECT sd.domain FROM service_domains sd JOIN services s ON sd.service_id = s.id WHERE s.name = 'domains-svc'`).Scan(&domain)
	require.NoError(t, err)
	require.Equal(t, "test.local", domain)
}

func TestServiceUpdateHealthcheck(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "health-svc")

	w := doRequest(t, router, "PUT", "/api/v1/services/health-svc/healthcheck",
		`{"test":"CMD curl -f http://localhost","interval_seconds":30,"timeout_seconds":10,"retries":3,"start_period_seconds":5}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var test string
	var interval, timeout, retries, startPeriod int
	err := testDB.QueryRow(`
		SELECT sh.test, sh.interval_seconds, sh.timeout_seconds, sh.retries, sh.start_period_seconds
		FROM service_healthchecks sh JOIN services s ON sh.service_id = s.id
		WHERE s.name = 'health-svc'
	`).Scan(&test, &interval, &timeout, &retries, &startPeriod)
	require.NoError(t, err)
	require.Equal(t, "CMD curl -f http://localhost", test)
	require.Equal(t, 30, interval)
	require.Equal(t, 10, timeout)
	require.Equal(t, 3, retries)
	require.Equal(t, 5, startPeriod)
}

func TestServiceUpdateDependencies(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	// Need two services: the main one and its dependency
	seedFullService(t, testDB, "dep-target")
	seedFullService(t, testDB, "dep-source")

	w := doRequest(t, router, "PUT", "/api/v1/services/dep-source/dependencies",
		`{"dependencies":["dep-target"]}`)
	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var count int
	err := testDB.QueryRow(`
		SELECT COUNT(*) FROM service_dependencies sd
		JOIN services s ON sd.service_id = s.id
		WHERE s.name = 'dep-source'
	`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestServiceConfigFiles(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "cfg-svc")

	// PUT a config file
	w := doRequest(t, router, "PUT", "/api/v1/services/cfg-svc/files/app.conf",
		`{"content":"server { listen 80; }","file_mode":"0644"}`)
	require.Equal(t, http.StatusOK, w.Code, "put body: %s", w.Body.String())

	// GET list of config files
	w = doRequest(t, router, "GET", "/api/v1/services/cfg-svc/files", "")
	require.Equal(t, http.StatusOK, w.Code, "list body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	data, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data array")
	require.Equal(t, 1, len(data))

	// GET single file
	w = doRequest(t, router, "GET", "/api/v1/services/cfg-svc/files/app.conf", "")
	require.Equal(t, http.StatusOK, w.Code, "get body: %s", w.Body.String())

	env = decodeEnvelope(t, w.Body.Bytes())
	fileData, ok := env["data"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "server { listen 80; }", fileData["content"])

	// DELETE config file
	w = doRequest(t, router, "DELETE", "/api/v1/services/cfg-svc/files/app.conf", "")
	require.Equal(t, http.StatusOK, w.Code, "delete body: %s", w.Body.String())

	// Verify gone
	w = doRequest(t, router, "GET", "/api/v1/services/cfg-svc/files", "")
	require.Equal(t, http.StatusOK, w.Code)
	env = decodeEnvelope(t, w.Body.Bytes())
	data, ok = env["data"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 0, len(data))
}

func TestServiceExports(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "export-svc")

	// PUT exports
	w := doRequest(t, router, "PUT", "/api/v1/services/export-svc/exports",
		`{"exports":[{"name":"http","type":"http","port":80,"protocol":"tcp"}]}`)
	require.Equal(t, http.StatusOK, w.Code, "put body: %s", w.Body.String())

	// GET exports
	w = doRequest(t, router, "GET", "/api/v1/services/export-svc/exports", "")
	require.Equal(t, http.StatusOK, w.Code, "get body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	data, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data array")
	require.Equal(t, 1, len(data))

	exportMap := data[0].(map[string]interface{})
	require.Equal(t, "http", exportMap["name"])
	require.Equal(t, float64(80), exportMap["port"])
}

func TestServiceImports(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	seedFullService(t, testDB, "import-svc")

	// PUT imports
	w := doRequest(t, router, "PUT", "/api/v1/services/import-svc/imports",
		`{"imports":[{"name":"database","type":"postgres","required":true,"env_vars":{"DB_HOST":"{{.Host}}","DB_PORT":"{{.Port}}"}}]}`)
	require.Equal(t, http.StatusOK, w.Code, "put body: %s", w.Body.String())

	// GET imports
	w = doRequest(t, router, "GET", "/api/v1/services/import-svc/imports", "")
	require.Equal(t, http.StatusOK, w.Code, "get body: %s", w.Body.String())

	env := decodeEnvelope(t, w.Body.Bytes())
	data, ok := env["data"].([]interface{})
	require.True(t, ok, "expected data array")
	require.Equal(t, 1, len(data))

	importMap := data[0].(map[string]interface{})
	require.Equal(t, "database", importMap["name"])
	require.Equal(t, "postgres", importMap["type"])
	require.Equal(t, true, importMap["required"])
}
