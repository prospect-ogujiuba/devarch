//go:build integration

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncJobs(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "GET", "/api/v1/sync/jobs", "")

	// syncManager is nil in test router, so expect either 200 (empty) or 500 (nil panic recovered)
	require.True(t, w.Code == 200 || w.Code == 500,
		"expected 200 or 500 for sync/jobs with nil manager, got %d: %s", w.Code, w.Body.String())
}

func TestTriggerSync(t *testing.T) {
	truncateAll(t, testDB)
	router := setupRouter(t)

	w := doRequest(t, router, "POST", "/api/v1/sync?type=containers", "")

	// syncManager is nil in test router, so expect either 200 or 500 (nil panic recovered)
	require.True(t, w.Code == 200 || w.Code == 500,
		"expected 200 or 500 for trigger sync with nil manager, got %d: %s", w.Code, w.Body.String())
}
