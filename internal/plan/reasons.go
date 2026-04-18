package plan

import "sort"

func workspaceNetworkAddReasons() []string {
	return []string{"isolated workspace network is desired but missing from runtime snapshot"}
}

func workspaceNetworkRemoveReasons() []string {
	return []string{"runtime snapshot has a managed workspace network that is no longer desired"}
}

func workspaceNetworkNoopReasons() []string {
	return []string{"isolated workspace network already matches desired state"}
}

func resourceAddReasons() []string {
	return []string{"resource is not present in runtime snapshot"}
}

func resourceRemoveReasons(disabled bool) []string {
	if disabled {
		return []string{"resource is disabled in desired workspace but still exists in runtime snapshot"}
	}
	return []string{"runtime snapshot contains a managed resource that is no longer desired"}
}

func resourceDisabledNoopReasons() []string {
	return []string{"resource is disabled and absent from runtime snapshot"}
}

func resourceNoopReasons() []string {
	return []string{"resource runtime configuration matches desired state"}
}

func resourceRestartReasons(running bool, health string) []string {
	reasons := make([]string, 0, 2)
	if !running {
		reasons = append(reasons, "resource exists but is not running")
	}
	if health == "unhealthy" {
		reasons = append(reasons, "resource health is unhealthy")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "resource needs restart to reconcile runtime state")
	}
	return reasons
}

func modifyReasons(fields []string) []string {
	messages := make([]string, 0, len(fields))
	for _, field := range fields {
		switch field {
		case "build":
			messages = append(messages, "build configuration changed")
		case "command":
			messages = append(messages, "command changed")
		case "developWatch":
			messages = append(messages, "develop.watch changed")
		case "entrypoint":
			messages = append(messages, "entrypoint changed")
		case "env":
			messages = append(messages, "environment changed")
		case "health":
			messages = append(messages, "health check changed")
		case "image":
			messages = append(messages, "image changed")
		case "labels":
			messages = append(messages, "labels changed")
		case "ports":
			messages = append(messages, "port bindings changed")
		case "projectSource":
			messages = append(messages, "project source handling changed")
		case "volumes":
			messages = append(messages, "volumes changed")
		case "workingDir":
			messages = append(messages, "working directory changed")
		default:
			messages = append(messages, field+" changed")
		}
	}
	sort.Strings(messages)
	return messages
}
