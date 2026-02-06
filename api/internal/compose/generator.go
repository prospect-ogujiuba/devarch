package compose

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/priz/devarch-api/pkg/models"
	"gopkg.in/yaml.v3"
)

type Generator struct {
	db              *sql.DB
	networkName     string
	projectRoot     string
	hostProjectRoot string
}

func NewGenerator(db *sql.DB, networkName string) *Generator {
	return &Generator{
		db:          db,
		networkName: networkName,
	}
}

func (g *Generator) SetProjectRoot(root string) {
	g.projectRoot = root
}

func (g *Generator) SetHostProjectRoot(root string) {
	g.hostProjectRoot = root
}

type generatedCompose struct {
	Networks map[string]networkConfig `yaml:"networks"`
	Volumes  map[string]interface{}   `yaml:"volumes,omitempty"`
	Services map[string]serviceConfig `yaml:"services"`
}

type networkConfig struct {
	External bool `yaml:"external"`
}

type serviceConfig struct {
	Image         string                 `yaml:"image,omitempty"`
	ContainerName string                 `yaml:"container_name"`
	Restart       string                 `yaml:"restart,omitempty"`
	Command       interface{}            `yaml:"command,omitempty"`
	User          string                 `yaml:"user,omitempty"`
	Ports         []string               `yaml:"ports,omitempty"`
	Volumes       []string               `yaml:"volumes,omitempty"`
	Environment   map[string]string      `yaml:"environment,omitempty"`
	DependsOn     []string               `yaml:"depends_on,omitempty"`
	Labels        []string               `yaml:"labels,omitempty"`
	Healthcheck   *healthcheckConfig     `yaml:"healthcheck,omitempty"`
	Networks      []string               `yaml:"networks"`
}

