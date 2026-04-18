package projectscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Diagnostic reports non-fatal scan findings that callers may want to surface
// in human-readable or machine-readable output.
type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// ComposeService captures the small structured compose slice exposed by the
// Phase 4 scan command.
type ComposeService struct {
	Name        string   `json:"name"`
	Image       string   `json:"image,omitempty"`
	ServiceType string   `json:"serviceType,omitempty"`
	Ports       []string `json:"ports,omitempty"`
	DependsOn   []string `json:"dependsOn,omitempty"`
}

// Result is the transport-safe project scan shape used by the shared Phase 4
// service boundary and CLI.
type Result struct {
	Name               string           `json:"name"`
	Path               string           `json:"path"`
	ProjectType        string           `json:"projectType,omitempty"`
	Framework          string           `json:"framework,omitempty"`
	Language           string           `json:"language,omitempty"`
	PackageManager     string           `json:"packageManager,omitempty"`
	Description        string           `json:"description,omitempty"`
	Version            string           `json:"version,omitempty"`
	EntryPoint         string           `json:"entryPoint,omitempty"`
	HasFrontend        bool             `json:"hasFrontend,omitempty"`
	FrontendFramework  string           `json:"frontendFramework,omitempty"`
	ComposeFiles       []string         `json:"composeFiles,omitempty"`
	ServiceCount       int              `json:"serviceCount,omitempty"`
	Services           []ComposeService `json:"services,omitempty"`
	SuggestedTemplates []string         `json:"suggestedTemplates,omitempty"`
	Diagnostics        []Diagnostic     `json:"diagnostics,omitempty"`
}

type composeFile struct {
	Services map[string]composeServiceDef `yaml:"services"`
}

type composeServiceDef struct {
	Image     string      `yaml:"image"`
	Ports     interface{} `yaml:"ports"`
	DependsOn interface{} `yaml:"depends_on"`
}

// Scan inspects a project directory and returns a small structured summary plus
// suggested builtin templates when there is a clear Phase 4 mapping.
func Scan(path string) (*Result, error) {
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("scan project %s: %w", path, err)
	}
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("scan project %s: %w", cleanPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scan project %s: not a directory", cleanPath)
	}

	result := &Result{
		Name:        filepath.Base(cleanPath),
		Path:        cleanPath,
		ProjectType: "unknown",
	}

	hasComposer := fileExists(filepath.Join(cleanPath, "composer.json"))
	hasPackageJSON := fileExists(filepath.Join(cleanPath, "package.json"))
	hasGoMod := fileExists(filepath.Join(cleanPath, "go.mod"))
	hasArtisan := fileExists(filepath.Join(cleanPath, "artisan"))
	hasWPConfig := fileExists(filepath.Join(cleanPath, "wp-config.php")) || fileExists(filepath.Join(cleanPath, "wp-config-sample.php")) || fileExists(filepath.Join(cleanPath, "wp-includes", "version.php")) || fileExists(filepath.Join(cleanPath, "wp-content"))

	switch {
	case hasComposer && hasArtisan:
		scanLaravel(result, cleanPath)
	case hasWPConfig:
		scanWordPress(result, cleanPath)
	case hasGoMod:
		scanGo(result, cleanPath)
	case hasPackageJSON:
		scanNode(result, cleanPath)
	default:
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: "warning",
			Code:     "no-known-markers",
			Message:  "no known project markers found; scan output is limited",
		})
	}

	if hasPackageJSON && result.ProjectType != "node" {
		detectFrontend(result, cleanPath)
	}
	composeFiles, services, diagnostics := scanCompose(cleanPath)
	result.ComposeFiles = composeFiles
	result.Services = services
	result.ServiceCount = len(services)
	result.Diagnostics = append(result.Diagnostics, diagnostics...)
	result.SuggestedTemplates = suggestedTemplates(result)
	return result, nil
}

func scanLaravel(result *Result, dir string) {
	result.ProjectType = "laravel"
	result.Language = "php"
	result.PackageManager = "composer"
	result.EntryPoint = "artisan"

	data := readJSON(filepath.Join(dir, "composer.json"))
	if data == nil {
		return
	}
	result.Description = stringField(data, "description")
	if require, ok := data["require"].(map[string]any); ok {
		if version, ok := require["laravel/framework"].(string); ok {
			result.Framework = "Laravel " + version
		}
		if version, ok := require["php"].(string); ok {
			result.Version = version
		}
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		detectFrontend(result, dir)
	}
}

