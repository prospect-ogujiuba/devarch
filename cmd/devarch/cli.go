package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/apply"
	"github.com/prospect-ogujiuba/devarch/internal/appsvc"
	planpkg "github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

type cliConfig struct {
	workspaceRoots []string
	catalogRoots   []string
	json           bool
}

type stringSliceFlag []string

func (f *stringSliceFlag) String() string {
	if f == nil {
		return ""
	}
	return strings.Join(*f, ",")
}

func (f *stringSliceFlag) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("flag value must not be empty")
	}
	*f = append(*f, trimmed)
	return nil
}

type serviceAPI interface {
	Doctor(context.Context) (*appsvc.DoctorReport, error)
	RuntimeStatus(context.Context) (*appsvc.RuntimeStatusReport, error)
	SocketStatus(context.Context) (*appsvc.SocketStatusReport, error)
	SocketStart(context.Context) (*appsvc.WorkflowCommandResult, error)
	SocketStop(context.Context) (*appsvc.WorkflowCommandResult, error)
	CatalogTemplates(context.Context) ([]appsvc.TemplateSummary, error)
	CatalogTemplate(context.Context, string) (*appsvc.TemplateDetail, error)
	Workspaces(context.Context) ([]appsvc.WorkspaceSummary, error)
	Workspace(context.Context, string) (*appsvc.WorkspaceDetail, error)
	WorkspacePlan(context.Context, string) (*planpkg.Result, error)
	ApplyWorkspace(context.Context, string) (*apply.Result, error)
	WorkspaceStatus(context.Context, string) (*appsvc.WorkspaceStatusView, error)
	WorkspaceLogs(context.Context, string, string, runtimepkg.LogsRequest) ([]runtimepkg.LogChunk, error)
	ExecWorkspace(context.Context, string, string, runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error)
	RestartWorkspaceResource(context.Context, string, string) error
	ScanProject(context.Context, string) (*appsvc.ProjectScanView, error)
	ImportV1Stack(context.Context, string) (*appsvc.ImportPreview, error)
	ImportV1Library(context.Context, string) (*appsvc.ImportPreview, error)
}

type serviceFactory func(cliConfig) (serviceAPI, error)

type exitStatusError struct {
	code int
}

func (e *exitStatusError) Error() string { return fmt.Sprintf("exit status %d", e.code) }
func (e *exitStatusError) ExitCode() int { return e.code }
func (e *exitStatusError) Silent() bool  { return true }

func defaultServiceFactory(cfg cliConfig) (serviceAPI, error) {
	return appsvc.New(appsvc.Config{
		WorkspaceRoots: cfg.workspaceRoots,
		CatalogRoots:   cfg.catalogRoots,
	})
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	cfg, rest, err := parseRootFlags(args, stderr)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if len(rest) == 0 {
		writeRootUsage(stderr)
		return fmt.Errorf("command is required")
	}
	if factory == nil {
		factory = defaultServiceFactory
	}

	switch rest[0] {
	case "doctor":
		return runDoctor(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "runtime":
		return runRuntime(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "socket":
		return runSocket(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "workspace":
		return runWorkspace(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "catalog":
		return runCatalog(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "import":
		return runImport(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "scan":
		return runScan(ctx, cfg, rest[1:], stdout, stderr, factory)
	case "help", "-h", "--help":
		writeRootUsage(stdout)
		return nil
	default:
		writeRootUsage(stderr)
		return fmt.Errorf("unknown command %q", rest[0])
	}
}

func parseRootFlags(args []string, stderr io.Writer) (cliConfig, []string, error) {
	cfg := cliConfig{}
	fs := flag.NewFlagSet("devarch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Var((*stringSliceFlag)(&cfg.workspaceRoots), "workspace-root", "Repeatable workspace root scanned recursively for devarch.workspace.yaml")
	fs.Var((*stringSliceFlag)(&cfg.catalogRoots), "catalog-root", "Repeatable catalog root scanned for template.yaml")
	fs.BoolVar(&cfg.json, "json", false, "Emit stable JSON output (place before the command)")
	fs.Usage = func() { writeRootUsage(stderr) }
	if err := fs.Parse(args); err != nil {
		return cliConfig{}, nil, err
	}
	return cfg, fs.Args(), nil
}

func runDoctor(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "Usage: devarch [global flags] doctor")
		return fmt.Errorf("doctor does not accept positional arguments")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}
	report, err := svc.Doctor(ctx)
	if err != nil {
		return err
	}
	if cfg.json {
		return writeJSON(stdout, report)
	}
	printChecks(stdout, "Doctor", report.Status, report.Checks)
	return nil
}

func runRuntime(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(args) != 1 || args[0] != "status" {
		fmt.Fprintln(stderr, "Usage: devarch [global flags] runtime status")
		return fmt.Errorf("runtime status requires no positional arguments")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}
	report, err := svc.RuntimeStatus(ctx)
	if err != nil {
		return err
	}
	if cfg.json {
		return writeJSON(stdout, report)
	}
	printChecks(stdout, "Runtime", report.Status, report.Checks)
	return nil
}

func runSocket(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(args) != 1 {
		writeSocketUsage(stderr)
		return fmt.Errorf("socket subcommand is required")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}
	switch args[0] {
	case "status":
		report, err := svc.SocketStatus(ctx)
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, report)
		}
		printChecks(stdout, "Socket", report.Status, []appsvc.WorkflowCheckResult{report.Check})
		return nil
	case "start":
		result, err := svc.SocketStart(ctx)
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, result)
		}
		printCommandResult(stdout, result)
		return nil
	case "stop":
		result, err := svc.SocketStop(ctx)
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, result)
		}
		printCommandResult(stdout, result)
		return nil
	case "help", "-h", "--help":
		writeSocketUsage(stdout)
		return nil
	default:
		writeSocketUsage(stderr)
		return fmt.Errorf("unknown socket subcommand %q", args[0])
	}
}

