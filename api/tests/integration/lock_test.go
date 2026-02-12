//go:build integration

package integration_test

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryLockAcquire(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack
	stackID := createStackViaDB(t, testDB, "test-stack")

	// Acquire advisory lock
	var acquired bool
	err := testDB.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquired)
	require.NoError(t, err)
	require.True(t, acquired, "should acquire lock successfully")

	// Cleanup: release lock
	_, err = testDB.Exec("SELECT pg_advisory_unlock($1)", stackID)
	require.NoError(t, err)
}

func TestAdvisoryLockConflict(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack
	stackID := createStackViaDB(t, testDB, "test-stack")

	// Open second DB connection
	db2, err := sql.Open("postgres", testConnStr)
	require.NoError(t, err)
	defer db2.Close()

	// Connection A acquires lock
	var acquiredA bool
	err = testDB.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredA)
	require.NoError(t, err)
	require.True(t, acquiredA, "connection A should acquire lock")

	// Connection B attempts same lock
	var acquiredB bool
	err = db2.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredB)
	require.NoError(t, err)
	require.False(t, acquiredB, "connection B should fail to acquire lock")

	// Cleanup: A releases lock
	_, err = testDB.Exec("SELECT pg_advisory_unlock($1)", stackID)
	require.NoError(t, err)
}

func TestAdvisoryLockReleaseAndReacquire(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack
	stackID := createStackViaDB(t, testDB, "test-stack")

	// Open second DB connection
	db2, err := sql.Open("postgres", testConnStr)
	require.NoError(t, err)
	defer db2.Close()

	// Connection A acquires lock
	var acquiredA bool
	err = testDB.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredA)
	require.NoError(t, err)
	require.True(t, acquiredA, "connection A should acquire lock")

	// A releases lock
	_, err = testDB.Exec("SELECT pg_advisory_unlock($1)", stackID)
	require.NoError(t, err)

	// Connection B acquires same lock
	var acquiredB bool
	err = db2.QueryRow("SELECT pg_try_advisory_lock($1)", stackID).Scan(&acquiredB)
	require.NoError(t, err)
	require.True(t, acquiredB, "connection B should acquire lock after A releases")

	// Cleanup: B releases lock
	_, err = db2.Exec("SELECT pg_advisory_unlock($1)", stackID)
	require.NoError(t, err)
}
