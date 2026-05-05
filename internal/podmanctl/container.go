package podmanctl

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

type ContainerSpec struct {
	Name          string
	Image         string
	Command       []string
	Entrypoint    []string
	WorkingDir    string
	Env           map[string]workspace.EnvValue
	Ports         []PortSpec
	Volumes       []VolumeSpec
	Labels        map[string]string
	Network       string
	RestartPolicy string
	Health        *workspace.Health
}

type PortSpec struct {
	Container int
	Published int
	Protocol  string
	HostIP    string
}

type VolumeSpec struct {
	Source   string
	Target   string
	ReadOnly bool
	Kind     string
	Type     string
}

func BuildRunArgs(spec ContainerSpec) []string {
	args := []string{"run", "--detach", "--replace"}
	if spec.Name != "" {
		args = append(args, "--name", spec.Name)
	}
	if spec.WorkingDir != "" {
		args = append(args, "--workdir", spec.WorkingDir)
	}
	for _, entry := range spec.Entrypoint {
		args = append(args, "--entrypoint", entry)
	}
	for _, key := range sortedEnvKeys(spec.Env) {
		args = append(args, "--env", key+"="+spec.Env[key].Text())
	}
	ports := append([]PortSpec(nil), spec.Ports...)
	sort.SliceStable(ports, func(i, j int) bool { return portValue(ports[i]) < portValue(ports[j]) })
	for _, port := range ports {
		args = append(args, "--publish", portValue(port))
	}
	volumes := append([]VolumeSpec(nil), spec.Volumes...)
	sort.SliceStable(volumes, func(i, j int) bool { return volumeValue(volumes[i]) < volumeValue(volumes[j]) })
	for _, volume := range volumes {
		args = append(args, "--volume", volumeValue(volume))
	}
	for _, key := range sortedKeys(spec.Labels) {
		args = append(args, "--label", key+"="+spec.Labels[key])
	}
	if spec.Network != "" {
		args = append(args, "--network", spec.Network)
	}
	if spec.RestartPolicy != "" {
		args = append(args, "--restart", spec.RestartPolicy)
	}
	appendHealthArgs(&args, spec.Health)
	if spec.Image != "" {
		args = append(args, spec.Image)
	}
	args = append(args, spec.Command...)
	return args
}

// ApplyContainer uses podman run --replace so apply is idempotent without a
// separate remove/create race in the adapter layer.
func ApplyContainer(ctx context.Context, runner Runner, spec ContainerSpec) error {
	if spec.Image == "" {
		return fmt.Errorf("podman run %q: image is required", spec.Name)
	}
	output, err := Podman(ctx, runner, BuildRunArgs(spec)...)
	if err != nil {
		return fmt.Errorf("podman run %q: %w%s", spec.Name, err, outputSuffix(output))
	}
	return nil
}

func RemoveContainer(ctx context.Context, runner Runner, name string) error {
	output, err := Podman(ctx, runner, "rm", "--force", name)
	if err == nil || isNotFound(output, err) {
		return nil
	}
	return fmt.Errorf("podman rm %q: %w", name, err)
}

func RestartContainer(ctx context.Context, runner Runner, name string) error {
	if _, err := Podman(ctx, runner, "restart", name); err != nil {
		return fmt.Errorf("podman restart %q: %w", name, err)
	}
	return nil
}

func sortedEnvKeys(values map[string]workspace.EnvValue) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func portValue(port PortSpec) string {
	protocol := port.Protocol
	if protocol == "" {
		protocol = "tcp"
	}
	target := strconv.Itoa(port.Container)
	if protocol != "" {
		target += "/" + protocol
	}
	published := ""
	if port.Published > 0 {
		published = strconv.Itoa(port.Published)
	}
	if port.HostIP != "" {
		return port.HostIP + ":" + published + ":" + target
	}
	if published != "" {
		return published + ":" + target
	}
	return target
}

func volumeValue(volume VolumeSpec) string {
	parts := []string{volume.Target}
	if volume.Source != "" {
		parts = []string{volume.Source, volume.Target}
	}
	if volume.ReadOnly {
		parts = append(parts, "ro")
	}
	return strings.Join(parts, ":")
}

func appendHealthArgs(args *[]string, health *workspace.Health) {
	if health == nil {
		return
	}
	if len(health.Test) > 0 {
		*args = append(*args, "--health-cmd", strings.Join(health.Test, " "))
	}
	if health.Interval != "" {
		*args = append(*args, "--health-interval", health.Interval)
	}
	if health.Timeout != "" {
		*args = append(*args, "--health-timeout", health.Timeout)
	}
	if health.Retries > 0 {
		*args = append(*args, "--health-retries", strconv.Itoa(health.Retries))
	}
	if health.StartPeriod != "" {
		*args = append(*args, "--health-start-period", health.StartPeriod)
	}
}
