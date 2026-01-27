package scanner

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ProjectInfo struct {
	Name              string          `json:"name"`
	Path              string          `json:"path"`
	ProjectType       string          `json:"project_type"`
	Framework         string          `json:"framework,omitempty"`
	Language          string          `json:"language,omitempty"`
	PackageManager    string          `json:"package_manager,omitempty"`
	Description       string          `json:"description,omitempty"`
	Version           string          `json:"version,omitempty"`
	License           string          `json:"license,omitempty"`
	EntryPoint        string          `json:"entry_point,omitempty"`
	HasFrontend       bool            `json:"has_frontend"`
	FrontendFramework string          `json:"frontend_framework,omitempty"`
	Dependencies      json.RawMessage `json:"dependencies"`
	Scripts           json.RawMessage `json:"scripts"`
	GitRemote         string          `json:"git_remote,omitempty"`
	GitBranch         string          `json:"git_branch,omitempty"`
}

type Scanner struct {
	db      *sql.DB
	appsDir string
}

func NewScanner(db *sql.DB, appsDir string) *Scanner {
	return &Scanner{db: db, appsDir: appsDir}
}

func (s *Scanner) ScanAll() ([]ProjectInfo, error) {
	entries, err := os.ReadDir(s.appsDir)
	if err != nil {
		return nil, err
	}

	var projects []ProjectInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dir := filepath.Join(s.appsDir, entry.Name())
		info := s.scanProject(entry.Name(), dir)
		projects = append(projects, info)
	}

	return projects, nil
}

func (s *Scanner) ScanAndPersist() ([]ProjectInfo, error) {
	projects, err := s.ScanAll()
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if err := s.upsert(p); err != nil {
			log.Printf("failed to upsert project %s: %v", p.Name, err)
		}
	}

	return projects, nil
}

func (s *Scanner) scanProject(name, dir string) ProjectInfo {
	p := ProjectInfo{
		Name:         name,
		Path:         dir,
		ProjectType:  "unknown",
		Dependencies: json.RawMessage("{}"),
		Scripts:      json.RawMessage("{}"),
	}

	hasComposer := fileExists(filepath.Join(dir, "composer.json"))
	hasPackageJSON := fileExists(filepath.Join(dir, "package.json"))
	hasGoMod := fileExists(filepath.Join(dir, "go.mod"))
	hasCargo := fileExists(filepath.Join(dir, "Cargo.toml"))
	hasPyProject := fileExists(filepath.Join(dir, "pyproject.toml"))
	hasRequirements := fileExists(filepath.Join(dir, "requirements.txt"))
	hasWpConfig := fileExists(filepath.Join(dir, "wp-config.php")) || fileExists(filepath.Join(dir, "wp-config-sample.php")) ||
		fileExists(filepath.Join(dir, "wp-includes", "version.php")) || fileExists(filepath.Join(dir, "wp-content"))
	hasArtisan := fileExists(filepath.Join(dir, "artisan"))

	switch {
	case hasComposer && hasArtisan:
		s.scanLaravel(&p, dir)
	case hasComposer && hasWpConfig:
		s.scanWordPress(&p, dir)
	case hasWpConfig:
		s.scanWordPress(&p, dir)
	case hasComposer:
		s.scanComposer(&p, dir)
	case hasGoMod:
		s.scanGo(&p, dir)
	case hasCargo:
		s.scanRust(&p, dir)
	case hasPyProject || hasRequirements:
		s.scanPython(&p, dir)
	case hasPackageJSON:
		s.scanNode(&p, dir)
	default:
		if fileExists(filepath.Join(dir, "index.php")) {
			p.ProjectType = "php"
			p.Language = "php"
			p.EntryPoint = "index.php"
		}
	}

	if hasPackageJSON && p.ProjectType != "node" {
		s.detectFrontend(&p, dir)
	}

	s.scanGit(&p, dir)

	return p
}

