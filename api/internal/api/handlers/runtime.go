package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/priz/devarch-api/internal/api/respond"
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

// Status godoc
// @Summary      Get runtime status
// @Description  Returns current container runtime status (Docker/Podman) including availability, running containers, and network status
// @Tags         runtime
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=runtimeStatusResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /runtime/status [get]
// @Security     ApiKeyAuth
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
	respond.JSON(w, r, http.StatusOK,resp)
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

// Switch godoc
// @Summary      Switch container runtime
// @Description  Switches between Docker and Podman runtimes with options to stop services and preserve data
// @Tags         runtime
// @Accept       json
// @Produce      json
// @Param        request body switchRequest true "Switch request"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      503 {object} respond.ErrorEnvelope
// @Router       /runtime/switch [post]
// @Security     ApiKeyAuth
func (h *RuntimeHandler) Switch(w http.ResponseWriter, r *http.Request) {
	var req switchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	if req.Runtime != "docker" && req.Runtime != "podman" {
		respond.BadRequest(w, r, `runtime must be "docker" or "podman"`)
		return
	}

	current := "none"
	if _, err := execRuntime("podman", "info"); err == nil {
		current = "podman"
	} else if _, err := execRuntime("docker", "info"); err == nil {
		current = "docker"
	}

	if current == req.Runtime {
		respond.Action(w, r, http.StatusOK, "no_change",
			respond.WithMessage(fmt.Sprintf("Already running on %s", req.Runtime)),
			respond.WithMetadata("current", req.Runtime),
		)
		return
	}

	if _, err := execRuntime(req.Runtime, "version"); err != nil {
		respond.Error(w, r, http.StatusServiceUnavailable, "service_unavailable", fmt.Sprintf("%s is not installed or not accessible", req.Runtime), nil)
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
			containerIDs := strings.Fields(out)
			args := append([]string{"stop"}, containerIDs...)
			execRuntime(current, args...)
			if !req.Options.PreserveData {
				rmArgs := append([]string{"rm", "-v"}, containerIDs...)
				execRuntime(current, rmArgs...)
			}
		}
	}

	configUpdated := false
	if req.Options.UpdateConfig {
		// Legacy scripts/config.sh runtime mutation was retired with shell shims.
		// Runtime changes are now reported through API metadata; persisted runtime
		// selection belongs in the v2 Go configuration surface when needed.
	}

	respond.Action(w, r, http.StatusOK, "switched",
		respond.WithMessage(fmt.Sprintf("Successfully switched to %s", req.Runtime)),
		respond.WithMetadata("previous", current),
		respond.WithMetadata("current", req.Runtime),
		respond.WithMetadata("services_stopped", servicesStopped),
		respond.WithMetadata("config_updated", configUpdated),
	)
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

// SocketStatus godoc
// @Summary      Get Podman socket status
// @Description  Returns detailed status of Podman rootless and rootful sockets including connectivity and environment info
// @Tags         runtime
// @Produce      json
// @Success      200 {object} respond.SuccessEnvelope{data=socketStatusResponse}
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /socket/status [get]
// @Security     ApiKeyAuth
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

	respond.JSON(w, r, http.StatusOK,resp)
}

type socketStartRequest struct {
	Type    string             `json:"type"`
	Options socketStartOptions `json:"options"`
}

type socketStartOptions struct {
	EnableLingering bool `json:"enable_lingering"`
	StopConflicting bool `json:"stop_conflicting"`
}

// SocketStart godoc
// @Summary      Start Podman socket
// @Description  Starts a Podman socket (rootless or rootful) via systemd with options to enable lingering and stop conflicting sockets
// @Tags         runtime
// @Accept       json
// @Produce      json
// @Param        request body socketStartRequest true "Socket start request"
// @Success      200 {object} respond.SuccessEnvelope{data=respond.ActionResponse}
// @Failure      400 {object} respond.ErrorEnvelope
// @Failure      500 {object} respond.ErrorEnvelope
// @Router       /socket/start [post]
// @Security     ApiKeyAuth
func (h *RuntimeHandler) SocketStart(w http.ResponseWriter, r *http.Request) {
	var req socketStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.BadRequest(w, r, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "rootless"
	}
	if req.Type != "rootless" && req.Type != "rootful" {
		respond.BadRequest(w, r, `type must be "rootless" or "rootful"`)
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
		respond.InternalError(w, r, fmt.Errorf("could not start %s podman socket", req.Type))
		return
	}

	uid := os.Getuid()
	socketPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", uid)
	if req.Type == "rootful" {
		socketPath = "/run/podman/podman.sock"
	}

	respond.Action(w, r, http.StatusOK, "started",
		respond.WithMessage(fmt.Sprintf("%s socket started", strings.ToUpper(req.Type[:1])+req.Type[1:])),
		respond.WithMetadata("type", req.Type),
		respond.WithMetadata("socket_path", socketPath),
	)
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
