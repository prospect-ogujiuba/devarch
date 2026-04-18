package contracts

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/resolve"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

// Resolve resolves imports against enabled providers in the effective graph.
// Explicit import.from targets win over auto-linking. Auto-linking only occurs
// when exactly one enabled provider exports the requested contract.
func Resolve(graph *resolve.Graph) *Result {
	if graph == nil {
		return &Result{}
	}

	providersByContract := make(map[string][]*resolve.Resource)
	resourcesByKey := make(map[string]*resolve.Resource, len(graph.Resources))
	for _, resource := range graph.Resources {
		if resource == nil {
			continue
		}
		resourcesByKey[resource.Key] = resource
		if !resource.Enabled {
			continue
		}
		for _, export := range resource.Exports {
			providersByContract[export.Contract] = append(providersByContract[export.Contract], resource)
		}
	}
	for contract := range providersByContract {
		sort.Slice(providersByContract[contract], func(i, j int) bool {
			return providersByContract[contract][i].Key < providersByContract[contract][j].Key
		})
	}

	result := &Result{}
	for _, consumer := range graph.Resources {
		if consumer == nil || !consumer.Enabled {
			continue
		}
		for _, imp := range consumer.Imports {
			if imp.From != "" {
				link, diagnostics := resolveExplicitImport(resourcesByKey, consumer, imp)
				if link != nil {
					result.Links = append(result.Links, *link)
				}
				result.Diagnostics = append(result.Diagnostics, diagnostics...)
				continue
			}

			matches := candidateProviders(providersByContract[imp.Contract], consumer.Key)
			switch len(matches) {
			case 0:
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: "warning",
					Code:     "unresolved-import",
					Consumer: consumer.Key,
					Contract: imp.Contract,
					Message:  fmt.Sprintf("no enabled providers export contract %q", imp.Contract),
				})
			case 1:
				link, diagnostics := buildLink(consumer, imp, matches[0], "auto")
				result.Links = append(result.Links, *link)
				result.Diagnostics = append(result.Diagnostics, diagnostics...)
			default:
				providerKeys := make([]string, 0, len(matches))
				for _, provider := range matches {
					providerKeys = append(providerKeys, provider.Key)
				}
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity:  "warning",
					Code:      "ambiguous-import",
					Consumer:  consumer.Key,
					Contract:  imp.Contract,
					Providers: providerKeys,
					Message:   fmt.Sprintf("multiple enabled providers export contract %q: %s", imp.Contract, strings.Join(providerKeys, ", ")),
				})
			}
		}
	}

	sortResult(result)
	return result
}

func resolveExplicitImport(resourcesByKey map[string]*resolve.Resource, consumer *resolve.Resource, imp resolve.Import) (*Link, []Diagnostic) {
	provider, ok := resourcesByKey[imp.From]
	if !ok {
		return nil, []Diagnostic{{
			Severity: "warning",
			Code:     "explicit-provider-not-found",
			Consumer: consumer.Key,
			Contract: imp.Contract,
			Provider: imp.From,
			Message:  fmt.Sprintf("explicit provider %q was not found", imp.From),
		}}
	}
	if !provider.Enabled {
		return nil, []Diagnostic{{
			Severity: "warning",
			Code:     "explicit-provider-disabled",
			Consumer: consumer.Key,
			Contract: imp.Contract,
			Provider: provider.Key,
			Message:  fmt.Sprintf("explicit provider %q is disabled", provider.Key),
		}}
	}
	if _, ok := exportForContract(provider, imp.Contract); !ok {
		return nil, []Diagnostic{{
			Severity: "warning",
			Code:     "explicit-provider-missing-contract",
			Consumer: consumer.Key,
			Contract: imp.Contract,
			Provider: provider.Key,
			Message:  fmt.Sprintf("explicit provider %q does not export contract %q", provider.Key, imp.Contract),
		}}
	}

	link, diagnostics := buildLink(consumer, imp, provider, "explicit")
	return link, diagnostics
}

func candidateProviders(providers []*resolve.Resource, consumerKey string) []*resolve.Resource {
	if len(providers) == 0 {
		return nil
	}

	filtered := make([]*resolve.Resource, 0, len(providers))
	for _, provider := range providers {
		if provider.Key == consumerKey {
			continue
		}
		filtered = append(filtered, provider)
	}
	return filtered
}

func buildLink(consumer *resolve.Resource, imp resolve.Import, provider *resolve.Resource, source string) (*Link, []Diagnostic) {
	export, _ := exportForContract(provider, imp.Contract)
	injectedEnv := make(map[string]workspace.EnvValue, len(export.Env))
	diagnostics := make([]Diagnostic, 0)
	for envKey, template := range export.Env {
		value, ok, interpolationDiagnostics := interpolateExportValue(consumer.Key, imp.Contract, envKey, provider, template)
		diagnostics = append(diagnostics, interpolationDiagnostics...)
		if !ok {
			continue
		}
		injectedEnv[envKey] = value
	}
	if len(injectedEnv) == 0 {
		injectedEnv = nil
	}

	return &Link{
		Consumer:    consumer.Key,
		Contract:    imp.Contract,
		Alias:       imp.Alias,
		Provider:    provider.Key,
		Source:      source,
		InjectedEnv: injectedEnv,
	}, diagnostics
}

func exportForContract(provider *resolve.Resource, contract string) (resolve.Export, bool) {
	for _, export := range provider.Exports {
		if export.Contract == contract {
			return export, true
		}
	}
	return resolve.Export{}, false
}

func sortResult(result *Result) {
	sort.Slice(result.Links, func(i, j int) bool {
		if result.Links[i].Consumer != result.Links[j].Consumer {
			return result.Links[i].Consumer < result.Links[j].Consumer
		}
		if result.Links[i].Contract != result.Links[j].Contract {
			return result.Links[i].Contract < result.Links[j].Contract
		}
		if result.Links[i].Provider != result.Links[j].Provider {
			return result.Links[i].Provider < result.Links[j].Provider
		}
		if result.Links[i].Alias != result.Links[j].Alias {
			return result.Links[i].Alias < result.Links[j].Alias
		}
		return result.Links[i].Source < result.Links[j].Source
	})

	sort.Slice(result.Diagnostics, func(i, j int) bool {
		if result.Diagnostics[i].Consumer != result.Diagnostics[j].Consumer {
			return result.Diagnostics[i].Consumer < result.Diagnostics[j].Consumer
		}
		if result.Diagnostics[i].Contract != result.Diagnostics[j].Contract {
			return result.Diagnostics[i].Contract < result.Diagnostics[j].Contract
		}
		if result.Diagnostics[i].Provider != result.Diagnostics[j].Provider {
			return result.Diagnostics[i].Provider < result.Diagnostics[j].Provider
		}
		providersI := strings.Join(result.Diagnostics[i].Providers, ",")
		providersJ := strings.Join(result.Diagnostics[j].Providers, ",")
		if providersI != providersJ {
			return providersI < providersJ
		}
		if result.Diagnostics[i].EnvKey != result.Diagnostics[j].EnvKey {
			return result.Diagnostics[i].EnvKey < result.Diagnostics[j].EnvKey
		}
		if result.Diagnostics[i].Code != result.Diagnostics[j].Code {
			return result.Diagnostics[i].Code < result.Diagnostics[j].Code
		}
		return result.Diagnostics[i].Message < result.Diagnostics[j].Message
	})
}
