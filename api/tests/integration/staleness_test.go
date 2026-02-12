//go:build integration

package integration_test

import (
	"errors"
	"testing"
	"time"

	"github.com/priz/devarch-api/internal/plan"
	"github.com/stretchr/testify/require"
)

func TestStalenessTokenValid(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack + instance via DB
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// Query updated_at from both
	var stackUpdatedAt time.Time
	err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
	require.NoError(t, err)

	rows, err := testDB.Query("SELECT instance_id, updated_at FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID)
	require.NoError(t, err)
	defer rows.Close()

	var instances []plan.InstanceTimestamp
	for rows.Next() {
		var inst plan.InstanceTimestamp
		err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt)
		require.NoError(t, err)
		instances = append(instances, inst)
	}
	require.Len(t, instances, 1)
	require.Equal(t, instanceID, instances[0].InstanceID)

	// Generate token
	token := plan.GenerateToken(stackUpdatedAt, instances)
	require.NotEmpty(t, token)

	// Validate token
	err = plan.ValidateToken(testDB, stackID, token)
	require.NoError(t, err)
}

func TestStalenessTokenInvalidatedByStackUpdate(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack + instance
	stackID := createStackViaDB(t, testDB, "test-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	// Query updated_at
	var stackUpdatedAt time.Time
	err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
	require.NoError(t, err)

	rows, err := testDB.Query("SELECT instance_id, updated_at FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID)
	require.NoError(t, err)
	defer rows.Close()

	var instances []plan.InstanceTimestamp
	for rows.Next() {
		var inst plan.InstanceTimestamp
		err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt)
		require.NoError(t, err)
		instances = append(instances, inst)
	}

	// Generate token
	token := plan.GenerateToken(stackUpdatedAt, instances)

	// Update stack (triggers updated_at change)
	_, err = testDB.Exec("UPDATE stacks SET description = 'changed', updated_at = NOW() WHERE id = $1", stackID)
	require.NoError(t, err)

	// Validate token - should be stale
	err = plan.ValidateToken(testDB, stackID, token)
	require.Error(t, err)
	require.True(t, errors.Is(err, plan.ErrStalePlan), "expected ErrStalePlan, got: %v", err)
}

func TestStalenessTokenInvalidatedByInstanceUpdate(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack + instance
	stackID := createStackViaDB(t, testDB, "test-stack")
	instanceID := createInstanceViaDB(t, testDB, stackID, "postgres")

	// Query updated_at
	var stackUpdatedAt time.Time
	err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
	require.NoError(t, err)

	rows, err := testDB.Query("SELECT instance_id, updated_at FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID)
	require.NoError(t, err)
	defer rows.Close()

	var instances []plan.InstanceTimestamp
	for rows.Next() {
		var inst plan.InstanceTimestamp
		err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt)
		require.NoError(t, err)
		instances = append(instances, inst)
	}

	// Generate token
	token := plan.GenerateToken(stackUpdatedAt, instances)

	// Update instance updated_at
	_, err = testDB.Exec("UPDATE service_instances SET updated_at = NOW() WHERE instance_id = $1 AND stack_id = $2", instanceID, stackID)
	require.NoError(t, err)

	// Validate token - should be stale
	err = plan.ValidateToken(testDB, stackID, token)
	require.Error(t, err)
	require.True(t, errors.Is(err, plan.ErrStalePlan), "expected ErrStalePlan, got: %v", err)
}

func TestStalenessTokenInvalidatedByNewInstance(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack + 1 instance
	stackID := createStackViaDB(t, testDB, "test-stack")
	createInstanceViaDB(t, testDB, stackID, "postgres")

	// Query updated_at
	var stackUpdatedAt time.Time
	err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
	require.NoError(t, err)

	rows, err := testDB.Query("SELECT instance_id, updated_at FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID)
	require.NoError(t, err)
	defer rows.Close()

	var instances []plan.InstanceTimestamp
	for rows.Next() {
		var inst plan.InstanceTimestamp
		err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt)
		require.NoError(t, err)
		instances = append(instances, inst)
	}

	// Generate token
	token := plan.GenerateToken(stackUpdatedAt, instances)

	// Add second instance
	createInstanceViaDB(t, testDB, stackID, "redis")

	// Validate token - should be stale
	err = plan.ValidateToken(testDB, stackID, token)
	require.Error(t, err)
	require.True(t, errors.Is(err, plan.ErrStalePlan), "expected ErrStalePlan, got: %v", err)
}

func TestStalenessTokenInvalidatedByInstanceDelete(t *testing.T) {
	truncateAll(t, testDB)

	// Create stack + 2 instances
	stackID := createStackViaDB(t, testDB, "test-stack")
	instance1 := createInstanceViaDB(t, testDB, stackID, "postgres")
	createInstanceViaDB(t, testDB, stackID, "redis")

	// Query updated_at
	var stackUpdatedAt time.Time
	err := testDB.QueryRow("SELECT updated_at FROM stacks WHERE id = $1", stackID).Scan(&stackUpdatedAt)
	require.NoError(t, err)

	rows, err := testDB.Query("SELECT instance_id, updated_at FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL", stackID)
	require.NoError(t, err)
	defer rows.Close()

	var instances []plan.InstanceTimestamp
	for rows.Next() {
		var inst plan.InstanceTimestamp
		err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt)
		require.NoError(t, err)
		instances = append(instances, inst)
	}

	// Generate token
	token := plan.GenerateToken(stackUpdatedAt, instances)

	// Soft-delete one instance
	_, err = testDB.Exec("UPDATE service_instances SET deleted_at = NOW() WHERE instance_id = $1 AND stack_id = $2", instance1, stackID)
	require.NoError(t, err)

	// Validate token - should be stale
	err = plan.ValidateToken(testDB, stackID, token)
	require.Error(t, err)
	require.True(t, errors.Is(err, plan.ErrStalePlan), "expected ErrStalePlan, got: %v", err)
}
