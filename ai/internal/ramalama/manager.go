package ramalama

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Status struct {
	Running bool   `json:"running"`
	Model   string `json:"model"`
	Port    int    `json:"port"`
	Uptime  string `json:"uptime,omitempty"`
}

type Manager struct {
	cfg           Config
	containerName string
	mu            sync.Mutex
	runtime       string
	lastUse       time.Time
	stopCh        chan struct{}
}

func detectRuntime() string {
	if _, err := exec.LookPath("podman"); err != nil {
		return "docker"
	}
	return "podman"
}

func NewManager() *Manager {
	return &Manager{
		cfg:           LoadConfig(),
		containerName: "devarch-llm",
		runtime:       detectRuntime(),
		stopCh:        make(chan struct{}),
	}
}

func NewEmbedManager() *Manager {
	return &Manager{
		cfg:           LoadEmbedConfig(),
		containerName: "devarch-llm-embed",
		runtime:       detectRuntime(),
		stopCh:        make(chan struct{}),
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
						slog.Info("stopping idle container", "container", m.containerName, "idle_for", time.Since(m.lastUse))
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
	}

	if running {
		if uptime, err := m.getUptime(); err == nil {
			s.Uptime = uptime
		}
	}

	return s, nil
}

func (m *Manager) Stop() error {
	output, err := m.ramalama("stop", "--ignore", m.containerName)
	if err != nil {
		return fmt.Errorf("stop %s: %s: %w", m.containerName, strings.TrimSpace(output), err)
	}
	return nil
}

func (m *Manager) PullModel(model string) (string, error) {
	output, err := m.ramalama("pull", model)
	if err != nil {
		return "", fmt.Errorf("pull model: %s: %w", strings.TrimSpace(output), err)
	}
	return output, nil
}

func (m *Manager) ListModels() ([]string, error) {
	output, err := m.ramalama("list", "--noheading")
	if err != nil {
		return nil, fmt.Errorf("list models: %s: %w", strings.TrimSpace(output), err)
	}
	var models []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
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
		return fmt.Sprintf("http://%s:%d", m.containerName, m.cfg.Port)
	}
	return fmt.Sprintf("http://localhost:%d", m.cfg.Port)
}

func (m *Manager) Config() Config {
	return m.cfg
}

func (m *Manager) start() error {
	m.ramalama("stop", "--ignore", m.containerName)

	args := []string{
		"serve", "-d",
		"-n", m.containerName,
		"-p", strconv.Itoa(m.cfg.Port),
	}

	if m.cfg.Network != "" {
		args = append(args, "--network", m.cfg.Network)
	}

	for _, extra := range m.cfg.RuntimeArgs {
		args = append(args, "--runtime-args="+extra)
	}

	args = append(args, m.cfg.Model)

	output, err := m.ramalama(args...)
	if err != nil {
		return fmt.Errorf("start container %s: %s: %w", m.containerName, strings.TrimSpace(output), err)
	}

	if err := m.waitReady(90 * time.Second); err != nil {
		return fmt.Errorf("container %s not ready: %w", m.containerName, err)
	}

	m.TouchLastUse()
	slog.Info("container started", "container", m.containerName, "model", m.cfg.Model, "port", m.cfg.Port)
	return nil
}

func (m *Manager) waitReady(timeout time.Duration) error {
	url := m.BaseURL() + "/v1/models"
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("curl", "-sf", url)
		if output, err := cmd.CombinedOutput(); err == nil && strings.Contains(string(output), "model") {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout waiting for %s after %s", m.containerName, timeout)
}

func (m *Manager) isRunning() (bool, error) {
	output, err := exec.Command(m.runtime, "inspect", "--format", "{{.State.Running}}", m.containerName).CombinedOutput()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

func (m *Manager) getUptime() (string, error) {
	output, err := exec.Command(m.runtime, "inspect", "--format", "{{json .State.StartedAt}}", m.containerName).CombinedOutput()
	if err != nil {
		return "", err
	}
	var startedAt string
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &startedAt); err != nil {
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

func (m *Manager) ramalama(args ...string) (string, error) {
	if m.cfg.Store != "" {
		args = append([]string{"--store", m.cfg.Store}, args...)
	}
	cmd := exec.Command("ramalama", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