func scanWordPress(result *Result, dir string) {
	result.ProjectType = "wordpress"
	result.Language = "php"
	result.EntryPoint = "index.php"
	result.Framework = "WordPress"

	versionFile := filepath.Join(dir, "wp-includes", "version.php")
	if content, err := os.ReadFile(versionFile); err == nil {
		re := regexp.MustCompile(`\$wp_version\s*=\s*'([^']+)'`)
		if matches := re.FindSubmatch(content); len(matches) > 1 {
			result.Framework = "WordPress " + string(matches[1])
			result.Version = string(matches[1])
		}
	}
}

func scanGo(result *Result, dir string) {
	result.ProjectType = "go"
	result.Language = "go"
	result.PackageManager = "go mod"

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		fields := strings.Fields(strings.TrimSpace(lines[0]))
		if len(fields) >= 2 && fields[0] == "module" {
			result.Description = fields[1]
		}
	}
	inRequire := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "go "):
			result.Version = strings.TrimSpace(strings.TrimPrefix(trimmed, "go "))
		case trimmed == "require (":
			inRequire = true
		case trimmed == ")":
			inRequire = false
		case strings.HasPrefix(trimmed, "require ") && !strings.Contains(trimmed, "("):
			result.Framework = appendFramework(result.Framework, detectGoFramework(strings.Fields(strings.TrimPrefix(trimmed, "require "))))
		case inRequire:
			result.Framework = appendFramework(result.Framework, detectGoFramework(strings.Fields(trimmed)))
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, "cmd", "*", "main.go")); len(matches) > 0 {
		if rel, err := filepath.Rel(dir, matches[0]); err == nil {
			result.EntryPoint = rel
		}
	} else if fileExists(filepath.Join(dir, "main.go")) {
		result.EntryPoint = "main.go"
	}
}

func scanNode(result *Result, dir string) {
	result.ProjectType = "node"
	result.Language = "javascript"
	result.PackageManager = "npm"

	data := readJSON(filepath.Join(dir, "package.json"))
	if data == nil {
		return
	}
	result.Description = stringField(data, "description")
	result.Version = stringField(data, "version")
	result.EntryPoint = stringField(data, "main")
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		result.PackageManager = "yarn"
	} else if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		result.PackageManager = "pnpm"
	} else if fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")) {
		result.PackageManager = "bun"
	}

	if fileExists(filepath.Join(dir, "tsconfig.json")) {
		result.Language = "typescript"
	}

	result.Framework = detectNodeFramework(data)
	detectFrontend(result, dir)
}

func detectFrontend(result *Result, dir string) {
	data := readJSON(filepath.Join(dir, "package.json"))
	if data == nil {
		return
	}
	deps := mergeStringMaps(mapField(data, "dependencies"), mapField(data, "devDependencies"))
	for name := range deps {
		switch name {
		case "react", "react-dom":
			result.HasFrontend = true
			result.FrontendFramework = "React"
		case "vue":
			result.HasFrontend = true
			result.FrontendFramework = "Vue"
		case "svelte":
			result.HasFrontend = true
			result.FrontendFramework = "Svelte"
		case "@angular/core":
			result.HasFrontend = true
			result.FrontendFramework = "Angular"
		}
	}
}

func scanCompose(dir string) ([]string, []ComposeService, []Diagnostic) {
	candidates := []string{
		filepath.Join(dir, "docker-compose.yml"),
		filepath.Join(dir, "docker-compose.yaml"),
		filepath.Join(dir, "compose.yml"),
		filepath.Join(dir, "compose.yaml"),
		filepath.Join(dir, "deploy", "docker-compose.yml"),
		filepath.Join(dir, "deploy", "compose.yml"),
	}
	for _, candidate := range candidates {
		if !fileExists(candidate) {
			continue
		}
		services, diagnostics := parseCompose(candidate)
		return []string{candidate}, services, diagnostics
	}
	return nil, nil, nil
}

func parseCompose(path string) ([]ComposeService, []Diagnostic) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []Diagnostic{{Severity: "warning", Code: "compose-read-failed", Message: fmt.Sprintf("failed to read compose file %s: %v", path, err)}}
	}
	var compose composeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, []Diagnostic{{Severity: "warning", Code: "compose-parse-failed", Message: fmt.Sprintf("failed to parse compose file %s: %v", path, err)}}
	}
	if len(compose.Services) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(compose.Services))
	for key := range compose.Services {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	services := make([]ComposeService, 0, len(keys))
	for _, key := range keys {
		service := compose.Services[key]
		services = append(services, ComposeService{
			Name:        key,
			Image:       strings.TrimSpace(service.Image),
			ServiceType: detectServiceType(key, service.Image),
			Ports:       stringifyList(service.Ports),
			DependsOn:   stringifyList(service.DependsOn),
		})
	}
	return services, nil
}

