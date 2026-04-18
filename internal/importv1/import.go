package importv1

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	StatusSucceeded = "succeeded"
	StatusPartial   = "partial"
	StatusRejected  = "rejected"

	ModeV1Stack   = "v1-stack"
	ModeV1Library = "v1-library"

	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityError   = "error"

	ArtifactKindTemplate  = "template"
	ArtifactKindWorkspace = "workspace"
)

// Diagnostic reports structured importer findings that transports can forward
// without scraping human-readable output.
type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Path     string `json:"path,omitempty"`
	Field    string `json:"field,omitempty"`
}

// Artifact is one emitted Phase 6 import output plus any per-file diagnostics.
type Artifact struct {
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	SourcePath    string       `json:"sourcePath,omitempty"`
	SuggestedPath string       `json:"suggestedPath,omitempty"`
	Status        string       `json:"status"`
	Document      string       `json:"document,omitempty"`
	Diagnostics   []Diagnostic `json:"diagnostics,omitempty"`
}

// Summary reports aggregate artifact outcomes.
type Summary struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Partial   int `json:"partial"`
	Rejected  int `json:"rejected"`
}

// Result is the transport-safe Phase 6 V1 import response.
type Result struct {
	Mode        string       `json:"mode"`
	SourcePath  string       `json:"sourcePath"`
	Status      string       `json:"status"`
	Message     string       `json:"message,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Artifacts   []Artifact   `json:"artifacts,omitempty"`
	Summary     Summary      `json:"summary,omitempty"`
}

// Preview remains as a compatibility alias for the Phase 4 service boundary.
type Preview = Result

func ImportStack(path string) (*Result, error) {
	cleanPath, err := resolveSourcePath(path, false)
	if err != nil {
		return nil, err
	}
	return importStack(cleanPath)
}

func ImportLibrary(path string) (*Result, error) {
	cleanPath, err := resolveSourcePath(path, true)
	if err != nil {
		return nil, err
	}
	return importLibrary(cleanPath)
}

// PrepareStackImport preserves the Phase 4 public function name while backing
// it with the real Phase 6 importer implementation.
func PrepareStackImport(path string) (*Preview, error) {
	return ImportStack(path)
}

// PrepareLibraryImport preserves the Phase 4 public function name while
// backing it with the real Phase 6 importer implementation.
func PrepareLibraryImport(path string) (*Preview, error) {
	return ImportLibrary(path)
}

func resolveSourcePath(path string, wantDir bool) (string, error) {
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolve import source %s: %w", path, err)
	}
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", fmt.Errorf("stat import source %s: %w", cleanPath, err)
	}
	if wantDir && !info.IsDir() {
		return "", fmt.Errorf("import source %s: expected directory", cleanPath)
	}
	if !wantDir && info.IsDir() {
		return "", fmt.Errorf("import source %s: expected file", cleanPath)
	}
	return cleanPath, nil
}

func newResult(mode, sourcePath string) *Result {
	return &Result{Mode: mode, SourcePath: slashPath(sourcePath)}
}

func finalizeResult(result *Result) *Result {
	if result == nil {
		return nil
	}

	sortDiagnostics(result.Diagnostics)
	for i := range result.Artifacts {
		result.Artifacts[i].SourcePath = slashPath(result.Artifacts[i].SourcePath)
		result.Artifacts[i].SuggestedPath = slashPath(result.Artifacts[i].SuggestedPath)
		sortDiagnostics(result.Artifacts[i].Diagnostics)
		if result.Artifacts[i].Status == "" {
			result.Artifacts[i].Status = deriveArtifactStatus(result.Artifacts[i].Document, result.Artifacts[i].Diagnostics)
		}
	}
	sortArtifacts(result.Artifacts)

	result.Summary = summarizeArtifacts(result.Artifacts)
	result.Status = deriveResultStatus(result.Artifacts, result.Diagnostics)
	if result.Message == "" {
		result.Message = buildResultMessage(result)
	}
	return result
}

func deriveArtifactStatus(document string, diagnostics []Diagnostic) string {
	if document == "" {
		return StatusRejected
	}
	if hasSeverity(diagnostics, SeverityError) || hasSeverity(diagnostics, SeverityWarning) {
		return StatusPartial
	}
	return StatusSucceeded
}

func summarizeArtifacts(artifacts []Artifact) Summary {
	summary := Summary{Total: len(artifacts)}
	for _, artifact := range artifacts {
		switch artifact.Status {
		case StatusSucceeded:
			summary.Succeeded++
		case StatusPartial:
			summary.Partial++
		case StatusRejected:
			summary.Rejected++
		}
	}
	return summary
}

func deriveResultStatus(artifacts []Artifact, diagnostics []Diagnostic) string {
	hasArtifactOutput := false
	hasPartial := false
	hasRejected := false
	for _, artifact := range artifacts {
		switch artifact.Status {
		case StatusSucceeded:
			hasArtifactOutput = true
		case StatusPartial:
			hasArtifactOutput = true
			hasPartial = true
		case StatusRejected:
			hasRejected = true
		}
	}

	hasWarning := hasSeverity(diagnostics, SeverityWarning)
	hasError := hasSeverity(diagnostics, SeverityError)

	switch {
	case !hasArtifactOutput && (hasRejected || hasError):
		return StatusRejected
	case !hasArtifactOutput && len(artifacts) == 0:
		return StatusRejected
	case hasRejected || hasPartial || hasWarning || hasError:
		return StatusPartial
	default:
		return StatusSucceeded
	}
}

func buildResultMessage(result *Result) string {
	if result == nil {
		return ""
	}
	if result.Summary.Total == 0 {
		return fmt.Sprintf("%s import produced no emitted artifacts.", result.Mode)
	}
	return fmt.Sprintf(
		"%s import emitted %d artifact(s): %d succeeded, %d partial, %d rejected.",
		result.Mode,
		result.Summary.Total,
		result.Summary.Succeeded,
		result.Summary.Partial,
		result.Summary.Rejected,
	)
}

func hasSeverity(diagnostics []Diagnostic, severity string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == severity {
			return true
		}
	}
	return false
}

func warning(code, message, path, field string) Diagnostic {
	return Diagnostic{Severity: SeverityWarning, Code: code, Message: message, Path: slashPath(path), Field: field}
}

func info(code, message, path, field string) Diagnostic {
	return Diagnostic{Severity: SeverityInfo, Code: code, Message: message, Path: slashPath(path), Field: field}
}

func failure(code, message, path, field string) Diagnostic {
	return Diagnostic{Severity: SeverityError, Code: code, Message: message, Path: slashPath(path), Field: field}
}

func slashPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func sortDiagnostics(diagnostics []Diagnostic) {
	sort.Slice(diagnostics, func(i, j int) bool {
		if severityRank(diagnostics[i].Severity) != severityRank(diagnostics[j].Severity) {
			return severityRank(diagnostics[i].Severity) < severityRank(diagnostics[j].Severity)
		}
		if diagnostics[i].Code != diagnostics[j].Code {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		if diagnostics[i].Path != diagnostics[j].Path {
			return diagnostics[i].Path < diagnostics[j].Path
		}
		if diagnostics[i].Field != diagnostics[j].Field {
			return diagnostics[i].Field < diagnostics[j].Field
		}
		return diagnostics[i].Message < diagnostics[j].Message
	})
}

func severityRank(severity string) int {
	switch severity {
	case SeverityError:
		return 0
	case SeverityWarning:
		return 1
	default:
		return 2
	}
}

func sortArtifacts(artifacts []Artifact) {
	sort.Slice(artifacts, func(i, j int) bool {
		if artifacts[i].Kind != artifacts[j].Kind {
			return artifacts[i].Kind < artifacts[j].Kind
		}
		if artifacts[i].SuggestedPath != artifacts[j].SuggestedPath {
			return artifacts[i].SuggestedPath < artifacts[j].SuggestedPath
		}
		if artifacts[i].Name != artifacts[j].Name {
			return artifacts[i].Name < artifacts[j].Name
		}
		return artifacts[i].SourcePath < artifacts[j].SourcePath
	})
}
