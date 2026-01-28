package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var knownComposeKeys = map[string]bool{
	"image": true, "container_name": true, "restart": true,
	"command": true, "user": true, "ports": true, "volumes": true,
	"environment": true, "env_file": true, "depends_on": true,
	"labels": true, "healthcheck": true, "networks": true,
}

type ComposeFile struct {
	Services map[string]ComposeService `yaml:"services"`
	Volumes  map[string]interface{}    `yaml:"volumes"`
	Networks map[string]interface{}    `yaml:"networks"`
}

type ComposeService struct {
	Image         string                 `yaml:"image"`
	ContainerName string                 `yaml:"container_name"`
	Restart       string                 `yaml:"restart"`
	Command       interface{}            `yaml:"command"`
	User          string                 `yaml:"user"`
	Ports         []string               `yaml:"ports"`
	Volumes       []string               `yaml:"volumes"`
	Environment   interface{}            `yaml:"environment"`
	EnvFile       interface{}            `yaml:"env_file"`
	DependsOn     interface{}            `yaml:"depends_on"`
	Labels        interface{}            `yaml:"labels"`
	Healthcheck   *ComposeHealthcheck    `yaml:"healthcheck"`
	Networks      interface{}            `yaml:"networks"`
}

type ComposeHealthcheck struct {
	Test        interface{} `yaml:"test"`
	Interval    string      `yaml:"interval"`
	Timeout     string      `yaml:"timeout"`
	Retries     int         `yaml:"retries"`
	StartPeriod string      `yaml:"start_period"`
}

type ParsedService struct {
	Name          string
	Category      string
	ImageName     string
	ImageTag      string
	RestartPolicy string
	Command       string
	UserSpec      string
	Ports         []ParsedPort
	Volumes       []ParsedVolume
	EnvVars       []ParsedEnvVar
	Dependencies  []string
	Labels        []ParsedLabel
	Healthcheck   *ParsedHealthcheck
	Overrides     map[string]interface{}
}

type ParsedPort struct {
	HostIP        string
	HostPort      int
	ContainerPort int
	Protocol      string
}

type ParsedVolume struct {
	VolumeType string
	Source     string
	Target     string
	ReadOnly   bool
}

type ParsedEnvVar struct {
	Key      string
	Value    string
	IsSecret bool
}

type ParsedLabel struct {
	Key   string
	Value string
}

type ParsedHealthcheck struct {
	Test               string
	IntervalSeconds    int
	TimeoutSeconds     int
	Retries            int
	StartPeriodSeconds int
}

func ParseFile(path string) (*ParsedService, error) {
	services, err := ParseFileAll(path)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("no services found")
	}
	return services[0], nil
}

func ParseFileAll(path string) ([]*ParsedService, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if len(compose.Services) == 0 {
		return nil, fmt.Errorf("no services found in compose file")
	}

	var rawCompose struct {
		Services map[string]map[string]interface{} `yaml:"services"`
	}
	yaml.Unmarshal(data, &rawCompose)

	category := extractCategory(path)
	var services []*ParsedService

	for name, svc := range compose.Services {
		parsed := &ParsedService{
			Name:          name,
			Category:      category,
			RestartPolicy: svc.Restart,
			UserSpec:      svc.User,
		}

		if svc.Restart == "" {
			parsed.RestartPolicy = "unless-stopped"
		}

		parseImage(svc.Image, parsed)
		parsed.Command = parseCommand(svc.Command)
		parsed.Ports = parsePorts(svc.Ports)
		parsed.Volumes = parseVolumes(svc.Volumes, compose.Volumes)
		parsed.EnvVars = parseEnvironment(svc.Environment)
		parsed.Dependencies = parseDependsOn(svc.DependsOn)
		parsed.Labels = parseLabels(svc.Labels)
		parsed.Healthcheck = parseHealthcheck(svc.Healthcheck)
		parsed.Overrides = extractOverrides(rawCompose.Services[name])

		services = append(services, parsed)
	}

	return services, nil
}

