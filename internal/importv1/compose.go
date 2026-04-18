package importv1

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/prospect-ogujiuba/devarch/internal/workspace"
	"gopkg.in/yaml.v3"
)

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
	Volumes  map[string]any           `yaml:"volumes"`
}

type composeService struct {
	Image         string              `yaml:"image,omitempty"`
	Build         any                 `yaml:"build,omitempty"`
	ContainerName string              `yaml:"container_name,omitempty"`
	Restart       string              `yaml:"restart,omitempty"`
	Command       any                 `yaml:"command,omitempty"`
	User          string              `yaml:"user,omitempty"`
	WorkingDir    string              `yaml:"working_dir,omitempty"`
	Ports         []string            `yaml:"ports,omitempty"`
	Volumes       []string            `yaml:"volumes,omitempty"`
	Environment   any                 `yaml:"environment,omitempty"`
	EnvFile       any                 `yaml:"env_file,omitempty"`
	DependsOn     any                 `yaml:"depends_on,omitempty"`
	Labels        any                 `yaml:"labels,omitempty"`
	Healthcheck   *composeHealthcheck `yaml:"healthcheck,omitempty"`
	Networks      any                 `yaml:"networks,omitempty"`
}

type composeHealthcheck struct {
	Test        any    `yaml:"test,omitempty"`
	Interval    string `yaml:"interval,omitempty"`
	Timeout     string `yaml:"timeout,omitempty"`
	Retries     int    `yaml:"retries,omitempty"`
	StartPeriod string `yaml:"start_period,omitempty"`
}

type parsedComposeService struct {
	Name          string
	Category      string
	SourcePath    string
	ServiceDir    string
	Runtime       map[string]any
	Ports         []workspace.Port
	Volumes       []parsedVolume
	Env           map[string]any
	EnvFiles      []string
	Dependencies  []string
	Labels        map[string]string
	Health        map[string]any
	Networks      []string
	ConfigMounts  []parsedConfigMount
	ConfigFiles   map[string]importedConfigFile
	ContainerName string
	RestartPolicy string
	User          string
}

type parsedVolume struct {
	Source   string
	Target   string
	ReadOnly bool
	Type     string
	External bool
}

type parsedConfigMount struct {
	SourcePath string
	TargetPath string
	ReadOnly   bool
}

type importedConfigFile struct {
	Content  string
	FileMode string
}

func discoverComposeFiles(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) != "compose.yml" {
			return nil
		}
		paths = append(paths, filepath.Clean(path))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover compose files under %s: %w", root, err)
	}
	sort.Strings(paths)
	return paths, nil
}

