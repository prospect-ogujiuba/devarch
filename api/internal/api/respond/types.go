package respond

type SuccessEnvelope struct {
	Data interface{} `json:"data"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ActionResponse provides consistent structure for action endpoints.
// Status is always present; all other fields are optional.
type ActionResponse struct {
	Status   string                 `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Output   string                 `json:"output,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
