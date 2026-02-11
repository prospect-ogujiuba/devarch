package orchestration

import "errors"

var (
	ErrStackNotFound  = errors.New("stack not found")
	ErrStackDisabled  = errors.New("stack is disabled")
	ErrStalePlan      = errors.New("plan is stale")
	ErrLockConflict   = errors.New("stack is being applied by another session")
	ErrProjectRoot    = errors.New("PROJECT_ROOT not set")
	ErrValidation     = errors.New("validation failed")
)
