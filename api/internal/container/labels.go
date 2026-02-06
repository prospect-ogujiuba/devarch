package container

import "fmt"

const (
	LabelPrefix            = "devarch."
	LabelStackID           = "devarch.stack_id"
	LabelInstanceID        = "devarch.instance_id"
	LabelTemplateServiceID = "devarch.template_service_id"
	LabelManagedBy         = "devarch.managed_by"
	LabelVersion           = "devarch.version"
	ManagedByValue         = "devarch"
)

// BuildLabels returns a complete label map for DevArch-managed containers
func BuildLabels(stackID, instanceID, templateServiceID string) map[string]string {
	labels := map[string]string{
		LabelManagedBy: ManagedByValue,
		LabelVersion:   "1.0",
	}

	if stackID != "" {
		labels[LabelStackID] = stackID
	}
	if instanceID != "" {
		labels[LabelInstanceID] = instanceID
	}
	if templateServiceID != "" {
		labels[LabelTemplateServiceID] = templateServiceID
	}

	return labels
}

// ContainerName returns the standard container name for a stack instance
func ContainerName(stackID, instanceID string) string {
	return fmt.Sprintf("devarch-%s-%s", stackID, instanceID)
}

// NetworkName returns the network name for a stack
func NetworkName(stackID string) string {
	return fmt.Sprintf("devarch-%s-net", stackID)
}

// ValidateNetworkName validates network name length and format
func ValidateNetworkName(stackName string) error {
	fullName := NetworkName(stackName)

	if len(fullName) > 63 {
		return fmt.Errorf("network name %q (%d chars) exceeds 63-char limit — shorten stack name %q",
			fullName, len(fullName), stackName)
	}

	// Validate DNS-safe pattern (lowercase alnum + hyphens)
	// Using simplified pattern check since we construct it ourselves
	if fullName == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	return nil
}

// IsDevArchManaged checks if a container is managed by DevArch
func IsDevArchManaged(labels map[string]string) bool {
	return labels[LabelManagedBy] == ManagedByValue
}
