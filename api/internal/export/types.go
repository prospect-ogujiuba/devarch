package export

type DevArchFile struct {
	Version   int                    `yaml:"version"`
	Stack     StackConfig            `yaml:"stack"`
	Instances map[string]InstanceDef `yaml:"instances"`
	Wires     []WireDef              `yaml:"wires,omitempty"`
}

type WireDef struct {
	ConsumerInstance string `yaml:"consumer_instance"`
	ProviderInstance string `yaml:"provider_instance"`
	ImportContract   string `yaml:"import_contract"`
	ExportContract   string `yaml:"export_contract"`
	Source           string `yaml:"source"`
}

type StackConfig struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	NetworkName string `yaml:"network_name"`
}

type InstanceDef struct {
	Template     string                  `yaml:"template"`
	Enabled      bool                    `yaml:"enabled"`
	Image        string                  `yaml:"image,omitempty"`
	Ports        []PortDef               `yaml:"ports,omitempty"`
	Volumes      []VolumeDef             `yaml:"volumes,omitempty"`
	Environment  map[string]string       `yaml:"environment,omitempty"`
	EnvFiles     []string                `yaml:"env_files,omitempty"`
	Networks     []string                `yaml:"networks,omitempty"`
	Labels       map[string]string       `yaml:"labels,omitempty"`
	Domains      []DomainDef             `yaml:"domains,omitempty"`
	Healthcheck  *HealthcheckDef         `yaml:"healthcheck,omitempty"`
	Dependencies []string                `yaml:"dependencies,omitempty"`
	ConfigFiles  map[string]ConfigFileDef `yaml:"config_files,omitempty"`
	ConfigMounts []ConfigMountDef        `yaml:"config_mounts,omitempty"`
	Command      string                  `yaml:"command,omitempty"`
}

type PortDef struct {
	HostIP        string `yaml:"host_ip"`
	HostPort      int    `yaml:"host_port"`
	ContainerPort int    `yaml:"container_port"`
	Protocol      string `yaml:"protocol"`
}

type VolumeDef struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"read_only,omitempty"`
}

type DomainDef struct {
	Domain    string `yaml:"domain"`
	ProxyPort int    `yaml:"proxy_port"`
}

type HealthcheckDef struct {
	Test        string `yaml:"test"`
	Interval    string `yaml:"interval,omitempty"`
	Timeout     string `yaml:"timeout,omitempty"`
	Retries     int    `yaml:"retries,omitempty"`
	StartPeriod string `yaml:"start_period,omitempty"`
}

type ConfigFileDef struct {
	Content    string `yaml:"content"`
	FileMode   string `yaml:"file_mode"`
	IsTemplate bool   `yaml:"is_template,omitempty"`
}

type ConfigMountDef struct {
	SourcePath     string `yaml:"source_path"`
	TargetPath     string `yaml:"target_path"`
	ReadOnly       bool   `yaml:"read_only,omitempty"`
	ConfigFilePath string `yaml:"config_file_path,omitempty"`
}