type healthcheckConfig struct {
	Test        []string `yaml:"test"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	Retries     int      `yaml:"retries"`
	StartPeriod string   `yaml:"start_period,omitempty"`
}

func (g *Generator) Generate(service *models.Service) ([]byte, error) {
	if err := g.loadServiceRelations(service); err != nil {
		return nil, fmt.Errorf("load relations: %w", err)
	}

	compose := generatedCompose{
		Networks: map[string]networkConfig{
			g.networkName: {External: true},
		},
		Services: make(map[string]serviceConfig),
	}

	svc := serviceConfig{
		ContainerName: service.Name,
		Restart:       service.RestartPolicy,
		Networks:      []string{g.networkName},
	}

	if service.ImageName != "" {
		svc.Image = fmt.Sprintf("%s:%s", service.ImageName, service.ImageTag)
	}

	if service.Command.Valid && service.Command.String != "" {
		svc.Command = service.Command.String
	}
	if service.UserSpec.Valid && service.UserSpec.String != "" {
		svc.User = service.UserSpec.String
	}
	namedVolumes := make(map[string]interface{})
	for _, port := range service.Ports {
		portStr := fmt.Sprintf("%s:%d:%d", port.HostIP, port.HostPort, port.ContainerPort)
		if port.Protocol != "tcp" {
			portStr += "/" + port.Protocol
		}
		svc.Ports = append(svc.Ports, portStr)
	}

	// Get category name for path rewriting
	var categoryName string
	g.db.QueryRow(`SELECT c.name FROM categories c JOIN services s ON s.category_id = c.id WHERE s.id = $1`, service.ID).Scan(&categoryName)

	for _, vol := range service.Volumes {
		source := g.rewriteConfigPath(vol.Source, categoryName, service.Name)
		source = g.resolveRelativePath(source, categoryName)
		volStr := fmt.Sprintf("%s:%s", source, vol.Target)
		if vol.ReadOnly {
			volStr += ":ro"
		}
		svc.Volumes = append(svc.Volumes, volStr)

		isNamed := vol.VolumeType == "named" || extractNamedVolume(volStr) != ""
		if isNamed {
			if vol.IsExternal {
				namedVolumes[vol.Source] = map[string]interface{}{"external": true}
			} else {
				namedVolumes[vol.Source] = nil
			}
		}
	}

	if len(namedVolumes) > 0 {
		compose.Volumes = namedVolumes
	}

	if len(service.EnvVars) > 0 {
		svc.Environment = make(map[string]string)
		for _, env := range service.EnvVars {
			svc.Environment[env.Key] = env.Value
		}
	}

	svc.DependsOn = service.Dependencies

	for _, label := range service.Labels {
		svc.Labels = append(svc.Labels, fmt.Sprintf("%s=%s", label.Key, label.Value))
	}

	if service.Healthcheck != nil {
		hc := &healthcheckConfig{
			Retries: service.Healthcheck.Retries,
		}

		testParts := strings.Fields(service.Healthcheck.Test)
		if len(testParts) > 0 && testParts[0] == "CMD" {
			hc.Test = testParts
		} else {
			hc.Test = append([]string{"CMD-SHELL"}, service.Healthcheck.Test)
		}

		hc.Interval = fmt.Sprintf("%ds", service.Healthcheck.IntervalSeconds)
		hc.Timeout = fmt.Sprintf("%ds", service.Healthcheck.TimeoutSeconds)
		if service.Healthcheck.StartPeriodSeconds > 0 {
			hc.StartPeriod = fmt.Sprintf("%ds", service.Healthcheck.StartPeriodSeconds)
		}

		svc.Healthcheck = hc
	}

	if service.ComposeOverrides.Valid && len(service.ComposeOverrides.Data) > 2 {
		var overrides map[string]interface{}
		if err := json.Unmarshal(service.ComposeOverrides.Data, &overrides); err == nil && len(overrides) > 0 {
			svcBytes, _ := yaml.Marshal(svc)
			var svcMap map[string]interface{}
			yaml.Unmarshal(svcBytes, &svcMap)

			// Resolve relative paths in override volumes
			if volumes, ok := overrides["volumes"]; ok {
				if volList, ok := volumes.([]interface{}); ok {
					for i, v := range volList {
						if volStr, ok := v.(string); ok {
							volName := extractNamedVolume(volStr)
							if volName != "" && namedVolumes[volName] == nil {
								namedVolumes[volName] = nil
							}
							volList[i] = g.resolveRelativeVolStr(volStr, categoryName)
						}
					}
					overrides["volumes"] = volList
				}
			}

			// Resolve relative build context paths
			if build, ok := overrides["build"]; ok {
				if buildMap, ok := build.(map[string]interface{}); ok {
					if ctx, ok := buildMap["context"].(string); ok {
						buildMap["context"] = g.resolveRelativePath(ctx, categoryName)
					}
				} else if ctx, ok := build.(string); ok {
					overrides["build"] = g.resolveRelativePath(ctx, categoryName)
				}
			}

			// Merge overrides into service map (after path resolution)
			for k, v := range overrides {
				svcMap[k] = v
			}

			rawCompose := map[string]interface{}{
				"networks": compose.Networks,
				"services": map[string]interface{}{service.Name: svcMap},
			}
			if len(namedVolumes) > 0 {
				rawCompose["volumes"] = namedVolumes
			}
			return yaml.Marshal(rawCompose)
		}
	}

	compose.Services[service.Name] = svc

	return yaml.Marshal(compose)
}

func (g *Generator) loadServiceRelations(service *models.Service) error {
	rows, err := g.db.Query(`
		SELECT host_ip, host_port, container_port, protocol
		FROM service_ports WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.ServicePort
		if err := rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol); err != nil {
			return err
		}
		service.Ports = append(service.Ports, p)
	}

	rows, err = g.db.Query(`
		SELECT volume_type, source, target, read_only, is_external
		FROM service_volumes WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var v models.ServiceVolume
		if err := rows.Scan(&v.VolumeType, &v.Source, &v.Target, &v.ReadOnly, &v.IsExternal); err != nil {
			return err
		}
		service.Volumes = append(service.Volumes, v)
	}

	rows, err = g.db.Query(`
		SELECT key, value, is_secret
		FROM service_env_vars WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.ServiceEnvVar
		if err := rows.Scan(&e.Key, &e.Value, &e.IsSecret); err != nil {
			return err
		}
		service.EnvVars = append(service.EnvVars, e)
	}

	rows, err = g.db.Query(`
		SELECT s.name FROM service_dependencies d
		JOIN services s ON d.depends_on_service_id = s.id
		WHERE d.service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		service.Dependencies = append(service.Dependencies, name)
	}

	rows, err = g.db.Query(`
		SELECT key, value FROM service_labels WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.ServiceLabel
		if err := rows.Scan(&l.Key, &l.Value); err != nil {
			return err
		}
		service.Labels = append(service.Labels, l)
	}

	var hc models.ServiceHealthcheck
	err = g.db.QueryRow(`
		SELECT test, interval_seconds, timeout_seconds, retries, start_period_seconds
		FROM service_healthchecks WHERE service_id = $1
	`, service.ID).Scan(&hc.Test, &hc.IntervalSeconds, &hc.TimeoutSeconds, &hc.Retries, &hc.StartPeriodSeconds)
	if err == nil {
		service.Healthcheck = &hc
	}

	return nil
}

