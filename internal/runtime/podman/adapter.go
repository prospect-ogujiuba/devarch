package podman

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

type Adapter struct {
	runner runtimepkg.CommandRunner
}

func New(runner runtimepkg.CommandRunner) *Adapter {
	if runner == nil {
		runner = execRunner{}
	}
	return &Adapter{runner: runner}
}

func (a *Adapter) Provider() string {
	return runtimepkg.ProviderPodman
}

func (a *Adapter) Capabilities() runtimepkg.AdapterCapabilities {
	return runtimepkg.AdapterCapabilities{Inspect: true, Logs: true, Exec: true}
}

func (a *Adapter) InspectWorkspace(ctx context.Context, desired *runtimepkg.DesiredWorkspace) (*runtimepkg.Snapshot, error) {
	if desired == nil {
		return nil, fmt.Errorf("podman inspect workspace: nil desired workspace")
	}
	args := []string{"ps", "-aq", "--filter", fmt.Sprintf("label=%s=%s", runtimepkg.LabelWorkspace, desired.Name), "--filter", fmt.Sprintf("label=%s=%s", runtimepkg.LabelManagedBy, runtimepkg.ManagedByValue)}
	idsOutput, err := a.runner.Run(ctx, "podman", args...)
	if err != nil {
		return nil, err
	}

	var inspectOutput []byte
	ids := parseLines(idsOutput)
	if len(ids) > 0 {
		inspectOutput, err = a.runner.Run(ctx, "podman", append([]string{"inspect"}, ids...)...)
		if err != nil {
			return nil, err
		}
	}

	var networkOutput []byte
	if desired.Network != nil {
		networkOutput, err = a.runner.Run(ctx, "podman", "network", "inspect", desired.Network.Name)
		if err != nil && !isNotFoundError(err) {
			return nil, err
		}
		if isNotFoundError(err) {
			networkOutput = nil
		}
	}

	return runtimepkg.NormalizeInspectSnapshot(runtimepkg.ProviderPodman, desired, inspectOutput, networkOutput)
}

func (a *Adapter) EnsureNetwork(ctx context.Context, network *runtimepkg.DesiredNetwork) error {
	return unsupported("ensure-network")
}

func (a *Adapter) RemoveNetwork(ctx context.Context, network *runtimepkg.DesiredNetwork) error {
	return unsupported("remove-network")
}

func (a *Adapter) ApplyResource(ctx context.Context, request runtimepkg.ApplyResourceRequest) error {
	return unsupported("apply-resource")
}

func (a *Adapter) RemoveResource(ctx context.Context, resource runtimepkg.ResourceRef) error {
	return unsupported("remove-resource")
}

func (a *Adapter) RestartResource(ctx context.Context, resource runtimepkg.ResourceRef) error {
	return unsupported("restart-resource")
}

func (a *Adapter) StreamLogs(ctx context.Context, resource runtimepkg.ResourceRef, request runtimepkg.LogsRequest, consume runtimepkg.LogsConsumer) error {
	if consume == nil {
		return fmt.Errorf("podman logs: nil consumer")
	}
	args := []string{"logs", "--timestamps"}
	if request.Tail > 0 {
		args = append(args, "--tail", strconv.Itoa(request.Tail))
	}
	if request.Since != nil {
		args = append(args, "--since", request.Since.Format(timeLayout()))
	}
	if request.Follow {
		args = append(args, "--follow")
	}
	args = append(args, resource.RuntimeName)
	output, err := a.runner.Run(ctx, "podman", args...)
	if err != nil {
		return err
	}
	for _, chunk := range runtimepkg.ParseLogOutput("combined", output) {
		if err := consume(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) Exec(ctx context.Context, resource runtimepkg.ResourceRef, request runtimepkg.ExecRequest) (*runtimepkg.ExecResult, error) {
	if request.Interactive || request.TTY {
		return nil, unsupported("exec-interactive")
	}
	args := append([]string{"exec", resource.RuntimeName}, request.Command...)
	output, err := a.runner.Run(ctx, "podman", args...)
	if err != nil {
		return nil, err
	}
	return &runtimepkg.ExecResult{ExitCode: 0, Stdout: string(output)}, nil
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	return cmd.CombinedOutput()
}

func parseLines(output []byte) []string {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	parsed := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parsed = append(parsed, line)
	}
	if len(parsed) == 0 {
		return nil
	}
	return parsed
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "no such network") || strings.Contains(message, "not found")
}

func unsupported(operation string) error {
	return &runtimepkg.UnsupportedOperationError{Provider: runtimepkg.ProviderPodman, Operation: operation, Reason: "apply mutations are deferred behind explicit seams in Phase 3"}
}

func timeLayout() string {
	return "2006-01-02T15:04:05Z07:00"
}
