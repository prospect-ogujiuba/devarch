package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Category struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name,omitempty"`
	Color        string    `json:"color,omitempty"`
	StartupOrder int       `json:"startup_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Service struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	CategoryID    int            `json:"category_id"`
	ImageName     string         `json:"image_name"`
	ImageTag      string         `json:"image_tag"`
	RestartPolicy string         `json:"restart_policy"`
	Command       sql.NullString `json:"-"`
	CommandStr    string         `json:"command,omitempty"`
	UserSpec      sql.NullString `json:"-"`
	UserSpecStr   string         `json:"user_spec,omitempty"`
	Enabled       bool           `json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	Category     *Category         `json:"category,omitempty"`
	Ports        []ServicePort     `json:"ports,omitempty"`
	Volumes      []ServiceVolume   `json:"volumes,omitempty"`
	EnvVars      []ServiceEnvVar   `json:"env_vars,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Healthcheck  *ServiceHealthcheck `json:"healthcheck,omitempty"`
	Labels       []ServiceLabel    `json:"labels,omitempty"`
	Domains      []ServiceDomain   `json:"domains,omitempty"`

	ConfigStatus     string          `json:"config_status"`
	LastValidatedAt  sql.NullTime    `json:"-"`
	LastValidatedStr *time.Time      `json:"last_validated_at,omitempty"`
	ValidationErrors NullableJSON    `json:"validation_errors,omitempty"`

	Status  *ContainerState   `json:"status,omitempty"`
	Metrics *ContainerMetrics `json:"metrics,omitempty"`
}

type NullableJSON struct {
	Data  json.RawMessage
	Valid bool
}

func (n *NullableJSON) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case []byte:
		n.Data = json.RawMessage(v)
	case string:
		n.Data = json.RawMessage(v)
	}
	return nil
}

func (n NullableJSON) MarshalJSON() ([]byte, error) {
	if !n.Valid || n.Data == nil {
		return []byte("null"), nil
	}
	return n.Data, nil
}

type ConfigVersion struct {
	ID             int             `json:"id"`
	ServiceID      int             `json:"service_id"`
	Version        int             `json:"version"`
	ConfigSnapshot json.RawMessage `json:"config_snapshot"`
	ChangeSummary  string          `json:"change_summary,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ServicePort struct {
	ID            int    `json:"id"`
	ServiceID     int    `json:"service_id"`
	HostIP        string `json:"host_ip"`
	HostPort      int    `json:"host_port"`
	ContainerPort int    `json:"container_port"`
	Protocol      string `json:"protocol"`
}

type ServiceVolume struct {
	ID         int    `json:"id"`
	ServiceID  int    `json:"service_id"`
	VolumeType string `json:"volume_type"`
	Source     string `json:"source"`
	Target     string `json:"target"`
	ReadOnly   bool   `json:"read_only"`
}

type ServiceEnvVar struct {
	ID        int    `json:"id"`
	ServiceID int    `json:"service_id"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	IsSecret  bool   `json:"is_secret"`
}

type ServiceDependency struct {
	ID                 int    `json:"id"`
	ServiceID          int    `json:"service_id"`
	DependsOnServiceID int    `json:"depends_on_service_id"`
	Condition          string `json:"condition"`
}

type ServiceHealthcheck struct {
	ID                 int    `json:"id"`
	ServiceID          int    `json:"service_id"`
	Test               string `json:"test"`
	IntervalSeconds    int    `json:"interval_seconds"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	Retries            int    `json:"retries"`
	StartPeriodSeconds int    `json:"start_period_seconds"`
}

