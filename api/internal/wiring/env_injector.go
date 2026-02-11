package wiring

import (
	"strconv"
	"strings"

	"github.com/priz/devarch-api/internal/identity"
)

func InjectEnvVars(stackName string, provider Provider, consumer Consumer) map[string]string {
	hostname := identity.ContainerName(stackName, provider.InstanceName)
	portStr := strconv.Itoa(provider.Port)
	protocol := provider.Protocol
	name := provider.ContractName

	injections := make(map[string]string)

	for envKey, template := range consumer.EnvVars {
		value := template
		value = strings.ReplaceAll(value, "{{hostname}}", hostname)
		value = strings.ReplaceAll(value, "{{port}}", portStr)
		value = strings.ReplaceAll(value, "{{protocol}}", protocol)
		value = strings.ReplaceAll(value, "{{name}}", name)

		injections[envKey] = value
	}

	return injections
}