func runWorkspace(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(cfg.workspaceRoots) == 0 {
		return fmt.Errorf("workspace commands require at least one --workspace-root")
	}
	if len(args) == 0 {
		writeWorkspaceUsage(stderr)
		return fmt.Errorf("workspace subcommand is required")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			writeWorkspaceUsage(stderr)
			return fmt.Errorf("workspace list does not accept positional arguments")
		}
		workspaces, err := svc.Workspaces(ctx)
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, workspaces)
		}
		printWorkspaceList(stdout, workspaces)
		return nil
	case "open":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace open <name>")
			return fmt.Errorf("workspace open requires <name>")
		}
		workspace, err := svc.Workspace(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, workspace)
		}
		printWorkspaceDetail(stdout, workspace)
		return nil
	case "plan":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace plan <name>")
			return fmt.Errorf("workspace plan requires <name>")
		}
		plan, err := svc.WorkspacePlan(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, plan)
		}
		printPlan(stdout, plan)
		return nil
	case "apply":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace apply <name>")
			return fmt.Errorf("workspace apply requires <name>")
		}
		result, err := svc.ApplyWorkspace(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, result)
		}
		printApply(stdout, result)
		return nil
	case "status":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace status <name>")
			return fmt.Errorf("workspace status requires <name>")
		}
		status, err := svc.WorkspaceStatus(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, status)
		}
		printStatus(stdout, status)
		return nil
	case "logs":
		return runWorkspaceLogs(ctx, cfg, svc, args[1:], stdout, stderr)
	case "exec":
		return runWorkspaceExec(ctx, cfg, svc, args[1:], stdout, stderr)
	case "restart":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace restart <name> <resource>")
			return fmt.Errorf("workspace restart requires <name> and <resource>")
		}
		if err := svc.RestartWorkspaceResource(ctx, args[1], args[2]); err != nil {
			return err
		}
		result := map[string]string{"workspace": args[1], "resource": args[2], "status": "restarted"}
		if cfg.json {
			return writeJSON(stdout, result)
		}
		fmt.Fprintf(stdout, "Restarted %s/%s\n", args[1], args[2])
		return nil
	case "help", "-h", "--help":
		writeWorkspaceUsage(stdout)
		return nil
	default:
		writeWorkspaceUsage(stderr)
		return fmt.Errorf("unknown workspace subcommand %q", args[0])
	}
}