type ServiceLabel struct {
	ID        int    `json:"id"`
	ServiceID int    `json:"service_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

type ServiceDomain struct {
	ID        int    `json:"id"`
	ServiceID int    `json:"service_id"`
	Domain    string `json:"domain"`
	ProxyPort int    `json:"proxy_port,omitempty"`
}

type ContainerState struct {
	ID           int            `json:"id"`
	ServiceID    int            `json:"service_id"`
	ContainerID  sql.NullString `json:"-"`
	ContainerStr string         `json:"container_id,omitempty"`
	Status       string         `json:"status"`
	HealthStatus sql.NullString `json:"-"`
	HealthStr    string         `json:"health_status,omitempty"`
	RestartCount int            `json:"restart_count"`
	StartedAt    sql.NullTime   `json:"-"`
	StartedAtStr *time.Time     `json:"started_at,omitempty"`
	FinishedAt   sql.NullTime   `json:"-"`
	FinishedStr  *time.Time     `json:"finished_at,omitempty"`
	ExitCode     sql.NullInt32  `json:"-"`
	ExitCodeInt  *int           `json:"exit_code,omitempty"`
	Error        sql.NullString `json:"-"`
	ErrorStr     string         `json:"error,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type ContainerMetrics struct {
	ID               int       `json:"id"`
	ServiceID        int       `json:"service_id"`
	CPUPercentage    float64   `json:"cpu_percentage"`
	MemoryUsedMB     float64   `json:"memory_used_mb"`
	MemoryLimitMB    float64   `json:"memory_limit_mb"`
	MemoryPercentage float64   `json:"memory_percentage"`
	NetworkRxBytes   int64     `json:"network_rx_bytes"`
	NetworkTxBytes   int64     `json:"network_tx_bytes"`
	RecordedAt       time.Time `json:"recorded_at"`
}

