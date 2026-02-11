package lock

import (
	"database/sql"
	"fmt"

	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/identity"
)

type Validator struct {
	db              *sql.DB
	containerClient *container.Client
}

func NewValidator(db *sql.DB, cc *container.Client) *Validator {
	return &Validator{
		db:              db,
		containerClient: cc,
	}
}

func (v *Validator) Validate(lock *LockFile, stackName string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Warnings: []LockWarning{},
	}

	netInfo, err := v.containerClient.InspectNetwork(lock.Stack.NetworkName)
	if err != nil {
		result.Warnings = append(result.Warnings, LockWarning{
			Severity: "warn",
			Field:    "stack.network_id",
			Expected: lock.Stack.NetworkID,
			Actual:   "not found",
			Message:  fmt.Sprintf("Network %s not found", lock.Stack.NetworkName),
		})
	} else if lock.Stack.NetworkID != "" && netInfo.ID != lock.Stack.NetworkID {
		result.Warnings = append(result.Warnings, LockWarning{
			Severity: "warn",
			Field:    "stack.network_id",
			Expected: lock.Stack.NetworkID,
			Actual:   netInfo.ID,
			Message:  "Network ID changed",
		})
	}

	containers, err := v.containerClient.ListContainersWithLabels(map[string]string{
		"devarch.stack_id": stackName,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	runningInstances := make(map[string]string)
	for _, containerName := range containers {
		instanceName := identity.ExtractInstanceName(stackName, containerName)
		if instanceName != "" {
			runningInstances[instanceName] = containerName
		}
	}

	for instanceName, lockedInst := range lock.Instances {
		containerName, exists := runningInstances[instanceName]
		if !exists {
			result.Warnings = append(result.Warnings, LockWarning{
				Severity: "warn",
				Field:    fmt.Sprintf("instances.%s", instanceName),
				Expected: "running",
				Actual:   "not found",
				Message:  fmt.Sprintf("Instance %s not running", instanceName),
			})
			continue
		}

		gen := &Generator{db: v.db, containerClient: v.containerClient}
		currentInst, err := gen.buildInstanceLock(stackName, instanceName, containerName)
		if err != nil {
			continue
		}

		if lockedInst.ImageDigest != "" && currentInst.ImageDigest != "" && lockedInst.ImageDigest != currentInst.ImageDigest {
			result.Warnings = append(result.Warnings, LockWarning{
				Severity: "warn",
				Field:    fmt.Sprintf("instances.%s.image_digest", instanceName),
				Expected: lockedInst.ImageDigest,
				Actual:   currentInst.ImageDigest,
				Message:  fmt.Sprintf("Image digest changed for %s", instanceName),
			})
		}

		if lockedInst.TemplateHash != currentInst.TemplateHash {
			result.Warnings = append(result.Warnings, LockWarning{
				Severity: "warn",
				Field:    fmt.Sprintf("instances.%s.template_hash", instanceName),
				Expected: lockedInst.TemplateHash,
				Actual:   currentInst.TemplateHash,
				Message:  fmt.Sprintf("Template version changed for %s", instanceName),
			})
		}

		for portProto, expectedPort := range lockedInst.ResolvedPorts {
			if actualPort, ok := currentInst.ResolvedPorts[portProto]; !ok || actualPort != expectedPort {
				actualPortStr := "not bound"
				if ok {
					actualPortStr = fmt.Sprintf("%d", actualPort)
				}
				result.Warnings = append(result.Warnings, LockWarning{
					Severity: "warn",
					Field:    fmt.Sprintf("instances.%s.resolved_ports.%s", instanceName, portProto),
					Expected: fmt.Sprintf("%d", expectedPort),
					Actual:   actualPortStr,
					Message:  fmt.Sprintf("Port binding changed for %s:%s", instanceName, portProto),
				})
			}
		}
	}

	for instanceName := range runningInstances {
		if _, locked := lock.Instances[instanceName]; !locked {
			result.Warnings = append(result.Warnings, LockWarning{
				Severity: "warn",
				Field:    fmt.Sprintf("instances.%s", instanceName),
				Expected: "not present",
				Actual:   "running",
				Message:  fmt.Sprintf("Unlocked container %s running", instanceName),
			})
		}
	}

	if len(result.Warnings) > 0 {
		result.Valid = false
	}

	return result, nil
}

func (v *Validator) ValidateConfigHash(lock *LockFile, ymlContent []byte) bool {
	return lock.ConfigHash == ComputeHash(ymlContent)
}