func suggestedTemplates(result *Result) []string {
	if result == nil {
		return nil
	}
	seen := make(map[string]struct{})
	templates := make([]string, 0)
	add := func(name string) {
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		templates = append(templates, name)
	}

	switch result.ProjectType {
	case "laravel":
		add("laravel-app")
	case "node":
		if result.HasFrontend && !isNodeBackendFramework(result.Framework) {
			add("vite-web")
		} else {
			add("node-api")
		}
	case "wordpress", "go", "unknown":
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: "warning",
			Code:     "no-builtin-app-template",
			Message:  fmt.Sprintf("Phase 4 has no direct builtin app template for project type %q", result.ProjectType),
		})
	}

	hasDatabase := false
	hasCache := false
	hasProxy := false
	for _, service := range result.Services {
		switch service.ServiceType {
		case "database":
			hasDatabase = true
		case "cache":
			hasCache = true
		case "proxy":
			hasProxy = true
		}
	}
	if hasDatabase {
		add("postgres")
	}
	if hasCache {
		add("redis")
	}
	if hasProxy {
		add("nginx")
	}
	if len(templates) == 0 {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: "warning",
			Code:     "no-suggested-templates",
			Message:  "no builtin template suggestions were derived from the scanned project",
		})
	}
	return templates
}

func detectNodeFramework(data map[string]any) string {
	deps := mergeStringMaps(mapField(data, "dependencies"), mapField(data, "devDependencies"))
	for name := range deps {
		switch name {
		case "express":
			return "Express"
		case "fastify":
			return "Fastify"
		case "@nestjs/core":
			return "NestJS"
		case "hono":
			return "Hono"
		case "next":
			return "Next.js"
		case "nuxt":
			return "Nuxt"
		case "react":
			return "React"
		case "vue":
			return "Vue"
		case "svelte":
			return "Svelte"
		case "astro":
			return "Astro"
		case "gatsby":
			return "Gatsby"
		case "remix", "@remix-run/react":
			return "Remix"
		}
	}
	return ""
}

func detectGoFramework(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	module := fields[0]
	switch {
	case strings.HasPrefix(module, "github.com/gin-gonic/gin"):
		return "Gin"
	case strings.HasPrefix(module, "github.com/labstack/echo"):
		return "Echo"
	case strings.HasPrefix(module, "github.com/gofiber/fiber"):
		return "Fiber"
	case strings.HasPrefix(module, "github.com/go-chi/chi"):
		return "Chi"
	case strings.HasPrefix(module, "github.com/gorilla/mux"):
		return "Gorilla Mux"
	default:
		return ""
	}
}

func detectServiceType(name, image string) string {
	combined := strings.ToLower(name + " " + image)
	switch {
	case strings.Contains(combined, "nginx") || strings.Contains(combined, "traefik") || strings.Contains(combined, "caddy"):
		return "proxy"
	case strings.Contains(combined, "postgres") || strings.Contains(combined, "mysql") || strings.Contains(combined, "mariadb") || strings.Contains(combined, "mongo"):
		return "database"
	case strings.Contains(combined, "redis") || strings.Contains(combined, "memcache"):
		return "cache"
	default:
		return "app"
	}
}

func stringifyList(value any) []string {
	switch typed := value.(type) {
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, fmt.Sprint(item))
		}
		return items
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return keys
	default:
		return nil
	}
}

func mapField(data map[string]any, key string) map[string]any {
	value, _ := data[key].(map[string]any)
	return value
}

func stringField(data map[string]any, key string) string {
	value, _ := data[key].(string)
	return strings.TrimSpace(value)
}

func mergeStringMaps(values ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, value := range values {
		for key, item := range value {
			merged[key] = item
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func readJSON(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil
	}
	return decoded
}

func appendFramework(current, next string) string {
	if next == "" {
		return current
	}
	if current == "" {
		return next
	}
	if current == next || strings.Contains(current, next) {
		return current
	}
	return current + ", " + next
}

func isNodeBackendFramework(framework string) bool {
	switch framework {
	case "Express", "Fastify", "NestJS", "Hono":
		return true
	default:
		return false
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
