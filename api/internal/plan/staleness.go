package plan

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

var ErrStalePlan = fmt.Errorf("plan is stale: stack state has changed since plan was generated")

type InstanceTimestamp struct {
	InstanceID string
	UpdatedAt  time.Time
}

func GenerateToken(stackUpdatedAt time.Time, instances []InstanceTimestamp) string {
	sorted := make([]InstanceTimestamp, len(instances))
	copy(sorted, instances)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].InstanceID < sorted[j].InstanceID
	})

	h := sha256.New()
	h.Write([]byte(stackUpdatedAt.Format(time.RFC3339Nano)))
	for _, inst := range sorted {
		h.Write([]byte(inst.UpdatedAt.Format(time.RFC3339Nano)))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func ValidateToken(db *sql.DB, stackID int, token string) error {
	var stackUpdatedAt time.Time
	err := db.QueryRow(`
		SELECT updated_at
		FROM stacks
		WHERE id = $1 AND deleted_at IS NULL
	`, stackID).Scan(&stackUpdatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("stack not found: %d", stackID)
	}
	if err != nil {
		return fmt.Errorf("query stack updated_at: %w", err)
	}

	rows, err := db.Query(`
		SELECT instance_id, updated_at
		FROM service_instances
		WHERE stack_id = $1 AND deleted_at IS NULL
	`, stackID)
	if err != nil {
		return fmt.Errorf("query instance timestamps: %w", err)
	}
	defer rows.Close()

	var instances []InstanceTimestamp
	for rows.Next() {
		var inst InstanceTimestamp
		if err := rows.Scan(&inst.InstanceID, &inst.UpdatedAt); err != nil {
			return fmt.Errorf("scan instance timestamp: %w", err)
		}
		instances = append(instances, inst)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate instance rows: %w", err)
	}

	currentToken := GenerateToken(stackUpdatedAt, instances)
	if currentToken != token {
		return ErrStalePlan
	}

	return nil
}