func runWorkspaceLogs(ctx context.Context, cfg cliConfig, svc serviceAPI, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("devarch workspace logs", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var tail int
	var sinceRaw string
	var follow bool
	fs.IntVar(&tail, "tail", 0, "Show the last N lines")
	fs.StringVar(&sinceRaw, "since", "", "Filter logs since RFC3339 timestamp")
	fs.BoolVar(&follow, "follow", false, "Follow log output until interrupted")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace logs [--tail N] [--since RFC3339] [--follow] <name> <resource>")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) != 2 {
		fs.Usage()
		return fmt.Errorf("workspace logs requires <name> and <resource>")
	}
	request := runtimepkg.LogsRequest{Tail: tail, Follow: follow}
	if sinceRaw != "" {
		since, err := time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			return fmt.Errorf("parse --since: %w", err)
		}
		request.Since = &since
	}
	chunks, err := svc.WorkspaceLogs(ctx, fs.Arg(0), fs.Arg(1), request)
	if err != nil {
		return err
	}
	if cfg.json {
		return writeJSON(stdout, chunks)
	}
	printLogs(stdout, chunks)
	return nil
}

func runWorkspaceExec(ctx context.Context, cfg cliConfig, svc serviceAPI, args []string, stdout, stderr io.Writer) error {
	if len(args) < 3 {
		fmt.Fprintln(stderr, "Usage: devarch [global flags] workspace exec <name> <resource> [--] <command...>")
		return fmt.Errorf("workspace exec requires <name> <resource> and <command...>")
	}
	name := args[0]
	resource := args[1]
	command := append([]string(nil), args[2:]...)
	if len(command) > 0 && command[0] == "--" {
		command = command[1:]
	}
	if len(command) == 0 {
		return fmt.Errorf("workspace exec requires <command...>")
	}
	result, err := svc.ExecWorkspace(ctx, name, resource, runtimepkg.ExecRequest{Command: command})
	if err != nil {
		return err
	}
	if cfg.json {
		if err := writeJSON(stdout, result); err != nil {
			return err
		}
	} else {
		printExecResult(stdout, stderr, result)
	}
	if result != nil && result.ExitCode != 0 {
		return &exitStatusError{code: result.ExitCode}
	}
	return nil
}

func runCatalog(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(cfg.catalogRoots) == 0 {
		return fmt.Errorf("catalog commands require at least one --catalog-root")
	}
	if len(args) == 0 {
		writeCatalogUsage(stderr)
		return fmt.Errorf("catalog subcommand is required")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] catalog list")
			return fmt.Errorf("catalog list does not accept positional arguments")
		}
		templates, err := svc.CatalogTemplates(ctx)
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, templates)
		}
		printCatalogList(stdout, templates)
		return nil
	case "show":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] catalog show <template>")
			return fmt.Errorf("catalog show requires <template>")
		}
		template, err := svc.CatalogTemplate(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, template)
		}
		printCatalogDetail(stdout, template)
		return nil
	case "help", "-h", "--help":
		writeCatalogUsage(stdout)
		return nil
	default:
		writeCatalogUsage(stderr)
		return fmt.Errorf("unknown catalog subcommand %q", args[0])
	}
}

func runImport(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(args) == 0 {
		writeImportUsage(stderr)
		return fmt.Errorf("import subcommand is required")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "v1-stack":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] import v1-stack <file>")
			return fmt.Errorf("import v1-stack requires <file>")
		}
		preview, err := svc.ImportV1Stack(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, preview)
		}
		printImportPreview(stdout, preview)
		return nil
	case "v1-library":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] import v1-library <path>")
			return fmt.Errorf("import v1-library requires <path>")
		}
		preview, err := svc.ImportV1Library(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, preview)
		}
		printImportPreview(stdout, preview)
		return nil
	case "help", "-h", "--help":
		writeImportUsage(stdout)
		return nil
	default:
		writeImportUsage(stderr)
		return fmt.Errorf("unknown import subcommand %q", args[0])
	}
}

