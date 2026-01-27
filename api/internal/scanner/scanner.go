package scanner

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	hasWpConfig := fileExists(filepath.Join(dir, "wp-config.php")) || fileExists(filepath.Join(dir, "wp-config-sample.php"))
	hasArtisan := fileExists(filepath.Join(dir, "artisan"))

	switch {
	case hasComposer && hasArtisan:
		s.scanLaravel(&p, dir)
	case hasComposer && hasWpConfig:
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
		deps, _ := json.Marshal(req)
		p.Dependencies = deps
	}

	if scripts, ok := data["scripts"].(map[string]interface{}); ok {
		s, _ := json.Marshal(scripts)
		p.Scripts = s
	}

	if fileExists(filepath.Join(dir, "package.json")) {
		s.detectFrontend(p, dir)
	}
}

func (s *Scanner) scanWordPress(p *ProjectInfo, dir string) {
	p.ProjectType = "wordpress"
	p.Language = "php"
	p.PackageManager = "composer"
	p.EntryPoint = "index.php"
	p.Framework = "WordPress"

	data := readJSON(filepath.Join(dir, "composer.json"))
	if data != nil {
		if req, ok := data["require"].(map[string]interface{}); ok {
			deps, _ := json.Marshal(req)
			p.Dependencies = deps
		}
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
}

func (s *Scanner) scanRust(p *ProjectInfo, dir string) {
	p.ProjectType = "rust"
	p.Language = "rust"
	p.PackageManager = "cargo"
	p.EntryPoint = "src/main.rs"
}

func (s *Scanner) scanPython(p *ProjectInfo, dir string) {
	p.ProjectType = "python"
	p.Language = "python"

	if fileExists(filepath.Join(dir, "pyproject.toml")) {
		p.PackageManager = "pip"
	} else {
		p.PackageManager = "pip"
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
				p.Framework = "React"
			case name == "vue":
				p.Framework = "Vue"
			case name == "svelte":
				p.Framework = "Svelte"
			case name == "@angular/core":
				p.Framework = "Angular"
			case name == "express":
				p.Framework = "Express"
			case name == "fastify":
				p.Framework = "Fastify"
			}
		}
	}

	if scripts, ok := data["scripts"].(map[string]interface{}); ok {
		s, _ := json.Marshal(scripts)
		p.Scripts = s
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
