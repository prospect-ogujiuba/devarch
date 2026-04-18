package runtime

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

type containerInspectDocument struct {
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	RestartCount int    `json:"RestartCount"`
	Config       struct {
		Image       string            `json:"Image"`
		Env         []string          `json:"Env"`
		Cmd         []string          `json:"Cmd"`
		Entrypoint  []string          `json:"Entrypoint"`
		WorkingDir  string            `json:"WorkingDir"`
		Labels      map[string]string `json:"Labels"`
		Healthcheck *struct {
			Test        []string `json:"Test"`
			Interval    int64    `json:"Interval"`
			Timeout     int64    `json:"Timeout"`
			StartPeriod int64    `json:"StartPeriod"`
			Retries     int      `json:"Retries"`
		} `json:"Healthcheck"`
	} `json:"Config"`
	State struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		ExitCode   int    `json:"ExitCode"`
		Error      string `json:"Error"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
		Health     *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
	NetworkSettings struct {
		Ports    map[string][]portBinding           `json:"Ports"`
		Networks map[string]networkEndpointSettings `json:"Networks"`
	} `json:"NetworkSettings"`
	Mounts []mountDocument `json:"Mounts"`
}

type portBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type networkEndpointSettings struct {
	Aliases []string `json:"Aliases"`
}

type mountDocument struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	RW          bool   `json:"RW"`
}

type networkInspectDocument struct {
	Name   string            `json:"Name"`
	ID     string            `json:"Id"`
	Driver string            `json:"Driver"`
	Labels map[string]string `json:"Labels"`
}

func NormalizeInspectSnapshot(provider string, desired *DesiredWorkspace, containerInspectJSON, networkInspectJSON []byte) (*Snapshot, error) {
	if desired == nil {
		return nil, fmt.Errorf("normalize inspect snapshot: nil desired workspace")
	}

	snapshot := &Snapshot{
		Workspace: SnapshotWorkspace{
			Name:     desired.Name,
			Provider: provider,
		},
		Resources: nil,
	}

	if len(strings.TrimSpace(string(networkInspectJSON))) > 0 {
		var networkDocs []networkInspectDocument
		if err := json.Unmarshal(networkInspectJSON, &networkDocs); err != nil {
			return nil, fmt.Errorf("decode network inspect: %w", err)
		}
		if len(networkDocs) > 0 {
			doc := networkDocs[0]
			snapshot.Workspace.Network = &SnapshotNetwork{
				Name:   doc.Name,
				ID:     doc.ID,
				Driver: doc.Driver,
				Labels: cloneStringMap(doc.Labels),
			}
		}
	}

	if len(strings.TrimSpace(string(containerInspectJSON))) == 0 {
		return snapshot, nil
	}

	var docs []containerInspectDocument
	if err := json.Unmarshal(containerInspectJSON, &docs); err != nil {
		return nil, fmt.Errorf("decode container inspect: %w", err)
	}

	resources := make([]*SnapshotResource, 0, len(docs))
	for _, doc := range docs {
		labels := cloneStringMap(doc.Config.Labels)
		if labels[LabelWorkspace] != desired.Name {
			continue
		}
		resourceKey := labels[LabelResource]
		if resourceKey == "" {
			resourceKey = inferResourceKey(desired, trimContainerName(doc.Name))
		}
		logicalHost := labels[LabelHostAlias]
		if logicalHost == "" {
			logicalHost = resourceKey
		}
		resources = append(resources, &SnapshotResource{
			Key:         resourceKey,
			RuntimeName: trimContainerName(doc.Name),
			LogicalHost: logicalHost,
			ID:          doc.ID,
			State: ResourceState{
				Status:       doc.State.Status,
				Running:      doc.State.Running,
				Health:       healthStatus(doc.State.Health),
				ExitCode:     doc.State.ExitCode,
				RestartCount: doc.RestartCount,
				StartedAt:    parseTimePtr(doc.State.StartedAt),
				FinishedAt:   parseTimePtr(doc.State.FinishedAt),
				Error:        doc.State.Error,
			},
			Spec: ResourceSpec{
				Image:      doc.Config.Image,
				Command:    cloneStringSlice(doc.Config.Cmd),
				Entrypoint: cloneStringSlice(doc.Config.Entrypoint),
				WorkingDir: doc.Config.WorkingDir,
				Env:        envFromInspect(doc.Config.Env),
				Ports:      portsFromInspect(doc.NetworkSettings.Ports),
				Volumes:    volumesFromInspect(doc.Mounts),
				Health:     healthFromInspect(doc.Config.Healthcheck),
				Labels:     labels,
			},
		})
	}

	sort.Slice(resources, func(i, j int) bool {
		if resources[i].Key != resources[j].Key {
			return resources[i].Key < resources[j].Key
		}
		return resources[i].RuntimeName < resources[j].RuntimeName
	})
	snapshot.Resources = resources
	return snapshot, nil
}

func envFromInspect(values []string) map[string]workspace.EnvValue {
	if len(values) == 0 {
		return nil
	}
	env := make(map[string]workspace.EnvValue, len(values))
	for _, value := range values {
		key, rest, ok := strings.Cut(value, "=")
		if !ok {
			env[value] = workspace.StringEnvValue("")
			continue
		}
		env[key] = workspace.StringEnvValue(rest)
	}
	if len(env) == 0 {
		return nil
	}
	return env
}

func portsFromInspect(values map[string][]portBinding) []PortSpec {
	if len(values) == 0 {
		return nil
	}
	ports := make([]PortSpec, 0)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		containerPort, protocol := parsePortKey(key)
		bindings := values[key]
		if len(bindings) == 0 {
			ports = append(ports, PortSpec{Container: containerPort, Protocol: protocol})
			continue
		}
		for _, binding := range bindings {
			published, _ := strconv.Atoi(binding.HostPort)
			ports = append(ports, PortSpec{Container: containerPort, Published: published, Protocol: protocol, HostIP: binding.HostIP})
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Container != ports[j].Container {
			return ports[i].Container < ports[j].Container
		}
		if ports[i].Protocol != ports[j].Protocol {
			return ports[i].Protocol < ports[j].Protocol
		}
		if ports[i].Published != ports[j].Published {
			return ports[i].Published < ports[j].Published
		}
		return ports[i].HostIP < ports[j].HostIP
	})
	return ports
}

func volumesFromInspect(values []mountDocument) []VolumeSpec {
	if len(values) == 0 {
		return nil
	}
	volumes := make([]VolumeSpec, len(values))
	for i := range values {
		volumes[i] = VolumeSpec{
			Source:   values[i].Source,
			Target:   values[i].Destination,
			ReadOnly: !values[i].RW,
			Type:     values[i].Type,
		}
	}
	sort.Slice(volumes, func(i, j int) bool {
		if volumes[i].Target != volumes[j].Target {
			return volumes[i].Target < volumes[j].Target
		}
		if volumes[i].Source != volumes[j].Source {
			return volumes[i].Source < volumes[j].Source
		}
		return volumes[i].Type < volumes[j].Type
	})
	return volumes
}

func healthFromInspect(value *struct {
	Test        []string `json:"Test"`
	Interval    int64    `json:"Interval"`
	Timeout     int64    `json:"Timeout"`
	StartPeriod int64    `json:"StartPeriod"`
	Retries     int      `json:"Retries"`
}) *workspace.Health {
	if value == nil {
		return nil
	}
	health := &workspace.Health{
		Test:    append(workspace.StringList(nil), value.Test...),
		Retries: value.Retries,
	}
	if value.Interval > 0 {
		health.Interval = time.Duration(value.Interval).String()
	}
	if value.Timeout > 0 {
		health.Timeout = time.Duration(value.Timeout).String()
	}
	if value.StartPeriod > 0 {
		health.StartPeriod = time.Duration(value.StartPeriod).String()
	}
	return health
}

func trimContainerName(name string) string {
	return strings.TrimPrefix(name, "/")
}

func inferResourceKey(desired *DesiredWorkspace, runtimeName string) string {
	if desired == nil {
		return runtimeName
	}
	for _, resource := range desired.Resources {
		if resource != nil && resource.RuntimeName == runtimeName {
			return resource.Key
		}
	}
	prefix := ResourceRuntimeName(desired.Name, "", desired.NamingStrategy)
	prefix = strings.TrimSuffix(prefix, "-")
	if strings.HasPrefix(runtimeName, prefix+"-") {
		return strings.TrimPrefix(runtimeName, prefix+"-")
	}
	return runtimeName
}

func parsePortKey(value string) (int, string) {
	container, protocol, ok := strings.Cut(value, "/")
	if !ok {
		protocol = "tcp"
	}
	port, _ := strconv.Atoi(container)
	if protocol == "" {
		protocol = "tcp"
	}
	return port, protocol
}

func parseTimePtr(value string) *time.Time {
	if value == "" || value == "0001-01-01T00:00:00Z" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func healthStatus(value *struct {
	Status string `json:"Status"`
}) string {
	if value == nil {
		return ""
	}
	return value.Status
}

func ParseLogOutput(stream string, output []byte) []LogChunk {
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	chunks := make([]LogChunk, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		chunk := LogChunk{Stream: stream, Line: line}
		if first, rest, ok := strings.Cut(line, " "); ok {
			if timestamp := parseTimePtr(first); timestamp != nil {
				chunk.Timestamp = timestamp
				chunk.Line = rest
			}
		}
		chunks = append(chunks, chunk)
	}
	if len(chunks) == 0 {
		return nil
	}
	return chunks
}