func runScan(ctx context.Context, cfg cliConfig, args []string, stdout, stderr io.Writer, factory serviceFactory) error {
	if len(args) == 0 {
		writeScanUsage(stderr)
		return fmt.Errorf("scan subcommand is required")
	}
	svc, err := factory(cfg)
	if err != nil {
		return err
	}

	switch args[0] {
	case "project":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Usage: devarch [global flags] scan project <path>")
			return fmt.Errorf("scan project requires <path>")
		}
		result, err := svc.ScanProject(ctx, args[1])
		if err != nil {
			return err
		}
		if cfg.json {
			return writeJSON(stdout, result)
		}
		printScanResult(stdout, result)
		return nil
	case "help", "-h", "--help":
		writeScanUsage(stdout)
		return nil
	default:
		writeScanUsage(stderr)
		return fmt.Errorf("unknown scan subcommand %q", args[0])
	}
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func printChecks(w io.Writer, title string, status appsvc.WorkflowStatus, checks []appsvc.WorkflowCheckResult) {
	fmt.Fprintf(w, "%s status: %s\n", title, status)
	if len(checks) == 0 {
		fmt.Fprintln(w, "Checks: none")
		return
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "ID\tSTATUS\tMESSAGE")
	for _, check := range checks {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", check.ID, check.Status, orDash(check.Message))
	}
	_ = tw.Flush()
}

func printCommandResult(w io.Writer, result *appsvc.WorkflowCommandResult) {
	if result == nil {
		fmt.Fprintln(w, "No command result.")
		return
	}
	fmt.Fprintf(w, "Command: %s %s\n", result.Command, strings.Join(result.Args, " "))
	fmt.Fprintf(w, "Status: %s\n", result.Status)
	if result.StdoutSummary != "" {
		fmt.Fprintf(w, "Stdout: %s\n", result.StdoutSummary)
	}
	if result.StderrSummary != "" {
		fmt.Fprintf(w, "Stderr: %s\n", result.StderrSummary)
	}
	if result.Error != "" {
		fmt.Fprintf(w, "Error: %s\n", result.Error)
	}
}

func printWorkspaceList(w io.Writer, workspaces []appsvc.WorkspaceSummary) {
	if len(workspaces) == 0 {
		fmt.Fprintln(w, "No workspaces found.")
		return
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDISPLAY NAME\tPROVIDER\tRESOURCES\tCAPABILITIES")
	for _, workspace := range workspaces {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", workspace.Name, orDash(workspace.DisplayName), orDash(workspace.Provider), workspace.ResourceCount, orDash(capabilitiesText(workspace.Capabilities)))
	}
	_ = tw.Flush()
}

func printWorkspaceDetail(w io.Writer, workspace *appsvc.WorkspaceDetail) {
	if workspace == nil {
		fmt.Fprintln(w, "No workspace data.")
		return
	}
	fmt.Fprintf(w, "Name: %s\n", workspace.Name)
	if workspace.DisplayName != "" {
		fmt.Fprintf(w, "Display name: %s\n", workspace.DisplayName)
	}
	if workspace.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", workspace.Description)
	}
	fmt.Fprintf(w, "Provider: %s\n", orDash(workspace.Provider))
	fmt.Fprintf(w, "Manifest: %s\n", workspace.ManifestPath)
	fmt.Fprintf(w, "Resources (%d): %s\n", workspace.ResourceCount, strings.Join(workspace.ResourceKeys, ", "))
	if capabilityText := capabilitiesText(workspace.Capabilities); capabilityText != "" {
		fmt.Fprintf(w, "Capabilities: %s\n", capabilityText)
	}
}

func printPlan(w io.Writer, plan *planpkg.Result) {
	if plan == nil {
		fmt.Fprintln(w, "No plan available.")
		return
	}
	fmt.Fprintf(w, "Workspace: %s\n", plan.Workspace)
	fmt.Fprintf(w, "Provider: %s\n", orDash(plan.Provider))
	fmt.Fprintf(w, "Blocked: %t\n", plan.Blocked)
	printRuntimeDiagnostics(w, plan.Diagnostics)
	if len(plan.Actions) == 0 {
		fmt.Fprintln(w, "Actions: none")
		return
	}
	fmt.Fprintln(w, "Actions:")
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "SCOPE\tTARGET\tKIND\tRUNTIME NAME\tREASONS")
	for _, action := range plan.Actions {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", action.Scope, action.Target, action.Kind, orDash(action.RuntimeName), orDash(strings.Join(action.Reasons, "; ")))
	}
	_ = tw.Flush()
}

