package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/priz/devarch-api/pkg/models"
)

var _ Runtime = (*Client)(nil)

var ErrNotImplemented = fmt.Errorf("not implemented: will be available in Phase 4")

type Client struct {
	runtime     RuntimeType
	composeCmd  string
	useSudo     bool
	networkName string
}

func NewClient() (*Client, error) {
	c := &Client{
		networkName: "microservices-net",
	}

	// Check DEVARCH_RUNTIME env var first
	if envRuntime := os.Getenv("DEVARCH_RUNTIME"); envRuntime != "" {
		switch envRuntime {
		case "podman":
			if _, err := exec.LookPath("podman"); err != nil {
				return nil, fmt.Errorf("DEVARCH_RUNTIME=podman but podman not found — install podman or unset DEVARCH_RUNTIME for auto-detection")
			}
			c.runtime = RuntimePodman
			if c.checkComposeAvailable("podman", "compose") {
				c.composeCmd = "compose"
			} else {
				c.composeCmd = "podman-compose"
			}
		case "docker":
			if _, err := exec.LookPath("docker"); err != nil {
				return nil, fmt.Errorf("DEVARCH_RUNTIME=docker but docker not found — install docker or unset DEVARCH_RUNTIME for auto-detection")
			}
			c.runtime = RuntimeDocker
			c.composeCmd = "compose"
		default:
			return nil, fmt.Errorf("DEVARCH_RUNTIME=%s invalid — must be 'podman' or 'docker'", envRuntime)
		}
	} else {
		// Auto-detect: prefer Podman
		if _, err := exec.LookPath("podman"); err == nil {
			c.runtime = RuntimePodman
			if c.checkComposeAvailable("podman", "compose") {
				c.composeCmd = "compose"
			} else {
				c.composeCmd = "podman-compose"
			}
		} else if _, err := exec.LookPath("docker"); err == nil {
			c.runtime = RuntimeDocker
			c.composeCmd = "compose"
		} else {
			return nil, fmt.Errorf("no container runtime found")
		}
	}

	c.useSudo = os.Getenv("DEVARCH_USE_SUDO") == "true"

	return c, nil
}

func (c *Client) checkComposeAvailable(runtime, subcmd string) bool {
	cmd := exec.Command(runtime, subcmd, "version")
	return cmd.Run() == nil
}

func (c *Client) RuntimeName() RuntimeType {
	return c.runtime
}

