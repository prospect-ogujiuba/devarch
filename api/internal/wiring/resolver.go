package wiring

import (
	"fmt"
	"sort"
)

type Provider struct {
	InstanceID       int
	InstanceName     string
	ExportContractID int
	ContractName     string
	ContractType     string
	Port             int
	Protocol         string
}

type Consumer struct {
	InstanceID       int
	InstanceName     string
	ImportContractID int
	ContractName     string
	ContractType     string
	Required         bool
	EnvVars          map[string]string
}

type ExistingWire struct {
	ID                 int
	ConsumerInstanceID int
	ProviderInstanceID int
	ImportContractID   int
	Source             string
}

type WireCandidate struct {
	ConsumerInstanceID int
	ProviderInstanceID int
	ImportContractID   int
	ExportContractID   int
	ContractName       string
	ConsumerType       string
	ProviderType       string
	EnvVarInjections   map[string]string
	Source             string
}

func ResolveAutoWires(stackName string, providers []Provider, consumers []Consumer, existingWires []ExistingWire) ([]WireCandidate, []string) {
	var candidates []WireCandidate
	var warnings []string

	explicitWires := make(map[wireKey]bool)
	for _, wire := range existingWires {
		if wire.Source == "explicit" {
			key := wireKey{
				consumerInstanceID: wire.ConsumerInstanceID,
				importContractID:   wire.ImportContractID,
			}
			explicitWires[key] = true
		}
	}

	for _, consumer := range consumers {
		key := wireKey{
			consumerInstanceID: consumer.InstanceID,
			importContractID:   consumer.ImportContractID,
		}

		if explicitWires[key] {
			continue
		}

		var matches []Provider
		for _, provider := range providers {
			if provider.ContractType == consumer.ContractType {
				matches = append(matches, provider)
			}
		}

		if len(matches) == 0 {
			if consumer.Required {
				warnings = append(warnings, fmt.Sprintf(
					"Missing required contract: %s needs %s (type: %s)",
					consumer.InstanceName,
					consumer.ContractName,
					consumer.ContractType,
				))
			}
			continue
		}

		if len(matches) == 1 {
			provider := matches[0]
			injections := InjectEnvVars(stackName, provider, consumer)

			candidate := WireCandidate{
				ConsumerInstanceID: consumer.InstanceID,
				ProviderInstanceID: provider.InstanceID,
				ImportContractID:   consumer.ImportContractID,
				ExportContractID:   provider.ExportContractID,
				ContractName:       consumer.ContractName,
				ConsumerType:       consumer.ContractType,
				ProviderType:       provider.ContractType,
				EnvVarInjections:   injections,
				Source:             "auto",
			}
			candidates = append(candidates, candidate)
			continue
		}

		warnings = append(warnings, fmt.Sprintf(
			"Ambiguous: %d %s providers for %s.%s — create explicit wire",
			len(matches),
			consumer.ContractType,
			consumer.InstanceName,
			consumer.ContractName,
		))
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].ConsumerInstanceID != candidates[j].ConsumerInstanceID {
			return candidates[i].ConsumerInstanceID < candidates[j].ConsumerInstanceID
		}
		return candidates[i].ContractName < candidates[j].ContractName
	})

	return candidates, warnings
}

type wireKey struct {
	consumerInstanceID int
	importContractID   int
}
