package runtime

import "fmt"

const (
	NamingStrategyWorkspaceResource = "workspace-resource"

	LabelManagedBy = "devarch.managed-by"
	LabelWorkspace = "devarch.workspace"
	LabelResource  = "devarch.resource"
	LabelHostAlias = "devarch.host"
	LabelNetwork   = "devarch.network"

	ManagedByValue = "devarch-v2"
)

func WorkspaceNetworkName(workspaceName, namingStrategy string) string {
	switch namingStrategy {
	case "", NamingStrategyWorkspaceResource:
		return fmt.Sprintf("devarch-%s-net", workspaceName)
	default:
		return fmt.Sprintf("devarch-%s-net", workspaceName)
	}
}

func ResourceRuntimeName(workspaceName, resourceKey, namingStrategy string) string {
	switch namingStrategy {
	case "", NamingStrategyWorkspaceResource:
		return fmt.Sprintf("devarch-%s-%s", workspaceName, resourceKey)
	default:
		return fmt.Sprintf("devarch-%s-%s", workspaceName, resourceKey)
	}
}

func WorkspaceLabels(workspaceName string) map[string]string {
	return map[string]string{
		LabelManagedBy: ManagedByValue,
		LabelWorkspace: workspaceName,
	}
}

func ResourceLabels(workspaceName, resourceKey, logicalHost, networkName string) map[string]string {
	labels := WorkspaceLabels(workspaceName)
	labels[LabelResource] = resourceKey
	labels[LabelHostAlias] = logicalHost
	if networkName != "" {
		labels[LabelNetwork] = networkName
	}
	return labels
}
