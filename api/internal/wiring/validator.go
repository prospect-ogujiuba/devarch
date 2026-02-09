package wiring

import (
	"database/sql"
	"fmt"
)

func ValidateWiring(wires []WireCandidate, existingWires []ExistingWire) ([]string, error) {
	var warnings []string

	adjacencyList := make(map[int][]int)

	for _, wire := range wires {
		adjacencyList[wire.ConsumerInstanceID] = append(
			adjacencyList[wire.ConsumerInstanceID],
			wire.ProviderInstanceID,
		)

		if wire.ConsumerInstanceID == wire.ProviderInstanceID {
			warnings = append(warnings, fmt.Sprintf(
				"Self-dependency detected: instance %d depends on itself",
				wire.ConsumerInstanceID,
			))
		}
	}

	for _, wire := range existingWires {
		adjacencyList[wire.ConsumerInstanceID] = append(
			adjacencyList[wire.ConsumerInstanceID],
			wire.ProviderInstanceID,
		)
	}

	visited := make(map[int]bool)
	recStack := make(map[int]bool)

	var dfs func(int) bool
	dfs = func(node int) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range adjacencyList[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range adjacencyList {
		if !visited[node] {
			if dfs(node) {
				return warnings, fmt.Errorf("circular dependency detected in wiring graph")
			}
		}
	}

	return warnings, nil
}

func FindOrphanedWires(db *sql.DB, stackID int) ([]int, error) {
	query := `
		SELECT w.id
		FROM service_instance_wires w
		WHERE w.stack_id = $1
		  AND (
		    w.consumer_instance_id NOT IN (
		      SELECT id FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL
		    )
		    OR w.provider_instance_id NOT IN (
		      SELECT id FROM service_instances WHERE stack_id = $1 AND deleted_at IS NULL
		    )
		  )
	`

	rows, err := db.Query(query, stackID)
	if err != nil {
		return nil, fmt.Errorf("query orphaned wires: %w", err)
	}
	defer rows.Close()

	var orphanedIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan orphaned wire id: %w", err)
		}
		orphanedIDs = append(orphanedIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orphaned wires: %w", err)
	}

	return orphanedIDs, nil
}
