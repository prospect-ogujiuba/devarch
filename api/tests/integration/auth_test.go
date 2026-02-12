//go:build integration

package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthValidate(t *testing.T) {
	router := setupRouter(t)

	w := doRequest(t, router, "POST", "/api/v1/auth/validate", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	require.Equal(t, true, data["valid"], "expected valid=true in DevOpen mode with no API key")
}

func TestAuthWSToken(t *testing.T) {
	router := setupRouter(t)

	w := doRequest(t, router, "POST", "/api/v1/auth/ws-token", "")
	require.Equal(t, 200, w.Code, "body: %s", w.Body.String())

	var env map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&env)
	require.NoError(t, err)
	require.Contains(t, env, "data")

	data := env["data"].(map[string]interface{})
	_, hasToken := data["token"]
	require.True(t, hasToken, "expected token field in response")
	require.Equal(t, "", data["token"], "expected empty token in DevOpen mode with no API key")
}