func parseComposeServices(path string) ([]parsedComposeService, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read compose file %s: %w", path, err)
	}

	var document composeFile
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode compose file %s: %w", path, err)
	}
	if len(document.Services) == 0 {
		return nil, fmt.Errorf("compose file %s: no services found", path)
	}

	var raw struct {
		Services map[string]map[string]any `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode raw compose file %s: %w", path, err)
	}

	services := make([]parsedComposeService, 0, len(document.Services))
	serviceDir := filepath.Dir(path)
	configFiles, err := loadConfigFiles(filepath.Join(serviceDir, "config"))
	if err != nil {
		return nil, err
	}
	category := categoryFromComposePath(path)

	for name, service := range document.Services {
		runtime, err := parseRuntime(serviceDir, service)
		if err != nil {
			return nil, fmt.Errorf("parse runtime for %s in %s: %w", name, path, err)
		}
		volumes := parseComposeVolumes(service.Volumes, document.Volumes)
		configMounts, remainingVolumes := extractConfigMounts(volumes, raw.Services[name])
		services = append(services, parsedComposeService{
			Name:          name,
			Category:      category,
			SourcePath:    filepath.Clean(path),
			ServiceDir:    serviceDir,
			Runtime:       runtime,
			Ports:         parseComposePorts(service.Ports),
			Volumes:       remainingVolumes,
			Env:           parseComposeEnvironment(service.Environment),
			EnvFiles:      parseComposeStringList(service.EnvFile),
			Dependencies:  parseComposeDependsOn(service.DependsOn),
			Labels:        parseComposeStringMap(service.Labels),
			Health:        parseComposeHealth(service.Healthcheck),
			Networks:      parseComposeNetworks(service.Networks),
			ConfigMounts:  configMounts,
			ConfigFiles:   cloneConfigFiles(configFiles),
			ContainerName: service.ContainerName,
			RestartPolicy: service.Restart,
			User:          service.User,
		})
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services, nil
}

func categoryFromComposePath(path string) string {
	return filepath.Base(filepath.Dir(filepath.Dir(path)))
}

func parseRuntime(serviceDir string, service composeService) (map[string]any, error) {
	runtime := make(map[string]any)
	if service.Image != "" {
		runtime["image"] = service.Image
	}
	if build, ok, err := parseBuild(serviceDir, service.Build); err != nil {
		return nil, err
	} else if ok {
		runtime["build"] = build
	}
	if command, ok := parseStringOrStringArray(service.Command); ok {
		runtime["command"] = command
	}
	if service.WorkingDir != "" {
		runtime["workingDir"] = service.WorkingDir
	}
	if len(runtime) == 0 {
		return nil, nil
	}
	return runtime, nil
}

func parseBuild(serviceDir string, value any) (map[string]any, bool, error) {
	if value == nil {
		return nil, false, nil
	}

	build := make(map[string]any)
	switch typed := value.(type) {
	case string:
		rebased, safe := rebasePathWithinService(serviceDir, typed)
		if !safe {
			return nil, false, fmt.Errorf("build context %q escapes the service entry", typed)
		}
		build["context"] = rebased
	case map[string]any:
		contextRaw, _ := typed["context"].(string)
		if contextRaw == "" {
			return nil, false, fmt.Errorf("build context is required")
		}
		rebased, safe := rebasePathWithinService(serviceDir, contextRaw)
		if !safe {
			return nil, false, fmt.Errorf("build context %q escapes the service entry", contextRaw)
		}
		build["context"] = rebased
		if dockerfile, _ := typed["dockerfile"].(string); dockerfile != "" {
			build["dockerfile"] = slashPath(dockerfile)
		}
		if target, _ := typed["target"].(string); target != "" {
			build["target"] = target
		}
		if argsRaw, ok := typed["args"]; ok {
			args := convertScalarMap(argsRaw)
			if len(args) > 0 {
				build["args"] = args
			}
		}
	case map[any]any:
		return parseBuild(serviceDir, convertAnyMap(typed))
	default:
		return nil, false, fmt.Errorf("unsupported build shape %T", value)
	}

	return build, true, nil
}

func convertScalarMap(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, raw := range typed {
			switch raw.(type) {
			case string, bool, int, int64, float64, float32, uint, uint64:
				result[key] = raw
			}
		}
		return result
	case map[any]any:
		return convertScalarMap(convertAnyMap(typed))
	default:
		return nil
	}
}

func convertAnyMap(value map[any]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, raw := range value {
		result[fmt.Sprint(key)] = raw
	}
	return result
}

func parseStringOrStringArray(value any) (any, bool) {
	if value == nil {
		return nil, false
	}
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil, false
		}
		return typed, true
	case []any:
		values := make([]string, 0, len(typed))
		for _, raw := range typed {
			text := strings.TrimSpace(fmt.Sprint(raw))
			if text == "" {
				continue
			}
			values = append(values, text)
		}
		if len(values) == 0 {
			return nil, false
		}
		return values, true
	case []string:
		if len(typed) == 0 {
			return nil, false
		}
		return append([]string(nil), typed...), true
	default:
		return nil, false
	}
}

var composePortPattern = regexp.MustCompile(`^(?:([0-9.]+):)?(\d+):(\d+)(?:/(\w+))?$`)

func parseComposePorts(values []string) []workspace.Port {
	ports := make([]workspace.Port, 0, len(values))
	for _, value := range values {
		matches := composePortPattern.FindStringSubmatch(strings.TrimSpace(value))
		if matches == nil {
			continue
		}
		hostIP := "127.0.0.1"
		if matches[1] != "" {
			hostIP = matches[1]
		}
		hostPort, _ := strconv.Atoi(matches[2])
		containerPort, _ := strconv.Atoi(matches[3])
		protocol := matches[4]
		if protocol == "" {
			protocol = "tcp"
		}
		ports = append(ports, workspace.Port{Host: hostPort, Container: containerPort, Protocol: protocol, HostIP: hostIP})
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Container != ports[j].Container {
			return ports[i].Container < ports[j].Container
		}
		if ports[i].Host != ports[j].Host {
			return ports[i].Host < ports[j].Host
		}
		return ports[i].HostIP < ports[j].HostIP
	})
	return ports
}

func parseComposeVolumes(values []string, namedVolumes map[string]any) []parsedVolume {
	volumes := make([]parsedVolume, 0, len(values))
	for _, value := range values {
		parts := strings.Split(value, ":")
		if len(parts) < 2 {
			continue
		}
		volume := parsedVolume{Source: parts[0], Target: parts[1]}
		if _, ok := namedVolumes[parts[0]]; ok {
			volume.Type = "named"
			volume.External = isExternalVolume(namedVolumes[parts[0]])
		} else if strings.HasPrefix(parts[0], ".") || strings.HasPrefix(parts[0], "/") {
			volume.Type = "bind"
		} else {
			volume.Type = "named"
		}
		if len(parts) > 2 && parts[2] == "ro" {
			volume.ReadOnly = true
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

func isExternalVolume(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		external, _ := typed["external"].(bool)
		return external
	case map[any]any:
		external, _ := typed["external"].(bool)
		return external
	default:
		return false
	}
}

func extractConfigMounts(volumes []parsedVolume, raw map[string]any) ([]parsedConfigMount, []parsedVolume) {
	mounts := make([]parsedConfigMount, 0)
	remaining := make([]parsedVolume, 0, len(volumes))
	for _, volume := range volumes {
		if volume.Type == "bind" {
			cleaned := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(volume.Source, "./"), "../"), "../")
			for strings.HasPrefix(cleaned, "../") {
				cleaned = strings.TrimPrefix(cleaned, "../")
			}
			if strings.HasPrefix(cleaned, "config/") {
				mounts = append(mounts, parsedConfigMount{SourcePath: slashPath(cleaned), TargetPath: volume.Target, ReadOnly: volume.ReadOnly})
				continue
			}
		}
		remaining = append(remaining, volume)
	}
	for _, mount := range parseDevarchConfig(raw) {
		mounts = append(mounts, mount)
	}
	sort.Slice(mounts, func(i, j int) bool {
		if mounts[i].TargetPath != mounts[j].TargetPath {
			return mounts[i].TargetPath < mounts[j].TargetPath
		}
		return mounts[i].SourcePath < mounts[j].SourcePath
	})
	return mounts, remaining
}

func parseDevarchConfig(raw map[string]any) []parsedConfigMount {
	xConfig, ok := raw["x-devarch-config"]
	if !ok {
		return nil
	}
	configMap, ok := xConfig.(map[string]any)
	if !ok {
		if converted, ok := xConfig.(map[any]any); ok {
			configMap = convertAnyMap(converted)
		} else {
			return nil
		}
	}
	mounts := make([]parsedConfigMount, 0, len(configMap))
	for filePath, targetRaw := range configMap {
		target, ok := targetRaw.(string)
		if !ok || strings.TrimSpace(target) == "" {
			continue
		}
		mounts = append(mounts, parsedConfigMount{SourcePath: slashPath(filepath.Join("config", filePath)), TargetPath: target})
	}
	return mounts
}

func parseComposeEnvironment(value any) map[string]any {
	if value == nil {
		return nil
	}
	env := make(map[string]any)
	switch typed := value.(type) {
	case []any:
		for _, raw := range typed {
			parts := strings.SplitN(fmt.Sprint(raw), "=", 2)
			if len(parts) == 0 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			if key == "" {
				continue
			}
			if len(parts) == 2 {
				env[key] = parts[1]
			} else {
				env[key] = ""
			}
		}
	case map[string]any:
		for key, raw := range typed {
			env[key] = normalizeScalar(raw)
		}
	case map[any]any:
		for key, raw := range typed {
			env[fmt.Sprint(key)] = normalizeScalar(raw)
		}
	}
	if len(env) == 0 {
		return nil
	}
	return env
}

func normalizeScalar(value any) any {
	switch typed := value.(type) {
	case string, bool, int, int64, float64, float32, uint, uint64:
		return typed
	default:
		return fmt.Sprint(value)
	}
}

func parseComposeStringList(value any) []string {
	if value == nil {
		return nil
	}
	var result []string
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) != "" {
			result = append(result, slashPath(typed))
		}
	case []any:
		for _, raw := range typed {
			text := strings.TrimSpace(fmt.Sprint(raw))
			if text == "" {
				continue
			}
			result = append(result, slashPath(text))
		}
	case []string:
		for _, raw := range typed {
			if strings.TrimSpace(raw) == "" {
				continue
			}
			result = append(result, slashPath(raw))
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func parseComposeDependsOn(value any) []string {
	if value == nil {
		return nil
	}
	deps := make([]string, 0)
	switch typed := value.(type) {
	case []any:
		for _, raw := range typed {
			deps = append(deps, strings.TrimSpace(fmt.Sprint(raw)))
		}
	case map[string]any:
		for key := range typed {
			deps = append(deps, strings.TrimSpace(key))
		}
	case map[any]any:
		for key := range typed {
			deps = append(deps, strings.TrimSpace(fmt.Sprint(key)))
		}
	}
	deps = compactStrings(deps)
	sort.Strings(deps)
	return deps
}

func parseComposeStringMap(value any) map[string]string {
	if value == nil {
		return nil
	}
	result := make(map[string]string)
	switch typed := value.(type) {
	case []any:
		for _, raw := range typed {
			parts := strings.SplitN(fmt.Sprint(raw), "=", 2)
			key := strings.TrimSpace(parts[0])
			if key == "" {
				continue
			}
			if len(parts) == 2 {
				result[key] = parts[1]
			} else {
				result[key] = ""
			}
		}
	case map[string]any:
		for key, raw := range typed {
			result[key] = fmt.Sprint(raw)
		}
	case map[any]any:
		for key, raw := range typed {
			result[fmt.Sprint(key)] = fmt.Sprint(raw)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func parseComposeHealth(value *composeHealthcheck) map[string]any {
	if value == nil {
		return nil
	}
	health := make(map[string]any)
	if test, ok := parseStringOrStringArray(value.Test); ok {
		health["test"] = test
	}
	if value.Interval != "" {
		health["interval"] = value.Interval
	}
	if value.Timeout != "" {
		health["timeout"] = value.Timeout
	}
	if value.Retries > 0 {
		health["retries"] = value.Retries
	}
	if value.StartPeriod != "" {
		health["startPeriod"] = value.StartPeriod
	}
	if len(health) == 0 {
		return nil
	}
	return health
}

func parseComposeNetworks(value any) []string {
	if value == nil {
		return nil
	}
	var networks []string
	switch typed := value.(type) {
	case []any:
		for _, raw := range typed {
			networks = append(networks, strings.TrimSpace(fmt.Sprint(raw)))
		}
	case map[string]any:
		for key := range typed {
			networks = append(networks, strings.TrimSpace(key))
		}
	case map[any]any:
		for key := range typed {
			networks = append(networks, strings.TrimSpace(fmt.Sprint(key)))
		}
	}
	networks = compactStrings(networks)
	sort.Strings(networks)
	return networks
}

func loadConfigFiles(path string) (map[string]importedConfigFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat config directory %s: %w", path, err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	files := make(map[string]importedConfigFile)
	err = filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fileInfo.IsDir() {
			return nil
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read config file %s: %w", filePath, err)
		}
		if isBinaryContent(content) {
			return nil
		}
		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return fmt.Errorf("compute config relative path for %s: %w", filePath, err)
		}
		mode := "0644"
		if fileInfo.Mode()&0o111 != 0 {
			mode = "0755"
		}
		files[slashPath(filepath.Join("config", relPath))] = importedConfigFile{Content: string(content), FileMode: mode}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk config directory %s: %w", path, err)
	}
	if len(files) == 0 {
		return nil, nil
	}
	return files, nil
}

func cloneConfigFiles(value map[string]importedConfigFile) map[string]importedConfigFile {
	if len(value) == 0 {
		return nil
	}
	cloned := make(map[string]importedConfigFile, len(value))
	for key, file := range value {
		cloned[key] = file
	}
	return cloned
}

func isBinaryContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	contentType := http.DetectContentType(data)
	if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "application/xml") {
		return false
	}
	if contentType == "application/octet-stream" {
		limit := data
		if len(limit) > 512 {
			limit = limit[:512]
		}
		for _, b := range limit {
			if b == 0 {
				return true
			}
		}
		return false
	}
	return true
}

func rebasePathWithinService(serviceDir, rawPath string) (string, bool) {
	if rawPath == "" {
		return "", false
	}
	if filepath.IsAbs(rawPath) {
		return "", false
	}
	resolved := filepath.Clean(filepath.Join(serviceDir, rawPath))
	rel, err := filepath.Rel(serviceDir, resolved)
	if err != nil {
		return "", false
	}
	rel = slashPath(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	if rel == "." {
		return ".", true
	}
	if strings.HasPrefix(rel, "config/") || rel == "config" {
		return "./" + rel, true
	}
	if strings.HasPrefix(rel, "./") {
		return rel, true
	}
	return "./" + rel, true
}

func compactStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