func printApply(w io.Writer, result *apply.Result) {
	if result == nil {
		fmt.Fprintln(w, "No apply result.")
		return
	}
	fmt.Fprintf(w, "Workspace: %s\n", result.Workspace)
	fmt.Fprintf(w, "Provider: %s\n", orDash(result.Provider))
	fmt.Fprintf(w, "Started: %s\n", result.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Finished: %s\n", result.FinishedAt.Format(time.RFC3339))
	if len(result.Operations) == 0 {
		fmt.Fprintln(w, "Operations: none")
	} else {
		fmt.Fprintln(w, "Operations:")
		tw := newTabWriter(w)
		fmt.Fprintln(tw, "SCOPE\tTARGET\tKIND\tSTATUS\tMESSAGE")
		for _, operation := range result.Operations {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", operation.Scope, operation.Target, operation.Kind, operation.Status, orDash(operation.Message))
		}
		_ = tw.Flush()
	}
	if result.Snapshot != nil {
		fmt.Fprintf(w, "Snapshot resources: %d\n", len(result.Snapshot.Resources))
	}
}

func printStatus(w io.Writer, status *appsvc.WorkspaceStatusView) {
	if status == nil || status.Desired == nil {
		fmt.Fprintln(w, "No workspace status available.")
		return
	}
	fmt.Fprintf(w, "Workspace: %s\n", status.Desired.Name)
	fmt.Fprintf(w, "Provider: %s\n", orDash(status.Desired.Provider))
	if status.Desired.Network != nil {
		fmt.Fprintf(w, "Network: %s\n", status.Desired.Network.Name)
	}
	printRuntimeDiagnostics(w, status.Desired.Diagnostics)
	if len(status.Desired.Resources) == 0 {
		fmt.Fprintln(w, "Resources: none")
		return
	}
	resources := append([]*runtimepkg.DesiredResource(nil), status.Desired.Resources...)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Key < resources[j].Key
	})
	fmt.Fprintln(w, "Resources:")
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "KEY\tRUNTIME NAME\tSTATUS\tHEALTH\tIMAGE")
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		state := "absent"
		health := "-"
		if snapshot := status.Snapshot; snapshot != nil {
			if observed := snapshot.Resource(resource.Key); observed != nil {
				if observed.State.Status != "" {
					state = observed.State.Status
				}
				if observed.State.Health != "" {
					health = observed.State.Health
				}
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", resource.Key, resource.RuntimeName, state, health, orDash(resource.Spec.Image))
	}
	_ = tw.Flush()
}

func printLogs(w io.Writer, chunks []runtimepkg.LogChunk) {
	if len(chunks) == 0 {
		fmt.Fprintln(w, "No log output.")
		return
	}
	for _, chunk := range chunks {
		parts := make([]string, 0, 3)
		if chunk.Timestamp != nil {
			parts = append(parts, chunk.Timestamp.Format(time.RFC3339))
		}
		if chunk.Stream != "" {
			parts = append(parts, "["+chunk.Stream+"]")
		}
		parts = append(parts, chunk.Line)
		fmt.Fprintln(w, strings.Join(parts, " "))
	}
}

func printExecResult(stdout, stderr io.Writer, result *runtimepkg.ExecResult) {
	if result == nil {
		return
	}
	if result.Stdout != "" {
		_, _ = io.WriteString(stdout, result.Stdout)
		if !strings.HasSuffix(result.Stdout, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
	}
	if result.Stderr != "" {
		_, _ = io.WriteString(stderr, result.Stderr)
		if !strings.HasSuffix(result.Stderr, "\n") {
			_, _ = io.WriteString(stderr, "\n")
		}
	}
	if result.ExitCode != 0 && result.Stdout == "" && result.Stderr == "" {
		fmt.Fprintf(stderr, "exit code: %d\n", result.ExitCode)
	}
}

func printCatalogList(w io.Writer, templates []appsvc.TemplateSummary) {
	if len(templates) == 0 {
		fmt.Fprintln(w, "No catalog templates found.")
		return
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tTAGS")
	for _, template := range templates {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", template.Name, orDash(template.Description), orDash(strings.Join(template.Tags, ", ")))
	}
	_ = tw.Flush()
}

func printCatalogDetail(w io.Writer, template *appsvc.TemplateDetail) {
	if template == nil {
		fmt.Fprintln(w, "No template data.")
		return
	}
	fmt.Fprintf(w, "Name: %s\n", template.Name)
	if template.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", template.Description)
	}
	if template.Kind != "" {
		fmt.Fprintf(w, "Kind: %s\n", template.Kind)
	}
	if template.APIVersion != "" {
		fmt.Fprintf(w, "API version: %s\n", template.APIVersion)
	}
	if len(template.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(template.Tags, ", "))
	}
	printStructuredBlock(w, "Runtime", template.Runtime)
	printStructuredBlock(w, "Env", template.Env)
	printStructuredBlock(w, "Ports", template.Ports)
	printStructuredBlock(w, "Volumes", template.Volumes)
	printStructuredBlock(w, "Imports", template.Imports)
	printStructuredBlock(w, "Exports", template.Exports)
	printStructuredBlock(w, "Health", template.Health)
	printStructuredBlock(w, "Develop", template.Develop)
}

