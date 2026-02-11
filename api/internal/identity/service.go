package identity

import (
	"fmt"
	"strings"
)

// NetworkName returns the standard network name for a stack
func NetworkName(stackName string) string {
	return fmt.Sprintf("devarch-%s-net", stackName)
}

// ContainerName returns the standard container name for a stack instance
func ContainerName(stackName, instanceName string) string {
	return fmt.Sprintf("devarch-%s-%s", stackName, instanceName)
}

// ResolveNetworkName returns custom network name if provided, otherwise computed default
func ResolveNetworkName(stackName string, customNetworkName *string) string {
	if customNetworkName != nil && *customNetworkName != "" {
		return *customNetworkName
	}
	return NetworkName(stackName)
}

// ResolveContainerName returns custom container name if provided, otherwise computed default
func ResolveContainerName(stackName, instanceName string, customContainerName *string) string {
	if customContainerName != nil && *customContainerName != "" {
		return *customContainerName
	}
	return ContainerName(stackName, instanceName)
}

// ExportFileName returns the standard export file name for a stack
func ExportFileName(stackName string) string {
	return fmt.Sprintf("%s-devarch.yml", stackName)
}

// LockFileName returns the standard lock file name for a stack
func LockFileName(stackName string) string {
	return fmt.Sprintf("%s-devarch.lock", stackName)
}

// ExtractInstanceName strips the devarch-{stackName}- prefix from a container name
// Returns empty string if the prefix doesn't match
func ExtractInstanceName(stackName, containerName string) string {
	prefix := fmt.Sprintf("devarch-%s-", stackName)
	if strings.HasPrefix(containerName, prefix) {
		return strings.TrimPrefix(containerName, prefix)
	}
	return ""
}
