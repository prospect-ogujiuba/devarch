package plan

import runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"

type ActionKind string

type ActionScope string

const (
	ActionAdd     ActionKind = "add"
	ActionModify  ActionKind = "modify"
	ActionRemove  ActionKind = "remove"
	ActionRestart ActionKind = "restart"
	ActionNoop    ActionKind = "noop"

	ScopeWorkspace ActionScope = "workspace"
	ScopeResource  ActionScope = "resource"
)

type Result struct {
	Workspace   string                  `json:"workspace"`
	Provider    string                  `json:"provider,omitempty"`
	Blocked     bool                    `json:"blocked,omitempty"`
	Diagnostics []runtimepkg.Diagnostic `json:"diagnostics,omitempty"`
	Actions     []Action                `json:"actions,omitempty"`
}

type Action struct {
	Scope       ActionScope `json:"scope"`
	Target      string      `json:"target"`
	RuntimeName string      `json:"runtimeName,omitempty"`
	Kind        ActionKind  `json:"kind"`
	Reasons     []string    `json:"reasons,omitempty"`
}
