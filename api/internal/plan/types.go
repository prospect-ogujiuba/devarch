package plan

import "time"

type Action string

const (
	ActionAdd    Action = "add"
	ActionModify Action = "modify"
	ActionRemove Action = "remove"
)

type FieldChange struct {
	Old    interface{} `json:"old"`
	New    interface{} `json:"new"`
	Source string      `json:"source"`
}

type Change struct {
	Action        Action                 `json:"action"`
	InstanceID    string                 `json:"instance_id"`
	TemplateName  string                 `json:"template_name"`
	ContainerName string                 `json:"container_name"`
	Fields        map[string]FieldChange `json:"fields,omitempty"`
}

type WiringSection struct {
	ActiveWires []WirePlanEntry `json:"active_wires"`
	Warnings    []WiringWarning `json:"warnings"`
}

type WirePlanEntry struct {
	ConsumerInstance string            `json:"consumer_instance"`
	ProviderInstance string            `json:"provider_instance"`
	ContractName     string            `json:"contract_name"`
	ContractType     string            `json:"contract_type"`
	Source           string            `json:"source"`
	InjectedEnvVars  map[string]string `json:"injected_env_vars"`
}

type ResourceLimitEntry struct {
	CPULimit          string `json:"cpu_limit,omitempty"`
	CPUReservation    string `json:"cpu_reservation,omitempty"`
	MemoryLimit       string `json:"memory_limit,omitempty"`
	MemoryReservation string `json:"memory_reservation,omitempty"`
}

type WiringWarning struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Instance string `json:"instance"`
	Contract string `json:"contract,omitempty"`
}

type Plan struct {
	StackName       string                         `json:"stack_name"`
	StackID         int                            `json:"stack_id"`
	Changes         []Change                       `json:"changes"`
	Token           string                         `json:"token"`
	GeneratedAt     time.Time                      `json:"generated_at"`
	Warnings        []string                       `json:"warnings,omitempty"`
	Wiring          *WiringSection                 `json:"wiring,omitempty"`
	ResourceLimits  map[string]ResourceLimitEntry  `json:"resource_limits,omitempty"`
}