// rewriteConfigPath rewrites volume source paths from config/{service}/ to compose/{category}/{service}/
func (g *Generator) rewriteConfigPath(source, categoryName, serviceName string) string {
	if g.projectRoot == "" || categoryName == "" {
		return source
	}

	configPrefix := filepath.Join(g.projectRoot, "config", serviceName)
	if strings.HasPrefix(source, configPrefix) {
		relPart := strings.TrimPrefix(source, configPrefix)
		return filepath.Join(g.projectRoot, "compose", categoryName, serviceName, relPart)
	}

	// Also handle relative config/ paths
	if strings.HasPrefix(source, "config/"+serviceName) {
		relPart := strings.TrimPrefix(source, "config/"+serviceName)
		return filepath.Join("compose", categoryName, serviceName, relPart)
	}

	return source
}

// MaterializeConfigFiles writes all config files from DB to compose/{category}/{service}/
func (g *Generator) MaterializeConfigFiles(service *models.Service, baseDir string) error {
	var categoryName string
	err := g.db.QueryRow(`SELECT c.name FROM categories c JOIN services s ON s.category_id = c.id WHERE s.id = $1`, service.ID).Scan(&categoryName)
	if err != nil {
		return fmt.Errorf("get category: %w", err)
	}

	rows, err := g.db.Query(`
		SELECT file_path, content, file_mode
		FROM service_config_files WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return fmt.Errorf("query config files: %w", err)
	}
	defer rows.Close()

	serviceDir := filepath.Join(baseDir, "compose", categoryName, service.Name)

	for rows.Next() {
		var filePath, content, fileMode string
		if err := rows.Scan(&filePath, &content, &fileMode); err != nil {
			return err
		}

		fullPath := filepath.Join(serviceDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}

		mode := parseFileMode(fileMode)
		if err := os.WriteFile(fullPath, []byte(content), mode); err != nil {
			return fmt.Errorf("write %s: %w", fullPath, err)
		}
	}

	return rows.Err()
}

// LoadConfigFiles loads config files for a service from DB into the model
func (g *Generator) LoadConfigFiles(service *models.Service) error {
	rows, err := g.db.Query(`
		SELECT id, service_id, file_path, content, file_mode, is_template, created_at, updated_at
		FROM service_config_files WHERE service_id = $1 ORDER BY file_path
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var f models.ServiceConfigFile
		if err := rows.Scan(&f.ID, &f.ServiceID, &f.FilePath, &f.Content, &f.FileMode, &f.IsTemplate, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return err
		}
		service.ConfigFiles = append(service.ConfigFiles, f)
	}
	return rows.Err()
}

func parseFileMode(mode string) os.FileMode {
	m, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0644
	}
	return os.FileMode(m)
}

// resolveRelativePath resolves a relative path against the original compose file directory on the host.
// The original compose files live in {hostProjectRoot}/apps/{category}/.
func (g *Generator) resolveRelativePath(source, categoryName string) string {
	if g.hostProjectRoot == "" || categoryName == "" {
		return source
	}
	if filepath.IsAbs(source) || source == "" {
		return source
	}
	// Skip named-volume-style sources (no dots or slashes at start)
	if !strings.HasPrefix(source, ".") && !strings.Contains(source, "/") {
		return source
	}
	base := filepath.Join(g.hostProjectRoot, "apps", categoryName)
	return filepath.Clean(filepath.Join(base, source))
}

// resolveRelativeVolStr resolves relative bind mount paths within a volume string (source:target[:opts]).
func (g *Generator) resolveRelativeVolStr(volStr, categoryName string) string {
	parts := strings.SplitN(volStr, ":", 3)
	if len(parts) < 2 {
		return volStr
	}
	resolved := g.resolveRelativePath(parts[0], categoryName)
	parts[0] = resolved
	return strings.Join(parts, ":")
}

// extractNamedVolume checks if a volume string is a named volume and returns the name
func extractNamedVolume(volStr string) string {
	parts := strings.Split(volStr, ":")
	if len(parts) < 2 {
		return ""
	}
	source := parts[0]
	// Named volumes don't start with . / or ~
	if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") {
		return ""
	}
	return source
}
