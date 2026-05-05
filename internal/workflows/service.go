package workflows

import (
	"context"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

// WorkspaceRuntimeOps is the appsvc/runtime boundary needed to express legacy
// service-manager intents in V2 workspace/resource terms.
type WorkspaceRuntimeOps interface {
	WorkspaceStatus(context.Context, string) (any, error)
	ApplyWorkspace(context.Context, string) (*apply.Result, error)
	WorkspaceLogs(context.Context, string, string, runtimepkg.LogsRequest) ([]runtimepkg.LogChunk, error)
	ExecWorkspace(context.Context, string, string, runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error)
	RestartWorkspaceResource(context.Context, string, string) error
}

func ServiceManagerIntents() map[string]string {
	return map[string]string{
		"status":  "workspace status",
		"apply":   "workspace apply",
		"logs":    "resource logs",
		"exec":    "resource exec",
		"restart": "resource restart",
	}
}
