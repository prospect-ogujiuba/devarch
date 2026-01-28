package nginx

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	db        *sql.DB
	outputDir string
}

func NewGenerator(db *sql.DB, outputDir string) *Generator {
	return &Generator{db: db, outputDir: outputDir}
}

type projectConf struct {
	Name               string
	Domain             string
	Upstream           string
	WebSocketUpstream  string
	ClientMaxBody      string
	LogName            string
}

type serviceConf struct {
	Domain      string
	VarName     string
	Upstream    string
	DisplayName string
	LogName     string
}

var defaultPorts = map[string]int{
	"laravel":   80,
	"php":       80,
	"wordpress": 80,
	"node":      3000,
	"go":        8080,
	"python":    8000,
	"rust":      8080,
	"dotnet":    80,
	"static":    80,
}

func (g *Generator) GenerateAll() error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if err := g.GenerateProjects(); err != nil {
		return err
	}
	return g.GenerateServices()
}

func (g *Generator) GenerateProjects() error {
	rows, err := g.db.Query(`
		SELECT name, project_type, domain, proxy_port, compose_path
		FROM projects
		WHERE domain IS NOT NULL AND domain != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	tmpl, err := template.New("project").Parse(projectTemplate)
	if err != nil {
		return err
	}

	for rows.Next() {
		var name, projectType string
		var domain sql.NullString
		var proxyPort sql.NullInt32
		var composePath sql.NullString

		if err := rows.Scan(&name, &projectType, &domain, &proxyPort, &composePath); err != nil {
			log.Printf("nginx: skip project scan error: %v", err)
			continue
		}

		if !domain.Valid || domain.String == "" {
			continue
		}

		port := defaultPorts[projectType]
		if proxyPort.Valid {
			port = int(proxyPort.Int32)
		}

		containerName := name + "-app"
		upstream := fmt.Sprintf("%s:%d", containerName, port)

		var wsUpstream string
		if projectType == "laravel" {
			var count int
			g.db.QueryRow(`SELECT COUNT(*) FROM project_services
				WHERE project_id = (SELECT id FROM projects WHERE name = $1)
				AND service_type = 'websocket'`, name).Scan(&count)
			if count > 0 {
				wsUpstream = name + "-reverb:8080"
			}
		}

		conf := projectConf{
			Name:              name,
			Domain:            domain.String,
			Upstream:          upstream,
			WebSocketUpstream: wsUpstream,
			LogName:           name,
		}

		if projectType == "laravel" || projectType == "wordpress" {
			conf.ClientMaxBody = "100M"
		}

		outPath := filepath.Join(g.outputDir, "project-"+name+".conf")
		if err := g.writeTemplate(tmpl, conf, outPath); err != nil {
			log.Printf("nginx: failed to generate %s: %v", name, err)
		}
	}
	return nil
}

func (g *Generator) GenerateProject(name string) error {
	tmpl, err := template.New("project").Parse(projectTemplate)
	if err != nil {
		return err
	}

	var projectType string
	var domain sql.NullString
	var proxyPort sql.NullInt32

	err = g.db.QueryRow(`SELECT project_type, domain, proxy_port FROM projects WHERE name = $1`, name).
		Scan(&projectType, &domain, &proxyPort)
	if err != nil {
		return err
	}

	if !domain.Valid || domain.String == "" {
		return nil
	}

	port := defaultPorts[projectType]
	if proxyPort.Valid {
		port = int(proxyPort.Int32)
	}

	containerName := name + "-app"
	upstream := fmt.Sprintf("%s:%d", containerName, port)

	var wsUpstream string
	if projectType == "laravel" {
		var count int
		g.db.QueryRow(`SELECT COUNT(*) FROM project_services
			WHERE project_id = (SELECT id FROM projects WHERE name = $1)
			AND service_type = 'websocket'`, name).Scan(&count)
		if count > 0 {
			wsUpstream = name + "-reverb:8080"
		}
	}

	conf := projectConf{
		Name:              name,
		Domain:            domain.String,
		Upstream:          upstream,
		WebSocketUpstream: wsUpstream,
		LogName:           name,
	}

	if projectType == "laravel" || projectType == "wordpress" {
		conf.ClientMaxBody = "100M"
	}

	outPath := filepath.Join(g.outputDir, "project-"+name+".conf")
	return g.writeTemplate(tmpl, conf, outPath)
}

func (g *Generator) GenerateServices() error {
	rows, err := g.db.Query(`
		SELECT s.name, sd.domain, sd.proxy_port
		FROM services s
		JOIN service_domains sd ON sd.service_id = s.id
		WHERE s.enabled = true`)
	if err != nil {
		return err
	}
	defer rows.Close()

	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return err
	}

	for rows.Next() {
		var name, domain string
		var proxyPort int

		if err := rows.Scan(&name, &domain, &proxyPort); err != nil {
			log.Printf("nginx: skip service scan error: %v", err)
			continue
		}

		varName := strings.ReplaceAll(name, "-", "_")
		upstream := fmt.Sprintf("%s:%d", name, proxyPort)

		conf := serviceConf{
			Domain:      domain,
			VarName:     varName,
			Upstream:    upstream,
			DisplayName: name,
			LogName:     name,
		}

		outPath := filepath.Join(g.outputDir, "service-"+name+".conf")
		if err := g.writeTemplate(tmpl, conf, outPath); err != nil {
			log.Printf("nginx: failed to generate service %s: %v", name, err)
		}
	}
	return nil
}

func (g *Generator) writeTemplate(tmpl *template.Template, data interface{}, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}
