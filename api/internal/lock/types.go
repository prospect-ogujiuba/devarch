package lock

type LockFile struct {
	Version     int                  `json:"version"`
	GeneratedAt string               `json:"generated_at"`
	ConfigHash  string               `json:"config_hash"`
	Stack       StackLock            `json:"stack"`
	Instances   map[string]InstLock  `json:"instances"`
}

type StackLock struct {
	Name        string `json:"name"`
	NetworkName string `json:"network_name"`
	NetworkID   string `json:"network_id,omitempty"`
}

type InstLock struct {
	Template      string         `json:"template"`
	TemplateHash  string         `json:"template_hash"`
	ImageTag      string         `json:"image_tag"`
	ImageDigest   string         `json:"image_digest"`
	ResolvedPorts map[string]int `json:"resolved_ports"`
}

type LockWarning struct {
	Severity string `json:"severity"`
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Message  string `json:"message"`
}

type ValidationResult struct {
	Valid           bool          `json:"valid"`
	Warnings        []LockWarning `json:"warnings"`
	ConfigHashMatch bool          `json:"config_hash_match"`
}