func printImportPreview(w io.Writer, preview *appsvc.ImportPreview) {
	if preview == nil {
		fmt.Fprintln(w, "No import result.")
		return
	}
	fmt.Fprintf(w, "Mode: %s\n", preview.Mode)
	fmt.Fprintf(w, "Source: %s\n", preview.SourcePath)
	fmt.Fprintf(w, "Status: %s\n", preview.Status)
	if preview.Message != "" {
		fmt.Fprintf(w, "Message: %s\n", preview.Message)
	}
	if preview.Summary.Total > 0 {
		fmt.Fprintf(w, "Artifacts: %d total (%d succeeded, %d partial, %d rejected)\n", preview.Summary.Total, preview.Summary.Succeeded, preview.Summary.Partial, preview.Summary.Rejected)
	}
	if len(preview.Diagnostics) > 0 {
		fmt.Fprintln(w, "Diagnostics:")
		for _, diagnostic := range preview.Diagnostics {
			fmt.Fprintf(w, "- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
		}
	}
	if len(preview.Artifacts) == 0 {
		return
	}
	fmt.Fprintln(w, "Artifacts:")
	for _, artifact := range preview.Artifacts {
		fmt.Fprintf(w, "- [%s] %s %s\n", artifact.Status, artifact.Kind, artifact.Name)
		if artifact.SuggestedPath != "" {
			fmt.Fprintf(w, "  Path: %s\n", artifact.SuggestedPath)
		}
		for _, diagnostic := range artifact.Diagnostics {
			fmt.Fprintf(w, "  - [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
		}
	}
}

func printScanResult(w io.Writer, result *appsvc.ProjectScanView) {
	if result == nil {
		fmt.Fprintln(w, "No scan result.")
		return
	}
	fmt.Fprintf(w, "Project: %s\n", result.Name)
	fmt.Fprintf(w, "Path: %s\n", result.Path)
	fmt.Fprintf(w, "Type: %s\n", orDash(result.ProjectType))
	if result.Framework != "" {
		fmt.Fprintf(w, "Framework: %s\n", result.Framework)
	}
	if result.Language != "" {
		fmt.Fprintf(w, "Language: %s\n", result.Language)
	}
	if result.PackageManager != "" {
		fmt.Fprintf(w, "Package manager: %s\n", result.PackageManager)
	}
	if result.EntryPoint != "" {
		fmt.Fprintf(w, "Entry point: %s\n", result.EntryPoint)
	}
	if len(result.ComposeFiles) > 0 {
		fmt.Fprintf(w, "Compose files: %s\n", strings.Join(result.ComposeFiles, ", "))
	}
	if len(result.SuggestedTemplates) > 0 {
		fmt.Fprintf(w, "Suggested templates: %s\n", strings.Join(result.SuggestedTemplates, ", "))
	}
	if len(result.Services) > 0 {
		fmt.Fprintln(w, "Compose services:")
		tw := newTabWriter(w)
		fmt.Fprintln(tw, "NAME\tTYPE\tIMAGE\tPORTS\tDEPENDS ON")
		for _, service := range result.Services {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", service.Name, orDash(service.ServiceType), orDash(service.Image), orDash(strings.Join(service.Ports, ", ")), orDash(strings.Join(service.DependsOn, ", ")))
		}
		_ = tw.Flush()
	}
	if len(result.Diagnostics) > 0 {
		fmt.Fprintln(w, "Diagnostics:")
		for _, diagnostic := range result.Diagnostics {
			fmt.Fprintf(w, "- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
		}
	}
}

func printRuntimeDiagnostics(w io.Writer, diagnostics []runtimepkg.Diagnostic) {
	if len(diagnostics) == 0 {
		return
	}
	fmt.Fprintln(w, "Diagnostics:")
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(w, "- [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
	}
}

func printStructuredBlock(w io.Writer, title string, value any) {
	if isZeroValue(value) {
		return
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "%s: <unavailable>\n", title)
		return
	}
	fmt.Fprintf(w, "%s:\n", title)
	for _, line := range strings.Split(string(data), "\n") {
		fmt.Fprintf(w, "  %s\n", line)
	}
}

func isZeroValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return typed == ""
	case []string:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	case map[string]string:
		return len(typed) == 0
	default:
		return false
	}
}

