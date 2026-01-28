package compose

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/priz/devarch-api/pkg/models"
	"gopkg.in/yaml.v3"
)

type Generator struct {
	db          *sql.DB
	networkName string
}

func NewGenerator(db *sql.DB, networkName string) *Generator {
	return &Generator{
		db:          db,
		networkName: networkName,
	}
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
	Image         string                 `yaml:"image"`
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
		Image:         fmt.Sprintf("%s:%s", service.ImageName, service.ImageTag),
		ContainerName: service.Name,
		Restart:       service.RestartPolicy,
		Networks:      []string{g.networkName},
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

	for _, vol := range service.Volumes {
		volStr := fmt.Sprintf("%s:%s", vol.Source, vol.Target)
		if vol.ReadOnly {
			volStr += ":ro"
		}
		svc.Volumes = append(svc.Volumes, volStr)

		if vol.VolumeType == "named" {
			namedVolumes[vol.Source] = nil
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

			for k, v := range overrides {
				svcMap[k] = v
			}

			rawCompose := map[string]interface{}{
				"networks": compose.Networks,
				"services": map[string]interface{}{service.Name: svcMap},
			}
			if len(compose.Volumes) > 0 {
				rawCompose["volumes"] = compose.Volumes
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
		SELECT volume_type, source, target, read_only
		FROM service_volumes WHERE service_id = $1
	`, service.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var v models.ServiceVolume
		if err := rows.Scan(&v.VolumeType, &v.Source, &v.Target, &v.ReadOnly); err != nil {
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
