package contracts

import "github.com/prospect-ogujiuba/devarch/internal/workspace"

// Result captures deterministic contract links and diagnostics for one resolved
// effective graph.
type Result struct {
	Links       []Link       `json:"links,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
}

type Link struct {
	Consumer    string                        `json:"consumer"`
	Contract    string                        `json:"contract"`
	Alias       string                        `json:"alias,omitempty"`
	Provider    string                        `json:"provider"`
	Source      string                        `json:"source"`
	InjectedEnv map[string]workspace.EnvValue `json:"injectedEnv,omitempty"`
}

type Diagnostic struct {
	Severity  string   `json:"severity"`
	Code      string   `json:"code"`
	Consumer  string   `json:"consumer,omitempty"`
	Contract  string   `json:"contract,omitempty"`
	Provider  string   `json:"provider,omitempty"`
	Providers []string `json:"providers,omitempty"`
	EnvKey    string   `json:"envKey,omitempty"`
	Message   string   `json:"message"`
}