type Registry struct {
	ID                 int            `json:"id"`
	Name               string         `json:"name"`
	BaseURL            string         `json:"base_url"`
	APIVersion         sql.NullString `json:"-"`
	APIVersionStr      string         `json:"api_version,omitempty"`
	Enabled            bool           `json:"enabled"`
	RateLimitRemaining sql.NullInt32  `json:"-"`
	RateLimitInt       *int           `json:"rate_limit_remaining,omitempty"`
	RateLimitResetAt   sql.NullTime   `json:"-"`
	RateLimitResetStr  *time.Time     `json:"rate_limit_reset_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type Image struct {
	ID           int            `json:"id"`
	RegistryID   int            `json:"registry_id"`
	Repository   string         `json:"repository"`
	Description  sql.NullString `json:"-"`
	DescriptionStr string       `json:"description,omitempty"`
	StarCount    sql.NullInt32  `json:"-"`
	StarCountInt *int           `json:"star_count,omitempty"`
	PullCount    sql.NullInt64  `json:"-"`
	PullCountInt *int64         `json:"pull_count,omitempty"`
	IsOfficial   bool           `json:"is_official"`
	LastSyncedAt sql.NullTime   `json:"-"`
	LastSynced   *time.Time     `json:"last_synced_at,omitempty"`
}

type ImageTag struct {
	ID           int            `json:"id"`
	ImageID      int            `json:"image_id"`
	Tag          string         `json:"tag"`
	Digest       sql.NullString `json:"-"`
	DigestStr    string         `json:"digest,omitempty"`
	SizeBytes    sql.NullInt64  `json:"-"`
	SizeBytesInt *int64         `json:"size_bytes,omitempty"`
	PushedAt     sql.NullTime   `json:"-"`
	PushedAtStr  *time.Time     `json:"pushed_at,omitempty"`
	LastSyncedAt sql.NullTime   `json:"-"`
	LastSynced   *time.Time     `json:"last_synced_at,omitempty"`
}

type ProjectService struct {
	ID            int             `json:"id"`
	ProjectID     int             `json:"project_id"`
	ServiceName   string          `json:"service_name"`
	ContainerName sql.NullString  `json:"-"`
	ContainerStr  string          `json:"container_name,omitempty"`
	Image         sql.NullString  `json:"-"`
	ImageStr      string          `json:"image,omitempty"`
	ServiceType   sql.NullString  `json:"-"`
	ServiceTypeStr string         `json:"service_type,omitempty"`
	Ports         json.RawMessage `json:"ports"`
	DependsOn     json.RawMessage `json:"depends_on"`
}

func (ps *ProjectService) ResolveNulls() {
	if ps.ContainerName.Valid {
		ps.ContainerStr = ps.ContainerName.String
	}
	if ps.Image.Valid {
		ps.ImageStr = ps.Image.String
	}
	if ps.ServiceType.Valid {
		ps.ServiceTypeStr = ps.ServiceType.String
	}
}

type Project struct {
	ID                int             `json:"id"`
	Name              string          `json:"name"`
	Path              string          `json:"path"`
	ProjectType       string          `json:"project_type"`
	Framework         sql.NullString  `json:"-"`
	FrameworkStr      string          `json:"framework,omitempty"`
	Language          sql.NullString  `json:"-"`
	LanguageStr       string          `json:"language,omitempty"`
	PackageManager    sql.NullString  `json:"-"`
	PackageManagerStr string          `json:"package_manager,omitempty"`
	Description       sql.NullString  `json:"-"`
	DescriptionStr    string          `json:"description,omitempty"`
	Version           sql.NullString  `json:"-"`
	VersionStr        string          `json:"version,omitempty"`
	License           sql.NullString  `json:"-"`
	LicenseStr        string          `json:"license,omitempty"`
	EntryPoint        sql.NullString  `json:"-"`
	EntryPointStr     string          `json:"entry_point,omitempty"`
	HasFrontend       bool            `json:"has_frontend"`
	FrontendFramework sql.NullString  `json:"-"`
	FrontendFwStr     string          `json:"frontend_framework,omitempty"`
	Domain            sql.NullString  `json:"-"`
	DomainStr         string          `json:"domain,omitempty"`
	ProxyPort         sql.NullInt32   `json:"-"`
	ProxyPortInt      *int            `json:"proxy_port,omitempty"`
	Dependencies      json.RawMessage `json:"dependencies"`
	Scripts           json.RawMessage `json:"scripts"`
	ComposePath       sql.NullString  `json:"-"`
	ComposePathStr    string          `json:"compose_path,omitempty"`
	ServiceCount      int             `json:"service_count"`
	GitRemote         sql.NullString  `json:"-"`
	GitRemoteStr      string          `json:"git_remote,omitempty"`
	GitBranch         sql.NullString  `json:"-"`
	GitBranchStr      string          `json:"git_branch,omitempty"`
	LastScannedAt     sql.NullTime    `json:"-"`
	LastScannedAtStr  *time.Time      `json:"last_scanned_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (p *Project) ResolveNulls() {
	if p.Framework.Valid {
		p.FrameworkStr = p.Framework.String
	}
	if p.Language.Valid {
		p.LanguageStr = p.Language.String
	}
	if p.PackageManager.Valid {
		p.PackageManagerStr = p.PackageManager.String
	}
	if p.Description.Valid {
		p.DescriptionStr = p.Description.String
	}
	if p.Version.Valid {
		p.VersionStr = p.Version.String
	}
	if p.License.Valid {
		p.LicenseStr = p.License.String
	}
	if p.EntryPoint.Valid {
		p.EntryPointStr = p.EntryPoint.String
	}
	if p.FrontendFramework.Valid {
		p.FrontendFwStr = p.FrontendFramework.String
	}
	if p.Domain.Valid {
		p.DomainStr = p.Domain.String
	}
	if p.ProxyPort.Valid {
		v := int(p.ProxyPort.Int32)
		p.ProxyPortInt = &v
	}
	if p.ComposePath.Valid {
		p.ComposePathStr = p.ComposePath.String
	}
	if p.GitRemote.Valid {
		p.GitRemoteStr = p.GitRemote.String
	}
	if p.GitBranch.Valid {
		p.GitBranchStr = p.GitBranch.String
	}
	if p.LastScannedAt.Valid {
		p.LastScannedAtStr = &p.LastScannedAt.Time
	}
}

type Vulnerability struct {
	ID          int            `json:"id"`
	CVEID       string         `json:"cve_id"`
	Severity    string         `json:"severity"`
	Title       sql.NullString `json:"-"`
	TitleStr    string         `json:"title,omitempty"`
	Description sql.NullString `json:"-"`
	DescStr     string         `json:"description,omitempty"`
	CVSSScore   sql.NullFloat64 `json:"-"`
	CVSSFloat   *float64       `json:"cvss_score,omitempty"`
	PublishedAt sql.NullTime   `json:"-"`
	PublishedStr *time.Time    `json:"published_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}
