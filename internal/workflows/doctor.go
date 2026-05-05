package workflows

import (
	"context"
	"os"
	"path/filepath"
)

// DoctorReport reports local development health.
type DoctorReport struct {
	Status WorkflowStatus `json:"status"`
	Checks []CheckResult  `json:"checks"`
}

type DoctorOptions struct {
	WorkspaceRoots []string
	CatalogRoots   []string
	RootDir        string
}

func Doctor(ctx context.Context, runner Runner, opts DoctorOptions) (*DoctorReport, error) {
	if runner == nil {
		runner = ExecRunner{}
	}
	checks := []CheckResult{
		podmanAvailable(ctx, runner),
		SocketStatus(ctx, runner).Check,
		readableRoots("workspace-roots", "Workspace roots", opts.WorkspaceRoots),
		readableRoots("catalog-roots", "Catalog roots", opts.CatalogRoots),
		packageDiscovery(ctx, runner, opts.RootDir),
	}
	return &DoctorReport{Status: ReportStatus(checks), Checks: checks}, nil
}

func podmanAvailable(ctx context.Context, runner Runner) CheckResult {
	result := runner.Run(ctx, "podman", "--version")
	if result.Status == StatusPass {
		return CheckResult{ID: "podman.available", Name: "Podman available", Status: StatusPass, Message: firstNonEmpty(result.StdoutSummary, "podman found")}
	}
	return CheckResult{ID: "podman.available", Name: "Podman available", Status: StatusFail, Message: "podman is not available", Diagnostics: []Diagnostic{{ID: "podman.missing", Severity: StatusFail, Message: "install podman or fix PATH", Detail: result.Error}}}
}

func readableRoots(id, name string, roots []string) CheckResult {
	if len(roots) == 0 {
		return CheckResult{ID: id, Name: name, Status: StatusWarn, Message: "no roots configured"}
	}
	for _, root := range roots {
		if info, err := os.Stat(root); err != nil || !info.IsDir() {
			return CheckResult{ID: id, Name: name, Status: StatusFail, Message: root + " is not readable"}
		}
	}
	return CheckResult{ID: id, Name: name, Status: StatusPass, Message: "all roots readable"}
}

func packageDiscovery(ctx context.Context, runner Runner, rootDir string) CheckResult {
	if rootDir == "" {
		rootDir = "."
	}
	if !filepath.IsAbs(rootDir) {
		rootDir = "."
	}
	result := runner.Run(ctx, "go", "list", "./cmd/devarch", "./internal/appsvc", "./internal/runtime/...", "./internal/workflows/...")
	if result.Status == StatusPass {
		return CheckResult{ID: "go.discovery", Name: "Go package discovery", Status: StatusPass, Message: "root V2 packages discoverable"}
	}
	return CheckResult{ID: "go.discovery", Name: "Go package discovery", Status: StatusFail, Message: "root V2 package discovery failed", Diagnostics: []Diagnostic{{ID: "go.discovery.failed", Severity: StatusFail, Message: result.StderrSummary, Detail: result.Error}}}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
