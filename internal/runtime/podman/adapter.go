package podman

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/podmanctl"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

type Adapter struct {
	runner podmanctl.Runner
}

func New(runner podmanctl.Runner) *Adapter {
	if runner == nil {
		runner = podmanctl.ExecRunner{}
	}
	return &Adapter{runner: runner}
}

func (a *Adapter) Provider() string {
	return runtimepkg.ProviderPodman
}

func (a *Adapter) Capabilities() runtimepkg.AdapterCapabilities {
	return runtimepkg.AdapterCapabilities{Inspect: true, Apply: true, Logs: true, Exec: true, Network: true}
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
	if network == nil || network.Name == "" {
		return fmt.Errorf("podman ensure-network: network name is required")
	}
	return podmanctl.EnsureNetwork(ctx, a.runner, network.Name, network.Labels)
}

func (a *Adapter) RemoveNetwork(ctx context.Context, network *runtimepkg.DesiredNetwork) error {
	if network == nil || network.Name == "" {
		return fmt.Errorf("podman remove-network: network name is required")
	}
	return podmanctl.RemoveNetwork(ctx, a.runner, network.Name)
}

func (a *Adapter) ApplyResource(ctx context.Context, request runtimepkg.ApplyResourceRequest) error {
	spec, err := containerSpecFromRequest(request)
	if err != nil {
		return err
	}
	return podmanctl.ApplyContainer(ctx, a.runner, spec)
}

func (a *Adapter) RemoveResource(ctx context.Context, resource runtimepkg.ResourceRef) error {
	if resource.RuntimeName == "" {
		return fmt.Errorf("podman remove-resource: runtime name is required")
	}
	return podmanctl.RemoveContainer(ctx, a.runner, resource.RuntimeName)
}

func (a *Adapter) RestartResource(ctx context.Context, resource runtimepkg.ResourceRef) error {
	if resource.RuntimeName == "" {
		return fmt.Errorf("podman restart-resource: runtime name is required")
	}
	return podmanctl.RestartContainer(ctx, a.runner, resource.RuntimeName)
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

func containerSpecFromRequest(request runtimepkg.ApplyResourceRequest) (podmanctl.ContainerSpec, error) {
	resource := request.Resource
	if resource.RuntimeName == "" {
		return podmanctl.ContainerSpec{}, fmt.Errorf("podman apply-resource: runtime name is required")
	}
	if resource.Spec.Build != nil && resource.Spec.Image == "" {
		return podmanctl.ContainerSpec{}, unsupported("apply-resource-build")
	}
	if resource.Spec.ProjectSource != nil {
		return podmanctl.ContainerSpec{}, unsupported("apply-resource-project-source")
	}
	if len(resource.Spec.DevelopWatch) > 0 {
		return podmanctl.ContainerSpec{}, unsupported("apply-resource-develop-watch")
	}

	spec := podmanctl.ContainerSpec{
		Name:          resource.RuntimeName,
		Image:         resource.Spec.Image,
		Command:       append([]string(nil), resource.Spec.Command...),
		Entrypoint:    append([]string(nil), resource.Spec.Entrypoint...),
		WorkingDir:    resource.Spec.WorkingDir,
		Env:           resource.Spec.Env,
		Labels:        cloneStringMap(resource.Spec.Labels),
		Network:       request.NetworkName,
		RestartPolicy: "unless-stopped",
		Health:        resource.Spec.Health,
	}
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	if request.Workspace != "" {
		spec.Labels[runtimepkg.LabelWorkspace] = request.Workspace
	}
	if resource.Key != "" {
		spec.Labels[runtimepkg.LabelResource] = resource.Key
	}
	if resource.LogicalHost != "" {
		spec.Labels[runtimepkg.LabelHostAlias] = resource.LogicalHost
	}
	if request.NetworkName != "" {
		spec.Labels[runtimepkg.LabelNetwork] = request.NetworkName
	}
	for _, port := range resource.Spec.Ports {
		spec.Ports = append(spec.Ports, podmanctl.PortSpec{Container: port.Container, Published: port.Published, Protocol: port.Protocol, HostIP: port.HostIP})
	}
	for _, volume := range resource.Spec.Volumes {
		spec.Volumes = append(spec.Volumes, podmanctl.VolumeSpec{Source: volume.Source, Target: volume.Target, ReadOnly: volume.ReadOnly, Kind: volume.Kind, Type: volume.Type})
	}
	return spec, nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
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
	return strings.Contains(message, "no such network") || strings.Contains(message, "not found") || strings.Contains(message, "exit status 125")
}

func unsupported(operation string) error {
	return &runtimepkg.UnsupportedOperationError{Provider: runtimepkg.ProviderPodman, Operation: operation, Reason: "field cannot be safely mapped to podman run yet"}
}

func timeLayout() string {
	return "2006-01-02T15:04:05Z07:00"
}