func (s *Scanner) scanLaravel(p *ProjectInfo, dir string) {
	p.ProjectType = "laravel"
	p.Language = "php"
	p.PackageManager = "composer"
	p.EntryPoint = "artisan"

	data := readJSON(filepath.Join(dir, "composer.json"))
	if data == nil {
		return
	}

	if desc, ok := data["description"].(string); ok {
		p.Description = desc
	}
	if lic, ok := data["license"].(string); ok {
		p.License = lic
	}

	if req, ok := data["require"].(map[string]interface{}); ok {
		if v, ok := req["laravel/framework"].(string); ok {
			p.Framework = "Laravel " + v
		}
		if phpVer, ok := req["php"].(string); ok {
			p.Version = "PHP " + phpVer
		}
		// Detect sub-frameworks
		for name := range req {
			switch {
			case name == "livewire/livewire":
				p.Framework = appendFramework(p.Framework, "Livewire")
			case name == "inertiajs/inertia-laravel":
				p.Framework = appendFramework(p.Framework, "Inertia")
			case name == "filament/filament":
				p.Framework = appendFramework(p.Framework, "Filament")
			}
		}
		deps, _ := json.Marshal(req)
		p.Dependencies = deps
	}

	if scripts, ok := data["scripts"].(map[string]interface{}); ok {
		sc, _ := json.Marshal(scripts)
		p.Scripts = sc
	}

	// Fallback description from .env APP_NAME
	if p.Description == "" {
		if appName := readEnvValue(filepath.Join(dir, ".env"), "APP_NAME"); appName != "" {
			p.Description = appName
		}
	}

	if fileExists(filepath.Join(dir, "package.json")) {
		s.detectFrontend(p, dir)
	}
}

func (s *Scanner) scanWordPress(p *ProjectInfo, dir string) {
	p.ProjectType = "wordpress"
	p.Language = "php"
	p.EntryPoint = "index.php"
	p.Framework = "WordPress"

	// Parse wp version from wp-includes/version.php
	versionFile := filepath.Join(dir, "wp-includes", "version.php")
	if content, err := os.ReadFile(versionFile); err == nil {
		re := regexp.MustCompile(`\$wp_version\s*=\s*'([^']+)'`)
		if m := re.FindSubmatch(content); len(m) > 1 {
			p.Framework = "WordPress " + string(m[1])
			p.Version = string(m[1])
		}
	}

	// Build structured deps
	depsMap := make(map[string]interface{})

	// Plugins
	pluginsDir := filepath.Join(dir, "wp-content", "plugins")
	if entries, err := os.ReadDir(pluginsDir); err == nil {
		var plugins []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				plugins = append(plugins, e.Name())
			}
		}
		if len(plugins) > 0 {
			depsMap["plugins"] = plugins
		}
	}

	// Themes
	themesDir := filepath.Join(dir, "wp-content", "themes")
	if entries, err := os.ReadDir(themesDir); err == nil {
		var themes []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				themes = append(themes, e.Name())
			}
		}
		if len(themes) > 0 {
			depsMap["themes"] = themes
		}
	}

	// Detect active theme
	if themes, ok := depsMap["themes"].([]string); ok {
		for _, theme := range themes {
			stylePath := filepath.Join(themesDir, theme, "style.css")
			if content, err := os.ReadFile(stylePath); err == nil {
				re := regexp.MustCompile(`(?i)Theme Name:\s*(.+)`)
				if m := re.FindSubmatch(content); len(m) > 1 {
					themeName := strings.TrimSpace(string(m[1]))
					if p.Description == "" {
						p.Description = "Theme: " + themeName
					}
				}
			}
		}
	}

	// Composer deps
	if data := readJSON(filepath.Join(dir, "composer.json")); data != nil {
		p.PackageManager = "composer"
		if req, ok := data["require"].(map[string]interface{}); ok {
			depsMap["php"] = req
		}
	}

	// wp-config.php metadata (safe values only)
	wpConfig := filepath.Join(dir, "wp-config.php")
	if content, err := os.ReadFile(wpConfig); err == nil {
		src := string(content)
		if m := regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'([^']+)'`).FindStringSubmatch(src); len(m) > 1 {
			depsMap["db_name"] = m[1]
		}
		if m := regexp.MustCompile(`\$table_prefix\s*=\s*'([^']+)'`).FindStringSubmatch(src); len(m) > 1 {
			depsMap["table_prefix"] = m[1]
		}
	}

	if len(depsMap) > 0 {
		deps, _ := json.Marshal(depsMap)
		p.Dependencies = deps
	}
}