func extractOverrides(raw map[string]interface{}) map[string]interface{} {
	if raw == nil {
		return nil
	}
	overrides := make(map[string]interface{})
	for key, val := range raw {
		if !knownComposeKeys[key] {
			overrides[key] = val
		}
	}
	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

func OverridesToJSON(overrides map[string]interface{}) []byte {
	if overrides == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(overrides)
	if err != nil {
		return []byte("{}")
	}
	return data
}

func extractCategory(path string) string {
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}

func parseImage(image string, svc *ParsedService) {
	if image == "" {
		return
	}

	parts := strings.Split(image, ":")
	svc.ImageName = parts[0]
	if len(parts) > 1 {
		svc.ImageTag = parts[1]
	} else {
		svc.ImageTag = "latest"
	}
}

func parseCommand(cmd interface{}) string {
	if cmd == nil {
		return ""
	}

	switch v := cmd.(type) {
	case string:
		return v
	case []interface{}:
		parts := make([]string, len(v))
		for i, p := range v {
			parts[i] = fmt.Sprintf("%v", p)
		}
		return strings.Join(parts, " ")
	}
	return ""
}

var portRegex = regexp.MustCompile(`^(?:([0-9.]+):)?(\d+):(\d+)(?:/(\w+))?$`)

func parsePorts(ports []string) []ParsedPort {
	var result []ParsedPort
	for _, p := range ports {
		matches := portRegex.FindStringSubmatch(p)
		if matches == nil {
			continue
		}

		hostIP := "127.0.0.1"
		if matches[1] != "" {
			hostIP = matches[1]
		}

		hostPort, _ := strconv.Atoi(matches[2])
		containerPort, _ := strconv.Atoi(matches[3])
		protocol := "tcp"
		if matches[4] != "" {
			protocol = matches[4]
		}

		result = append(result, ParsedPort{
			HostIP:        hostIP,
			HostPort:      hostPort,
			ContainerPort: containerPort,
			Protocol:      protocol,
		})
	}
	return result
}

func parseVolumes(volumes []string, namedVolumes map[string]interface{}) []ParsedVolume {
	var result []ParsedVolume
	for _, v := range volumes {
		parts := strings.Split(v, ":")
		if len(parts) < 2 {
			continue
		}

		pv := ParsedVolume{
			Source: parts[0],
			Target: parts[1],
		}

		if _, isNamed := namedVolumes[parts[0]]; isNamed {
			pv.VolumeType = "named"
		} else if strings.HasPrefix(parts[0], "/") || strings.HasPrefix(parts[0], ".") {
			pv.VolumeType = "bind"
		} else {
			pv.VolumeType = "named"
		}

		if len(parts) > 2 && parts[2] == "ro" {
			pv.ReadOnly = true
		}

		result = append(result, pv)
	}
	return result
}

var secretPatterns = []string{
	"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL", "AUTH",
}

func parseEnvironment(env interface{}) []ParsedEnvVar {
	var result []ParsedEnvVar

	switch v := env.(type) {
	case []interface{}:
		for _, e := range v {
			s := fmt.Sprintf("%v", e)
			parts := strings.SplitN(s, "=", 2)
			ev := ParsedEnvVar{Key: parts[0]}
			if len(parts) > 1 {
				ev.Value = parts[1]
			}
			ev.IsSecret = isSecretKey(ev.Key)
			result = append(result, ev)
		}
	case map[string]interface{}:
		for key, val := range v {
			ev := ParsedEnvVar{
				Key:      key,
				Value:    fmt.Sprintf("%v", val),
				IsSecret: isSecretKey(key),
			}
			result = append(result, ev)
		}
	}

	return result
}

func isSecretKey(key string) bool {
	upper := strings.ToUpper(key)
	for _, pattern := range secretPatterns {
		if strings.Contains(upper, pattern) {
			return true
		}
	}
	return false
}

func parseDependsOn(deps interface{}) []string {
	var result []string

	switch v := deps.(type) {
	case []interface{}:
		for _, d := range v {
			result = append(result, fmt.Sprintf("%v", d))
		}
	case map[string]interface{}:
		for name := range v {
			result = append(result, name)
		}
	}

	return result
}

func parseLabels(labels interface{}) []ParsedLabel {
	var result []ParsedLabel

	switch v := labels.(type) {
	case []interface{}:
		for _, l := range v {
			s := fmt.Sprintf("%v", l)
			parts := strings.SplitN(s, "=", 2)
			pl := ParsedLabel{Key: parts[0]}
			if len(parts) > 1 {
				pl.Value = parts[1]
			}
			result = append(result, pl)
		}
	case map[string]interface{}:
		for key, val := range v {
			result = append(result, ParsedLabel{
				Key:   key,
				Value: fmt.Sprintf("%v", val),
			})
		}
	}

	return result
}

func parseHealthcheck(hc *ComposeHealthcheck) *ParsedHealthcheck {
	if hc == nil {
		return nil
	}

	parsed := &ParsedHealthcheck{
		Retries:            hc.Retries,
		IntervalSeconds:    parseDuration(hc.Interval),
		TimeoutSeconds:     parseDuration(hc.Timeout),
		StartPeriodSeconds: parseDuration(hc.StartPeriod),
	}

	switch v := hc.Test.(type) {
	case string:
		parsed.Test = v
	case []interface{}:
		parts := make([]string, len(v))
		for i, p := range v {
			parts[i] = fmt.Sprintf("%v", p)
		}
		parsed.Test = strings.Join(parts, " ")
	}

	if parsed.Retries == 0 {
		parsed.Retries = 3
	}
	if parsed.IntervalSeconds == 0 {
		parsed.IntervalSeconds = 30
	}
	if parsed.TimeoutSeconds == 0 {
		parsed.TimeoutSeconds = 10
	}

	return parsed
}

var durationRegex = regexp.MustCompile(`^(\d+)(s|m|h)?$`)

func parseDuration(d string) int {
	if d == "" {
		return 0
	}

	matches := durationRegex.FindStringSubmatch(d)
	if matches == nil {
		return 0
	}

	val, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "m":
		return val * 60
	case "h":
		return val * 3600
	default:
		return val
	}
}