func (c *Client) execCommand(args ...string) (string, error) {
	var cmd *exec.Cmd
	if c.useSudo {
		sudoArgs := append([]string{string(c.runtime)}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
	} else {
		cmd = exec.Command(string(c.runtime), args...)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func (c *Client) execComposeCommand(composeYAML []byte, args ...string) error {
	tmpFile, err := os.CreateTemp("", "devarch-compose-*.yml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(composeYAML); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write compose file: %w", err)
	}
	tmpFile.Close()

	var cmdArgs []string
	if c.composeCmd == "compose" {
		cmdArgs = []string{c.composeCmd, "-f", tmpFile.Name()}
	} else {
		cmdArgs = []string{"-f", tmpFile.Name()}
	}
	cmdArgs = append(cmdArgs, args...)

	var cmd *exec.Cmd
	if c.composeCmd == "compose" {
		if c.useSudo {
			sudoArgs := append([]string{string(c.runtime)}, cmdArgs...)
			cmd = exec.Command("sudo", sudoArgs...)
		} else {
			cmd = exec.Command(string(c.runtime), cmdArgs...)
		}
	} else {
		if c.useSudo {
			sudoArgs := append([]string{c.composeCmd}, cmdArgs...)
			cmd = exec.Command("sudo", sudoArgs...)
		} else {
			cmd = exec.Command(c.composeCmd, cmdArgs...)
		}
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, stderr.String())
	}

	return nil
}

func (c *Client) StartService(name string, composeYAML []byte) error {
	return c.execComposeCommand(composeYAML, "up", "-d")
}

func (c *Client) StopService(name string, composeYAML []byte) error {
	return c.execComposeCommand(composeYAML, "down")
}

func (c *Client) RestartService(name string, composeYAML []byte) error {
	if err := c.execComposeCommand(composeYAML, "down"); err != nil {
		return err
	}
	return c.execComposeCommand(composeYAML, "up", "-d")
}

func (c *Client) RebuildService(name string, composeYAML []byte, noCache bool) error {
	c.execComposeCommand(composeYAML, "down")

	buildArgs := []string{"build"}
	if noCache {
		buildArgs = append(buildArgs, "--no-cache")
	}
	if err := c.execComposeCommand(composeYAML, buildArgs...); err != nil {
		return err
	}

	return c.execComposeCommand(composeYAML, "up", "-d", "--force-recreate")
}

func (c *Client) GetStatus(name string) (*models.ContainerState, error) {
	output, err := c.execCommand("inspect", "--format", "{{json .State}}", name)
	if err != nil {
		return &models.ContainerState{
			Status: "stopped",
		}, nil
	}

	var state struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		Paused     bool   `json:"Paused"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
		ExitCode   int    `json:"ExitCode"`
		Error      string `json:"Error"`
		Health     struct {
			Status string `json:"Status"`
		} `json:"Health"`
	}

	if err := json.Unmarshal([]byte(output), &state); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}

	cs := &models.ContainerState{
		Status:       state.Status,
		RestartCount: 0,
	}

	if state.Health.Status != "" {
		cs.HealthStr = state.Health.Status
	}

	if state.StartedAt != "" && state.StartedAt != "0001-01-01T00:00:00Z" {
		if t, err := time.Parse(time.RFC3339Nano, state.StartedAt); err == nil {
			cs.StartedAtStr = &t
		}
	}

	if state.FinishedAt != "" && state.FinishedAt != "0001-01-01T00:00:00Z" {
		if t, err := time.Parse(time.RFC3339Nano, state.FinishedAt); err == nil {
			cs.FinishedStr = &t
		}
	}

	if state.ExitCode != 0 {
		ec := state.ExitCode
		cs.ExitCodeInt = &ec
	}

	if state.Error != "" {
		cs.ErrorStr = state.Error
	}

	return cs, nil
}

func (c *Client) GetLogs(name string, tail string) (string, error) {
	return c.execCommand("logs", "--tail", tail, name)
}

func (c *Client) GetMetrics(name string) (*models.ContainerMetrics, error) {
	output, err := c.execCommand("stats", "--no-stream", "--format", "{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}},{{.NetIO}}", name)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	parts := strings.Split(output, ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected stats format")
	}

	metrics := &models.ContainerMetrics{
		RecordedAt: time.Now(),
	}

	cpuStr := strings.TrimSuffix(parts[0], "%")
	if cpu, err := strconv.ParseFloat(cpuStr, 64); err == nil {
		metrics.CPUPercentage = cpu
	}

	memParts := strings.Split(parts[1], "/")
	if len(memParts) == 2 {
		metrics.MemoryUsedMB = parseMemory(strings.TrimSpace(memParts[0]))
		metrics.MemoryLimitMB = parseMemory(strings.TrimSpace(memParts[1]))
	}

	memPercStr := strings.TrimSuffix(parts[2], "%")
	if memPerc, err := strconv.ParseFloat(memPercStr, 64); err == nil {
		metrics.MemoryPercentage = memPerc
	}

	netParts := strings.Split(parts[3], "/")
	if len(netParts) == 2 {
		metrics.NetworkRxBytes = parseBytes(strings.TrimSpace(netParts[0]))
		metrics.NetworkTxBytes = parseBytes(strings.TrimSpace(netParts[1]))
	}

	return metrics, nil
}

func parseMemory(s string) float64 {
	s = strings.ToUpper(s)
	multiplier := 1.0

	if strings.HasSuffix(s, "GIB") || strings.HasSuffix(s, "GB") {
		multiplier = 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "GIB"), "GB")
	} else if strings.HasSuffix(s, "MIB") || strings.HasSuffix(s, "MB") {
		multiplier = 1
		s = strings.TrimSuffix(strings.TrimSuffix(s, "MIB"), "MB")
	} else if strings.HasSuffix(s, "KIB") || strings.HasSuffix(s, "KB") {
		multiplier = 1.0 / 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "KIB"), "KB")
	}

	val, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return val * multiplier
}

func parseBytes(s string) int64 {
	s = strings.ToUpper(s)
	multiplier := int64(1)

	if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		s = strings.TrimSuffix(s, "B")
	}

	val, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return int64(val) * multiplier
}

func (c *Client) GetRunningCount() (running int, stopped int) {
	output, err := c.execCommand("ps", "-a", "--format", "{{.Status}}")
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "up") {
			running++
		} else {
			stopped++
		}
	}
	return
}

func (c *Client) GetCategoryRunningCount(category string) int {
	return 0
}

func (c *Client) ListContainers() ([]string, error) {
	output, err := c.execCommand("ps", "--format", "{{.Names}}")
	if err != nil {
		return nil, err
	}
	return c.parseNamesList(output)
}

func (c *Client) parseNamesList(output string) ([]string, error) {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}
	return names, nil
}

func (c *Client) ListContainersWithLabels(labels map[string]string) ([]string, error) {
	args := []string{"ps", "-a", "--format", "{{.Names}}"}
	for k, v := range labels {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}
	output, err := c.execCommand(args...)
	if err != nil {
		return nil, err
	}
	return c.parseNamesList(output)
}

func (c *Client) Exec(containerName string, command []string) (string, error) {
	args := append([]string{"exec", containerName}, command...)
	return c.execCommand(args...)
}

func (c *Client) GetContainerStatus(name string) (*ContainerStatus, error) {
	state, err := c.GetStatus(name)
	if err != nil {
		return nil, err
	}

	cs := &ContainerStatus{
		Status:       state.Status,
		Health:       state.HealthStr,
		Error:        state.ErrorStr,
		RestartCount: state.RestartCount,
	}

	if state.StartedAtStr != nil {
		cs.StartedAt = state.StartedAtStr
	}

	if state.FinishedStr != nil {
		cs.FinishedAt = state.FinishedStr
	}

	if state.ExitCodeInt != nil {
		cs.ExitCode = state.ExitCodeInt
	}

	return cs, nil
}

func (c *Client) GetContainerMetrics(name string) (*ContainerMetrics, error) {
	m, err := c.GetMetrics(name)
	if err != nil {
		return nil, err
	}

	return &ContainerMetrics{
		CPUPercentage:    m.CPUPercentage,
		MemoryUsedMB:     m.MemoryUsedMB,
		MemoryLimitMB:    m.MemoryLimitMB,
		MemoryPercentage: m.MemoryPercentage,
		NetworkRxBytes:   m.NetworkRxBytes,
		NetworkTxBytes:   m.NetworkTxBytes,
		RecordedAt:       m.RecordedAt,
	}, nil
}

func (c *Client) RunCompose(composePath string, args ...string) (string, error) {
	var cmdArgs []string
	if c.composeCmd == "compose" {
		cmdArgs = []string{c.composeCmd, "-f", composePath}
	} else {
		cmdArgs = []string{"-f", composePath}
	}
	cmdArgs = append(cmdArgs, args...)

	var cmd *exec.Cmd
	if c.composeCmd == "compose" {
		if c.useSudo {
			sudoArgs := append([]string{string(c.runtime)}, cmdArgs...)
			cmd = exec.Command("sudo", sudoArgs...)
		} else {
			cmd = exec.Command(string(c.runtime), cmdArgs...)
		}
	} else {
		if c.useSudo {
			sudoArgs := append([]string{c.composeCmd}, cmdArgs...)
			cmd = exec.Command("sudo", sudoArgs...)
		} else {
			cmd = exec.Command(c.composeCmd, cmdArgs...)
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("compose error: %s: %w", string(out), err)
	}
	return string(out), nil
}

func (c *Client) CreateNetwork(name string, labels map[string]string) error {
	return ErrNotImplemented
}

func (c *Client) RemoveNetwork(name string) error {
	return ErrNotImplemented
}

func (c *Client) ListNetworks() ([]string, error) {
	return nil, ErrNotImplemented
}
