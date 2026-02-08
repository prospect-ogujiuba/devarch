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

type Plan struct {
	StackName   string    `json:"stack_name"`
	StackID     int       `json:"stack_id"`
	Changes     []Change  `json:"changes"`
	Token       string    `json:"token"`
	GeneratedAt time.Time `json:"generated_at"`
	Warnings    []string  `json:"warnings,omitempty"`
}
