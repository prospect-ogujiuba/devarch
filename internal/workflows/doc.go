// Package workflows contains Go replacements for legacy shell-script operator workflows.
//
// The package is intentionally transport-free: CLI and HTTP layers call appsvc,
// appsvc delegates here, and this package returns JSON-safe models instead of
// printing shell-oriented text.
package workflows
