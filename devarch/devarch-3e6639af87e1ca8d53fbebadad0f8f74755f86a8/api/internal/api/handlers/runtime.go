package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/priz/devarch-api/internal/container"
	"github.com/priz/devarch-api/internal/podman"
)

type RuntimeHandler struct {
	containerClient *container.Client
	podmanClient    *podman.Client
}

func NewRuntimeHandler(cc *container.Client, pc *podman.Client) *RuntimeHandler {
	return &RuntimeHandler{
		containerClient: cc,
		podmanClient:    pc,
	}
}

type runtimeInfo struct {
	Installed  bool    `json:"installed"`
	Version    *string `json:"version"`
	Running    bool    `json:"running"`
	Responsive bool    `json:"responsive"`
}

type runtimeStatusResponse struct {
	Current   string                `json:"current"`
	Available map[string]runtimeInfo `json:"available"`
	Containers map[string]int       `json:"containers"`
	Microservices struct {
		Running       int    `json:"running"`
		Network       string `json:"network"`
		NetworkExists bool   `json:"network_exists"`
	} `json:"microservices"`
}

func (h *RuntimeHandler) Status(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp := runtimeStatusResponse{
		Current:    "none",
		Available:  make(map[string]runtimeInfo),
		Containers: make(map[string]int),
	}
	resp.Microservices.Network = "microservices-net"

	for _, rt := range []string{"podman", "docker"} {
		info := runtimeInfo{}
		version, err := execRuntime(rt, "version", "--format", "{{.Client.Version}}")
		if err == nil {
			info.Installed = true
			v := strings.TrimSpace(version)
			info.Version = &v

			_, checkErr := execRuntime(rt, "info")
			if checkErr == nil {
				info.Running = true
				info.Responsive = true
			}
		}
		resp.Available[rt] = info

		if info.Responsive {
			count, _ := execRuntime(rt, "ps", "-q")
			lines := strings.Split(strings.TrimSpace(count), "\n")
			n := 0
			for _, l := range lines {
				if l != "" {
					n++
				}
			}
			resp.Containers[rt] = n
		}
	}

	if resp.Available["podman"].Responsive {
		resp.Current = "podman"
	} else if resp.Available["docker"].Responsive {
		resp.Current = "docker"
	}

	if resp.Current != "none" {
		_, err := execRuntime(resp.Current, "network", "inspect", "microservices-net")
		resp.Microservices.NetworkExists = err == nil

		if resp.Microservices.NetworkExists {
			out, err := execRuntime(resp.Current, "ps", "--filter", "network=microservices-net", "--format", "{{.Names}}")
			if err == nil {
				lines := strings.Split(strings.TrimSpace(out), "\n")
				for _, l := range lines {
					if l != "" {
						resp.Microservices.Running++
					}
				}
			}
		}
	}

	_ = ctx
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type switchRequest struct {
	Runtime string       `json:"runtime"`
	Options switchOptions `json:"options"`
}

type switchOptions struct {
	StopServices bool `json:"stop_services"`
	PreserveData bool `json:"preserve_data"`
	UpdateConfig bool `json:"update_config"`
}

func (h *RuntimeHandler) Switch(w http.ResponseWriter, r *http.Request) {
	var req switchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Runtime != "docker" && req.Runtime != "podman" {
		http.Error(w, `runtime must be "docker" or "podman"`, http.StatusBadRequest)
		return
	}

	current := "none"
	if _, err := execRuntime("podman", "info"); err == nil {
		current = "podman"
	} else if _, err := execRuntime("docker", "info"); err == nil {
		current = "docker"
	}

	if current == req.Runtime {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":           fmt.Sprintf("Already running on %s", req.Runtime),
			"current":           req.Runtime,
			"previous":          req.Runtime,
			"no_change_required": true,
		})
		return
	}

	if _, err := execRuntime(req.Runtime, "version"); err != nil {
		http.Error(w, fmt.Sprintf("%s is not installed or not accessible", req.Runtime), http.StatusServiceUnavailable)
		return
	}

	servicesStopped := 0
	if req.Options.StopServices && current != "none" {
		out, _ := execRuntime(current, "ps", "--filter", "network=microservices-net", "-q")
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, l := range lines {
			if l != "" {
				servicesStopped++
			}
		}

		if servicesStopped > 0 {
			args := []string{"stop-all"}
			if !req.Options.PreserveData {
				args = append(args, "--remove-volumes")
			}
			exec.Command("/workspace/scripts/service-manager.sh", args...).Run()
		}
	}

	configUpdated := false
	if req.Options.UpdateConfig {
		configPath := "/workspace/scripts/config.sh"
		if data, err := os.ReadFile(configPath); err == nil {
			content := string(data)
			content = strings.ReplaceAll(content,
				fmt.Sprintf(`export CONTAINER_RUNTIME="%s"`, current),
				fmt.Sprintf(`export CONTAINER_RUNTIME="%s"`, req.Runtime),
			)
			if os.WriteFile(configPath, []byte(content), 0644) == nil {
				configUpdated = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"previous":         current,
		"current":          req.Runtime,
		"services_stopped": servicesStopped,
		"config_updated":   configUpdated,
		"message":          fmt.Sprintf("Successfully switched to %s", req.Runtime),
	})
}

