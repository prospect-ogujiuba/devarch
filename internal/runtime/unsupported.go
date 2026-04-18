package runtime

import "fmt"

// UnsupportedOperationError reports an adapter capability that is intentionally
// unavailable for a provider in this phase.
type UnsupportedOperationError struct {
	Provider  string
	Operation string
	Reason    string
}

func (e *UnsupportedOperationError) Error() string {
	if e == nil {
		return "unsupported operation"
	}
	if e.Reason == "" {
		return fmt.Sprintf("runtime %q does not support operation %q", e.Provider, e.Operation)
	}
	return fmt.Sprintf("runtime %q does not support operation %q: %s", e.Provider, e.Operation, e.Reason)
}

func unsupportedOperation(provider, operation, reason string) error {
	return &UnsupportedOperationError{Provider: provider, Operation: operation, Reason: reason}
}

func UnsupportedSourceDiagnostic(workspaceName, resourceKey, sourceType string) Diagnostic {
	return Diagnostic{
		Severity:  SeverityError,
		Code:      "unsupported-source-type",
		Workspace: workspaceName,
		Resource:  resourceKey,
		Message:   fmt.Sprintf("resource %q uses unsupported source.type %q in Phase 3", resourceKey, sourceType),
	}
}

func UnsupportedFieldDiagnostic(workspaceName, resourceKey, code, message string) Diagnostic {
	return Diagnostic{
		Severity:  SeverityError,
		Code:      code,
		Workspace: workspaceName,
		Resource:  resourceKey,
		Message:   message,
	}
}