func capabilitiesText(capabilities runtimepkg.AdapterCapabilities) string {
	values := make([]string, 0, 5)
	if capabilities.Inspect {
		values = append(values, "inspect")
	}
	if capabilities.Apply {
		values = append(values, "apply")
	}
	if capabilities.Logs {
		values = append(values, "logs")
	}
	if capabilities.Exec {
		values = append(values, "exec")
	}
	if capabilities.Network {
		values = append(values, "network")
	}
	return strings.Join(values, ",")
}

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
}

func writeRootUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: devarch [--workspace-root PATH ...] [--catalog-root PATH ...] [--json] <command> ...")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  workspace list")
	fmt.Fprintln(w, "  workspace open <name>")
	fmt.Fprintln(w, "  workspace plan <name>")
	fmt.Fprintln(w, "  workspace apply <name>")
	fmt.Fprintln(w, "  workspace status <name>")
	fmt.Fprintln(w, "  workspace logs [--tail N] [--since RFC3339] [--follow] <name> <resource>")
	fmt.Fprintln(w, "  workspace exec <name> <resource> [--] <command...>")
	fmt.Fprintln(w, "  workspace restart <name> <resource>")
	fmt.Fprintln(w, "  doctor")
	fmt.Fprintln(w, "  runtime status")
	fmt.Fprintln(w, "  socket status")
	fmt.Fprintln(w, "  socket start")
	fmt.Fprintln(w, "  socket stop")
	fmt.Fprintln(w, "  catalog list")
	fmt.Fprintln(w, "  catalog show <template>")
	fmt.Fprintln(w, "  import v1-stack <file>")
	fmt.Fprintln(w, "  import v1-library <path>")
	fmt.Fprintln(w, "  scan project <path>")
}

func writeWorkspaceUsage(w io.Writer) {
	fmt.Fprintln(w, "Workspace commands:")
	fmt.Fprintln(w, "  devarch [global flags] workspace list")
	fmt.Fprintln(w, "  devarch [global flags] workspace open <name>")
	fmt.Fprintln(w, "  devarch [global flags] workspace plan <name>")
	fmt.Fprintln(w, "  devarch [global flags] workspace apply <name>")
	fmt.Fprintln(w, "  devarch [global flags] workspace status <name>")
	fmt.Fprintln(w, "  devarch [global flags] workspace logs [--tail N] [--since RFC3339] [--follow] <name> <resource>")
	fmt.Fprintln(w, "  devarch [global flags] workspace exec <name> <resource> [--] <command...>")
	fmt.Fprintln(w, "  devarch [global flags] workspace restart <name> <resource>")
}

func writeSocketUsage(w io.Writer) {
	fmt.Fprintln(w, "Socket commands:")
	fmt.Fprintln(w, "  devarch [global flags] socket status")
	fmt.Fprintln(w, "  devarch [global flags] socket start")
	fmt.Fprintln(w, "  devarch [global flags] socket stop")
}

func writeCatalogUsage(w io.Writer) {
	fmt.Fprintln(w, "Catalog commands:")
	fmt.Fprintln(w, "  devarch [global flags] catalog list")
	fmt.Fprintln(w, "  devarch [global flags] catalog show <template>")
}

func writeImportUsage(w io.Writer) {
	fmt.Fprintln(w, "Import commands:")
	fmt.Fprintln(w, "  devarch [global flags] import v1-stack <file>")
	fmt.Fprintln(w, "  devarch [global flags] import v1-library <path>")
}

func writeScanUsage(w io.Writer) {
	fmt.Fprintln(w, "Scan commands:")
	fmt.Fprintln(w, "  devarch [global flags] scan project <path>")
}
