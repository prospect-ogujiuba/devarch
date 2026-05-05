package workflows

// WorkflowStatus is the shared machine-readable status for workflow checks.
type WorkflowStatus string

const (
	StatusPass        WorkflowStatus = "pass"
	StatusWarn        WorkflowStatus = "warn"
	StatusFail        WorkflowStatus = "fail"
	StatusUnavailable WorkflowStatus = "unavailable"
)

// Diagnostic is a JSON-safe workflow finding.
type Diagnostic struct {
	ID       string         `json:"id"`
	Severity WorkflowStatus `json:"severity"`
	Message  string         `json:"message"`
	Detail   string         `json:"detail,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
}

// CheckResult records one pass/warn/fail check from a workflow.
type CheckResult struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Status      WorkflowStatus `json:"status"`
	Message     string         `json:"message"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
}

// CommandResult records host command execution in a JSON-safe shape.
type CommandResult struct {
	Command       string         `json:"command"`
	Args          []string       `json:"args,omitempty"`
	Status        WorkflowStatus `json:"status"`
	ExitCode      int            `json:"exitCode"`
	StdoutSummary string         `json:"stdoutSummary,omitempty"`
	StderrSummary string         `json:"stderrSummary,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// ReportStatus collapses check results into a workflow status.
func ReportStatus(checks []CheckResult) WorkflowStatus {
	status := StatusPass
	for _, check := range checks {
		switch check.Status {
		case StatusFail:
			return StatusFail
		case StatusWarn, StatusUnavailable:
			if status == StatusPass {
				status = check.Status
			}
		}
	}
	return status
}