type socketStatusResponse struct {
	Active  string               `json:"active"`
	Sockets map[string]socketInfo `json:"sockets"`
	Environment struct {
		DockerHost string `json:"docker_host"`
		User       string `json:"user"`
		UID        int    `json:"uid"`
	} `json:"environment"`
	Integration struct {
		ProjectNetwork string `json:"project_network"`
		NetworkExists  bool   `json:"network_exists"`
		RunningServices int   `json:"running_services"`
	} `json:"integration"`
}

type socketInfo struct {
	Active       bool   `json:"active"`
	SocketPath   string `json:"socket_path"`
	Exists       bool   `json:"exists"`
	Connectivity string `json:"connectivity"`
}

func (h *RuntimeHandler) SocketStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid := os.Getuid()
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	resp := socketStatusResponse{
		Active:  "none",
		Sockets: make(map[string]socketInfo),
	}
	resp.Environment.DockerHost = os.Getenv("DOCKER_HOST")
	if resp.Environment.DockerHost == "" {
		resp.Environment.DockerHost = os.Getenv("CONTAINER_HOST")
	}
	resp.Environment.User = username
	resp.Environment.UID = uid
	resp.Integration.ProjectNetwork = "microservices-net"

	rootlessPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", uid)
	rootfulPath := "/run/podman/podman.sock"

	info, err := h.podmanClient.Info(ctx)
	if err == nil {
		if info.Host.Security.RootlessMode {
			resp.Active = "rootless"
		} else {
			resp.Active = "rootful"
		}
	}

	rootlessActive := resp.Active == "rootless"
	rootfulActive := resp.Active == "rootful"

	resp.Sockets["rootless"] = socketInfo{
		Active:       rootlessActive,
		SocketPath:   rootlessPath,
		Exists:       rootlessActive,
		Connectivity: boolToConnectivity(rootlessActive),
	}
	resp.Sockets["rootful"] = socketInfo{
		Active:       rootfulActive,
		SocketPath:   rootfulPath,
		Exists:       rootfulActive,
		Connectivity: boolToConnectivity(rootfulActive),
	}

	if resp.Active != "none" {
		_, netErr := execRuntime("podman", "network", "inspect", "microservices-net")
		resp.Integration.NetworkExists = netErr == nil

		if resp.Integration.NetworkExists {
			out, _ := execRuntime("podman", "ps", "--filter", "network=microservices-net", "--format", "{{.Names}}")
			for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
				if l != "" {
					resp.Integration.RunningServices++
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type socketStartRequest struct {
	Type    string             `json:"type"`
	Options socketStartOptions `json:"options"`
}

type socketStartOptions struct {
	EnableLingering bool `json:"enable_lingering"`
	StopConflicting bool `json:"stop_conflicting"`
}

func (h *RuntimeHandler) SocketStart(w http.ResponseWriter, r *http.Request) {
	var req socketStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "rootless"
	}
	if req.Type != "rootless" && req.Type != "rootful" {
		http.Error(w, `type must be "rootless" or "rootful"`, http.StatusBadRequest)
		return
	}

	isUser := req.Type == "rootless"

	if req.Options.StopConflicting {
		conflicting := "rootful"
		if req.Type == "rootful" {
			conflicting = "rootless"
		}
		conflictIsUser := conflicting == "rootless"
		stopSystemdService("podman.socket", conflictIsUser)
	}

	if isUser && req.Options.EnableLingering {
		if u, err := user.Current(); err == nil {
			exec.Command("loginctl", "enable-linger", u.Username).Run()
		}
	}

	if !startSystemdService("podman.socket", isUser) {
		http.Error(w, fmt.Sprintf("could not start %s podman socket", req.Type), http.StatusInternalServerError)
		return
	}

	uid := os.Getuid()
	socketPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", uid)
	if req.Type == "rootful" {
		socketPath = "/run/podman/podman.sock"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":        req.Type,
		"socket_path": socketPath,
		"connectivity": "started",
		"message":     fmt.Sprintf("%s socket started", strings.ToUpper(req.Type[:1])+req.Type[1:]),
	})
}

func execRuntime(runtime string, args ...string) (string, error) {
	cmd := exec.Command(runtime, args...)
	out, err := cmd.Output()
	return string(out), err
}

func boolToConnectivity(b bool) string {
	if b {
		return "working"
	}
	return "unavailable"
}

func stopSystemdService(service string, userMode bool) {
	args := []string{"stop", service}
	if userMode {
		args = append([]string{"--user"}, args...)
	}
	exec.Command("systemctl", args...).Run()
}

func startSystemdService(service string, userMode bool) bool {
	args := []string{"start", service}
	if userMode {
		args = append([]string{"--user"}, args...)
	}
	return exec.Command("systemctl", args...).Run() == nil
}
