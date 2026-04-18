package contracts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/resolve"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

func interpolateExportValue(consumer, contract, envKey string, provider *resolve.Resource, template string) (workspace.EnvValue, bool, []Diagnostic) {
	matches := exportPlaceholderMatches(template)
	if len(matches) == 0 {
		return workspace.StringEnvValue(template), true, nil
	}

	if len(matches) == 1 && matches[0].start == 0 && matches[0].end == len(template) {
		value, ok, diagnostic := resolvePlaceholder(consumer, contract, envKey, provider, matches[0].token)
		if !ok {
			return workspace.EnvValue{}, false, []Diagnostic{diagnostic}
		}
		return value, true, nil
	}

	var builder strings.Builder
	cursor := 0
	for _, match := range matches {
		builder.WriteString(template[cursor:match.start])
		value, ok, diagnostic := resolvePlaceholder(consumer, contract, envKey, provider, match.token)
		if !ok {
			return workspace.EnvValue{}, false, []Diagnostic{diagnostic}
		}
		if _, secret := value.SecretRef(); secret {
			return workspace.EnvValue{}, false, []Diagnostic{newInterpolationDiagnostic(
				"error",
				"secret-flatten",
				consumer,
				contract,
				provider.Key,
				envKey,
				fmt.Sprintf("export env %q from provider %q cannot flatten secretRef into composite string", envKey, provider.Key),
			)}
		}
		builder.WriteString(value.Text())
		cursor = match.end
	}
	builder.WriteString(template[cursor:])

	return workspace.StringEnvValue(builder.String()), true, nil
}

type placeholderMatch struct {
	start int
	end   int
	token string
}

func exportPlaceholderMatches(template string) []placeholderMatch {
	matches := make([]placeholderMatch, 0)
	for index := 0; index < len(template); {
		start := strings.Index(template[index:], "${")
		if start < 0 {
			break
		}
		start += index
		end := strings.IndexByte(template[start+2:], '}')
		if end < 0 {
			break
		}
		end += start + 2
		matches = append(matches, placeholderMatch{
			start: start,
			end:   end + 1,
			token: template[start+2 : end],
		})
		index = end + 1
	}
	return matches
}

func resolvePlaceholder(consumer, contract, envKey string, provider *resolve.Resource, token string) (workspace.EnvValue, bool, Diagnostic) {
	switch {
	case strings.HasPrefix(token, "env."):
		key := strings.TrimPrefix(token, "env.")
		value, ok := provider.Env[key]
		if !ok {
			return workspace.EnvValue{}, false, newInterpolationDiagnostic(
				"warning",
				"missing-export-placeholder",
				consumer,
				contract,
				provider.Key,
				envKey,
				fmt.Sprintf("provider %q has no env key %q required by export env %q", provider.Key, key, envKey),
			)
		}
		return value.Clone(), true, Diagnostic{}
	case token == "resource.host":
		return workspace.StringEnvValue(provider.Host), true, Diagnostic{}
	case strings.HasPrefix(token, "resource.port."):
		portValue := strings.TrimPrefix(token, "resource.port.")
		port, err := strconv.Atoi(portValue)
		if err != nil {
			return workspace.EnvValue{}, false, newInterpolationDiagnostic(
				"warning",
				"invalid-export-placeholder",
				consumer,
				contract,
				provider.Key,
				envKey,
				fmt.Sprintf("invalid port placeholder %q in export env %q", token, envKey),
			)
		}
		for _, resourcePort := range provider.Ports {
			if resourcePort.Container == port {
				return workspace.StringEnvValue(strconv.Itoa(resourcePort.Container)), true, Diagnostic{}
			}
		}
		return workspace.EnvValue{}, false, newInterpolationDiagnostic(
			"warning",
			"missing-export-placeholder",
			consumer,
			contract,
			provider.Key,
			envKey,
			fmt.Sprintf("provider %q exposes no container port %d required by export env %q", provider.Key, port, envKey),
		)
	default:
		return workspace.EnvValue{}, false, newInterpolationDiagnostic(
			"warning",
			"unknown-export-placeholder",
			consumer,
			contract,
			provider.Key,
			envKey,
			fmt.Sprintf("unknown placeholder %q in export env %q", token, envKey),
		)
	}
}

func newInterpolationDiagnostic(severity, code, consumer, contract, provider, envKey, message string) Diagnostic {
	return Diagnostic{
		Severity: severity,
		Code:     code,
		Consumer: consumer,
		Contract: contract,
		Provider: provider,
		EnvKey:   envKey,
		Message:  message,
	}
}