func (s *Scanner) scanComposer(p *ProjectInfo, dir string) {
	p.ProjectType = "php"
	p.Language = "php"
	p.PackageManager = "composer"

	data := readJSON(filepath.Join(dir, "composer.json"))
	if data == nil {
		return
	}

	if desc, ok := data["description"].(string); ok {
		p.Description = desc
	}
	if req, ok := data["require"].(map[string]interface{}); ok {
		deps, _ := json.Marshal(req)
		p.Dependencies = deps
	}
}

func (s *Scanner) scanGo(p *ProjectInfo, dir string) {
	p.ProjectType = "go"
	p.Language = "go"
	p.PackageManager = "go mod"

	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			p.Description = parts[1]
		}
	}

	// Parse go version
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") && !strings.HasPrefix(line, "go.") {
			p.Version = strings.TrimPrefix(line, "go ")
			break
		}
	}

	// Parse require block for deps and framework detection
	deps := make(map[string]interface{})
	inRequire := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "require (" {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}
		// Single-line require
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 2 {
				deps[parts[0]] = parts[1]
			}
			continue
		}
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 && !strings.HasPrefix(parts[0], "//") {
				deps[parts[0]] = parts[1]
			}
		}
	}

	// Framework detection
	frameworkMap := map[string]string{
		"github.com/gin-gonic/gin":   "Gin",
		"github.com/labstack/echo":   "Echo",
		"github.com/gofiber/fiber":   "Fiber",
		"github.com/go-chi/chi":      "Chi",
		"github.com/gorilla/mux":     "Gorilla Mux",
	}
	for dep := range deps {
		for prefix, fw := range frameworkMap {
			if strings.HasPrefix(dep, prefix) {
				p.Framework = fw
				break
			}
		}
		if p.Framework != "" {
			break
		}
	}

	if len(deps) > 0 {
		d, _ := json.Marshal(deps)
		p.Dependencies = d
	}

	// EntryPoint detection
	if matches, _ := filepath.Glob(filepath.Join(dir, "cmd", "*", "main.go")); len(matches) > 0 {
		rel, _ := filepath.Rel(dir, matches[0])
		p.EntryPoint = rel
	} else if fileExists(filepath.Join(dir, "main.go")) {
		p.EntryPoint = "main.go"
	}
}

func (s *Scanner) scanRust(p *ProjectInfo, dir string) {
	p.ProjectType = "rust"
	p.Language = "rust"
	p.PackageManager = "cargo"
	p.EntryPoint = "src/main.rs"

	content, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	section := ""
	deps := make(map[string]interface{})

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			section = strings.Trim(trimmed, "[]")
			continue
		}
		if !strings.Contains(trimmed, "=") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")

		switch section {
		case "package":
			switch key {
			case "name":
				if p.Description == "" {
					p.Description = val
				}
			case "version":
				p.Version = val
			case "description":
				p.Description = val
			case "edition":
				// store as metadata in deps
				deps["edition"] = val
			}
		case "dependencies":
			deps[key] = val
		}
	}

	// Framework detection
	frameworkMap := map[string]string{
		"actix-web": "Actix Web",
		"axum":      "Axum",
		"rocket":    "Rocket",
		"warp":      "Warp",
	}
	for dep := range deps {
		if fw, ok := frameworkMap[dep]; ok {
			p.Framework = fw
			break
		}
	}

	if len(deps) > 0 {
		d, _ := json.Marshal(deps)
		p.Dependencies = d
	}
}

