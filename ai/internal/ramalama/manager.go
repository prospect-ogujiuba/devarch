package ramalama

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"
)

const containerName = "devarch-llm"
const volumeName = "devarch-llm-models"

type Status struct {
	Running bool   `json:"running"`
	Model   string `json:"model"`
	Port    int    `json:"port"`
	GPU     string `json:"gpu"`
	Uptime  string `json:"uptime,omitempty"`
}

type Manager struct {
	cfg     Config
	mu      sync.Mutex
	runtime string
	lastUse time.Time
	stopCh  chan struct{}
}

func NewManager() *Manager {
	runtime := "podman"
	if _, err := exec.LookPath("podman"); err != nil {
		runtime = "docker"
	}

	return &Manager{
		cfg:    LoadConfig(),
		runtime: runtime,
		stopCh: make(chan struct{}),
	}
}

func (m *Manager) StartIdleWatcher(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.mu.Lock()
				if !m.lastUse.IsZero() && time.Since(m.lastUse) > m.cfg.IdleTimeout {
					m.mu.Unlock()
					if running, _ := m.isRunning(); running {
						slog.Info("stopping idle LLM container", "idle_for", time.Since(m.lastUse))
						m.Stop()
					}
				} else {
					m.mu.Unlock()
				}
			}
		}
	}()
}

func (m *Manager) TouchLastUse() {
	m.mu.Lock()
	m.lastUse = time.Now()
	m.mu.Unlock()
}

func (m *Manager) EnsureRunning() error {
	running, err := m.isRunning()
	if err != nil {
		return fmt.Errorf("check container status: %w", err)
	}
	if running {
		m.TouchLastUse()
		return nil
	}
	return m.start()
}

func (m *Manager) GetStatus() (*Status, error) {
	running, _ := m.isRunning()
	s := &Status{
		Running: running,
		Model:   m.cfg.Model,
		Port:    m.cfg.Port,
		GPU:     m.detectGPULabel(),
	}

	if running {
		if uptime, err := m.getUptime(); err == nil {
			s.Uptime = uptime
		}
	}

	return s, nil
}

func (m *Manager) Stop() error {
	output, err := m.exec("stop", containerName)
	if err != nil {
		combined := strings.ToLower(output + " " + err.Error())
		if strings.Contains(combined, "no such container") || strings.Contains(combined, "no container with") {
			return nil
		}
		return fmt.Errorf("%s: %w", strings.TrimSpace(output), err)
	}
	if _, err := m.exec("rm", containerName); err != nil {
		slog.Warn("failed to remove LLM container", "error", err)
	}
	return nil
}

func (m *Manager) PullModel(model string) (string, error) {
	if err := m.EnsureRunning(); err != nil {
		return "", fmt.Errorf("start LLM container: %w", err)
	}
	output, err := m.exec("exec", containerName, "ramalama", "pull", model)
	if err != nil {
		return "", fmt.Errorf("pull model: %s: %w", strings.TrimSpace(output), err)
	}
	return output, nil
}

func (m *Manager) ListModels() ([]string, error) {
	if err := m.EnsureRunning(); err != nil {
		return nil, fmt.Errorf("start LLM container: %w", err)
	}
	output, err := m.exec("exec", containerName, "ramalama", "list")
	if err != nil {
		return nil, fmt.Errorf("list models: %s: %w", strings.TrimSpace(output), err)
	}
	var models []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "NAME") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				models = append(models, parts[0])
			}
		}
	}
	return models, nil
}

func (m *Manager) BaseURL() string {
	if m.cfg.Network != "" {
		return fmt.Sprintf("http://%s:8080", containerName)
	}
	return fmt.Sprintf("http://localhost:%d", m.cfg.Port)
}

func (m *Manager) Config() Config {
	return m.cfg
}

func (m *Manager) start() error {
	m.exec("rm", "-f", containerName)
	m.exec("volume", "create", volumeName)

	args := []string{
		"run", "-d",
		"--name", containerName,
		"--replace",
		"-p", fmt.Sprintf("%d:8080", m.cfg.Port),
		"-v", volumeName + ":/root/.local/share/ramalama",
		"--label", "devarch.managed_by=devarch",
		"--label", "devarch.component=llm",
	}

	if m.cfg.Network != "" {
		args = append(args, "--network", m.cfg.Network)
	}

	gpuArgs := DetectGPU(m.cfg.GPUMode)
	args = append(args, gpuArgs...)
	args = append(args, m.cfg.Image, "ramalama", "serve", "--port", "8080", m.cfg.Model)

	output, err := m.exec(args...)
	if err != nil {
		return fmt.Errorf("start LLM container: %s: %w", strings.TrimSpace(output), err)
	}

	if err := m.waitReady(90 * time.Second); err != nil {
		return fmt.Errorf("LLM container not ready: %w", err)
	}

	m.TouchLastUse()
	slog.Info("LLM container started", "model", m.cfg.Model, "port", m.cfg.Port)
	return nil
}

func (m *Manager) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		output, err := m.exec("exec", containerName, "curl", "-sf", "http://localhost:8080/v1/models")
		if err == nil && strings.Contains(output, "model") {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for LLM to start after %s", timeout)
}

func (m *Manager) isRunning() (bool, error) {
	output, err := m.exec("inspect", "--format", "{{.State.Running}}", containerName)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) == "true", nil
}

func (m *Manager) getUptime() (string, error) {
	output, err := m.exec("inspect", "--format", "{{json .State.StartedAt}}", containerName)
	if err != nil {
		return "", err
	}
	var startedAt string
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &startedAt); err != nil {
		return "", err
	}
	t, err := time.Parse(time.RFC3339Nano, startedAt)
	if err != nil {
		return "", err
	}
	d := time.Since(t)
	if d < time.Minute {
		return strconv.Itoa(int(d.Seconds())) + "s", nil
	}
	if d < time.Hour {
		return strconv.Itoa(int(d.Minutes())) + "m", nil
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60), nil
}

func (m *Manager) detectGPULabel() string {
	gpuArgs := DetectGPU(m.cfg.GPUMode)
	if len(gpuArgs) == 0 {
		return "cpu"
	}
	for _, a := range gpuArgs {
		if strings.Contains(a, "nvidia") {
			return "nvidia"
		}
		if strings.Contains(a, "kfd") {
			return "amd"
		}
	}
	return "cpu"
}

func (m *Manager) exec(args ...string) (string, error) {
	cmd := exec.Command(m.runtime, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