func (s *Scanner) scanPython(p *ProjectInfo, dir string) {
	p.ProjectType = "python"
	p.Language = "python"
	p.PackageManager = "pip"

	deps := make(map[string]interface{})

	// Try pyproject.toml first
	pyprojectPath := filepath.Join(dir, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		lines := strings.Split(string(content), "\n")
		section := ""
		inDeps := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") {
				section = strings.Trim(trimmed, "[]")
				inDeps = section == "project.dependencies" || section == "tool.poetry.dependencies"
				continue
			}
			if !strings.Contains(trimmed, "=") && !inDeps {
				continue
			}

			if section == "project" || section == "tool.poetry" {
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, "\"")
				switch key {
				case "name":
					if p.Description == "" {
						p.Description = val
					}
				case "version":
					p.Version = val
				case "description":
					p.Description = val
				}
			}

			// Dependencies as array in [project]
			if inDeps {
				dep := strings.Trim(trimmed, " ,\"[]")
				if dep != "" && !strings.HasPrefix(dep, "#") {
					// Parse "package>=1.0" style
					re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)(.*)$`)
					if m := re.FindStringSubmatch(dep); len(m) > 1 {
						deps[m[1]] = strings.TrimSpace(m[2])
					}
				}
			}

			// Poetry-style key = "version" deps
			if section == "tool.poetry.dependencies" {
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					val = strings.Trim(val, "\"")
					deps[key] = val
				}
			}
		}

		// Check for poetry
		if strings.Contains(string(content), "[tool.poetry]") {
			p.PackageManager = "poetry"
		}
	}

	// Fallback: requirements.txt
	reqPath := filepath.Join(dir, "requirements.txt")
	if len(deps) == 0 {
		if content, err := os.ReadFile(reqPath); err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
					continue
				}
				re := regexp.MustCompile(`^([a-zA-Z0-9_.-]+)(.*)$`)
				if m := re.FindStringSubmatch(line); len(m) > 1 {
					deps[m[1]] = strings.TrimSpace(m[2])
				}
			}
		}
	}

	// Framework detection
	frameworkMap := map[string]string{
		"django":    "Django",
		"Django":    "Django",
		"flask":     "Flask",
		"Flask":     "Flask",
		"fastapi":   "FastAPI",
		"FastAPI":   "FastAPI",
		"starlette": "Starlette",
		"Starlette": "Starlette",
	}
	for dep := range deps {
		if fw, ok := frameworkMap[dep]; ok {
			p.Framework = fw
			break
		}
	}

	// EntryPoint detection
	if fileExists(filepath.Join(dir, "manage.py")) {
		p.EntryPoint = "manage.py"
	} else if fileExists(filepath.Join(dir, "app.py")) {
		p.EntryPoint = "app.py"
	} else if fileExists(filepath.Join(dir, "main.py")) {
		p.EntryPoint = "main.py"
	}

	if len(deps) > 0 {
		d, _ := json.Marshal(deps)
		p.Dependencies = d
	}
}

func (s *Scanner) scanNode(p *ProjectInfo, dir string) {
	p.ProjectType = "node"
	p.Language = "javascript"

	data := readJSON(filepath.Join(dir, "package.json"))
	if data == nil {
		return
	}

	if desc, ok := data["description"].(string); ok {
		p.Description = desc
	}
	if ver, ok := data["version"].(string); ok {
		p.Version = ver
	}
	if lic, ok := data["license"].(string); ok {
		p.License = lic
	}
	if main, ok := data["main"].(string); ok {
		p.EntryPoint = main
	}

	p.PackageManager = "npm"
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		p.PackageManager = "yarn"
	} else if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		p.PackageManager = "pnpm"
	} else if fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")) {
		p.PackageManager = "bun"
	}

	// Detect monorepo
	if _, ok := data["workspaces"]; ok {
		if p.Description == "" {
			p.Description = "monorepo"
		} else {
			p.Description = p.Description + " (monorepo)"
		}
	}

	if deps, ok := data["dependencies"].(map[string]interface{}); ok {
		d, _ := json.Marshal(deps)
		p.Dependencies = d

		for name := range deps {
			switch {
			case name == "next":
				p.Framework = "Next.js"
				p.Language = "typescript"
			case name == "nuxt":
				p.Framework = "Nuxt"
			case name == "react":
				if p.Framework == "" {
					p.Framework = "React"
				}
			case name == "vue":
				if p.Framework == "" {
					p.Framework = "Vue"
				}
			case name == "svelte":
				p.Framework = "Svelte"
			case name == "@angular/core":
				p.Framework = "Angular"
			case name == "express":
				if p.Framework == "" {
					p.Framework = "Express"
				}
			case name == "fastify":
				if p.Framework == "" {
					p.Framework = "Fastify"
				}
			case name == "remix" || name == "@remix-run/react":
				p.Framework = "Remix"
			case name == "astro":
				p.Framework = "Astro"
			case name == "gatsby":
				p.Framework = "Gatsby"
			case name == "@nestjs/core":
				p.Framework = "NestJS"
			case name == "hono":
				p.Framework = "Hono"
			}
		}
	}

	if scripts, ok := data["scripts"].(map[string]interface{}); ok {
		sc, _ := json.Marshal(scripts)
		p.Scripts = sc
	}

	if fileExists(filepath.Join(dir, "tsconfig.json")) {
		p.Language = "typescript"
	}
}

func (s *Scanner) detectFrontend(p *ProjectInfo, dir string) {
	data := readJSON(filepath.Join(dir, "package.json"))
	if data == nil {
		return
	}

	deps, _ := data["dependencies"].(map[string]interface{})
	devDeps, _ := data["devDependencies"].(map[string]interface{})

	allDeps := make(map[string]interface{})
	for k, v := range deps {
		allDeps[k] = v
	}
	for k, v := range devDeps {
		allDeps[k] = v
	}

	for name := range allDeps {
		switch name {
		case "react", "react-dom":
			p.HasFrontend = true
			p.FrontendFramework = "React"
		case "vue":
			p.HasFrontend = true
			p.FrontendFramework = "Vue"
		case "svelte":
			p.HasFrontend = true
			p.FrontendFramework = "Svelte"
		case "@angular/core":
			p.HasFrontend = true
			p.FrontendFramework = "Angular"
		}
	}

	clientDir := filepath.Join(dir, "client")
	if fileExists(filepath.Join(clientDir, "package.json")) {
		p.HasFrontend = true
		clientData := readJSON(filepath.Join(clientDir, "package.json"))
		if clientData != nil {
			if clientDeps, ok := clientData["dependencies"].(map[string]interface{}); ok {
				for name := range clientDeps {
					switch name {
					case "react":
						p.FrontendFramework = "React"
					case "vue":
						p.FrontendFramework = "Vue"
					}
				}
			}
		}
	}
}

func (s *Scanner) scanGit(p *ProjectInfo, dir string) {
	gitDir := filepath.Join(dir, ".git")
	if !fileExists(gitDir) {
		return
	}

	if out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output(); err == nil {
		p.GitRemote = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "-C", dir, "branch", "--show-current").Output(); err == nil {
		p.GitBranch = strings.TrimSpace(string(out))
	}
}

func (s *Scanner) upsert(p ProjectInfo) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO projects (name, path, project_type, framework, language, package_manager,
			description, version, license, entry_point, has_frontend, frontend_framework,
			dependencies, scripts, git_remote, git_branch, last_scanned_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		ON CONFLICT (name) DO UPDATE SET
			path=EXCLUDED.path, project_type=EXCLUDED.project_type, framework=EXCLUDED.framework,
			language=EXCLUDED.language, package_manager=EXCLUDED.package_manager,
			description=EXCLUDED.description, version=EXCLUDED.version, license=EXCLUDED.license,
			entry_point=EXCLUDED.entry_point, has_frontend=EXCLUDED.has_frontend,
			frontend_framework=EXCLUDED.frontend_framework, dependencies=EXCLUDED.dependencies,
			scripts=EXCLUDED.scripts, git_remote=EXCLUDED.git_remote, git_branch=EXCLUDED.git_branch,
			last_scanned_at=EXCLUDED.last_scanned_at, updated_at=EXCLUDED.updated_at`,
		p.Name, p.Path, p.ProjectType,
		nullStr(p.Framework), nullStr(p.Language), nullStr(p.PackageManager),
		nullStr(p.Description), nullStr(p.Version), nullStr(p.License),
		nullStr(p.EntryPoint), p.HasFrontend, nullStr(p.FrontendFramework),
		p.Dependencies, p.Scripts,
		nullStr(p.GitRemote), nullStr(p.GitBranch),
		now, now, now,
	)
	return err
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readJSON(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func readEnvValue(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	prefix := key + "="
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimPrefix(line, prefix)
			val = strings.Trim(val, "\"'")
			return val
		}
	}
	return ""
}

func appendFramework(base, addition string) string {
	if base == "" {
		return addition
	}
	return base + " + " + addition
}
