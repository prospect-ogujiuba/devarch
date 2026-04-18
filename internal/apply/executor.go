package apply

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cachepkg "github.com/prospect-ogujiuba/devarch/internal/cache"
	"github.com/prospect-ogujiuba/devarch/internal/events"
	"github.com/prospect-ogujiuba/devarch/internal/plan"
	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

var ErrBlocked = errors.New("apply blocked by diagnostics")

type Executor struct {
	Adapter   runtimepkg.Adapter
	Cache     cachepkg.Store
	Publisher events.Publisher
	Now       func() time.Time
}

func (e *Executor) Execute(ctx context.Context, diff *plan.Result, payload *Payload) (*Result, error) {
	if e == nil || e.Adapter == nil {
		return nil, fmt.Errorf("apply execute: nil adapter")
	}
	if diff == nil {
		return nil, fmt.Errorf("apply execute: nil plan result")
	}
	if payload == nil {
		return nil, fmt.Errorf("apply execute: nil payload")
	}
	if diff.Workspace != payload.Workspace {
		return nil, fmt.Errorf("apply execute: plan workspace %q does not match payload workspace %q", diff.Workspace, payload.Workspace)
	}
	if diff.Blocked || payload.Blocked {
		return nil, ErrBlocked
	}

	now := e.Now
	if now == nil {
		now = time.Now
	}
	startedAt := now()
	result := &Result{Workspace: payload.Workspace, Provider: payload.Provider, StartedAt: startedAt}
	succeeded := false
	store := cachepkg.Normalize(e.Cache)
	defer func() {
		result.FinishedAt = now()
		_ = store.SaveApply(ctx, cachepkg.ApplyRecord{
			Workspace:  result.Workspace,
			Provider:   result.Provider,
			StartedAt:  result.StartedAt,
			FinishedAt: result.FinishedAt,
			Succeeded:  succeeded,
			Operations: cacheOperations(result.Operations),
		})
	}()

	if e.Publisher != nil {
		if _, err := e.Publisher.Publish(events.ApplyStarted(payload.Workspace, len(diff.Actions))); err != nil {
			return nil, err
		}
	}

	for _, action := range diff.Actions {
		operation := Operation{Scope: action.Scope, Target: action.Target, RuntimeName: action.RuntimeName, Kind: action.Kind}
		message := strings.Join(action.Reasons, "; ")
		operation.Message = message
		switch action.Kind {
		case plan.ActionNoop:
			operation.Status = "skipped"
			result.Operations = append(result.Operations, operation)
			if err := e.publishProgress(payload.Workspace, action, operation.Status, message); err != nil {
				return nil, err
			}
			continue
		case plan.ActionAdd, plan.ActionModify, plan.ActionRemove, plan.ActionRestart:
		default:
			return nil, fmt.Errorf("apply execute: unsupported plan action %q", action.Kind)
		}

		if err := e.publishProgress(payload.Workspace, action, "started", message); err != nil {
			return nil, err
		}
		err := e.executeAction(ctx, action, payload)
		if err != nil {
			operation.Status = "failed"
			if operation.Message != "" {
				operation.Message += ": " + err.Error()
			} else {
				operation.Message = err.Error()
			}
			result.Operations = append(result.Operations, operation)
			_ = e.publishProgress(payload.Workspace, action, operation.Status, operation.Message)
			return result, err
		}
		operation.Status = "success"
		result.Operations = append(result.Operations, operation)
		if err := e.publishProgress(payload.Workspace, action, operation.Status, message); err != nil {
			return nil, err
		}
	}

	if e.Adapter.Capabilities().Inspect {
		desiredSnapshotBoundary := desiredBoundaryFromPayload(payload)
		snapshot, err := e.Adapter.InspectWorkspace(ctx, desiredSnapshotBoundary)
		if err == nil {
			result.Snapshot = snapshot
			_ = store.SaveSnapshot(ctx, cachepkg.SnapshotRecord{Workspace: payload.Workspace, CapturedAt: now(), Snapshot: snapshot})
		}
	}
	succeeded = true
	if e.Publisher != nil {
		if _, err := e.Publisher.Publish(events.ApplyCompleted(payload.Workspace, true, len(result.Operations))); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (e *Executor) executeAction(ctx context.Context, action plan.Action, payload *Payload) error {
	switch action.Scope {
	case plan.ScopeWorkspace:
		if payload.Network == nil {
			return fmt.Errorf("workspace action %q requires network payload", action.Kind)
		}
		network := &runtimepkg.DesiredNetwork{Name: payload.Network.Name, Labels: cloneStringMap(payload.Network.Labels)}
		switch action.Kind {
		case plan.ActionAdd, plan.ActionModify:
			return e.Adapter.EnsureNetwork(ctx, network)
		case plan.ActionRemove:
			return e.Adapter.RemoveNetwork(ctx, network)
		default:
			return nil
		}
	case plan.ScopeResource:
		resource := payload.Resource(action.Target)
		ref := runtimepkg.ResourceRef{Workspace: payload.Workspace, Key: action.Target, RuntimeName: action.RuntimeName}
		if resource != nil {
			ref.RuntimeName = resource.RuntimeName
		}
		switch action.Kind {
		case plan.ActionAdd, plan.ActionModify:
			if resource == nil {
				return fmt.Errorf("resource payload %q not found", action.Target)
			}
			return e.Adapter.ApplyResource(ctx, runtimepkg.ApplyResourceRequest{Workspace: payload.Workspace, NetworkName: networkName(payload), Resource: applyResource(resource)})
		case plan.ActionRemove:
			return e.Adapter.RemoveResource(ctx, ref)
		case plan.ActionRestart:
			return e.Adapter.RestartResource(ctx, ref)
		default:
			return nil
		}
	default:
		return fmt.Errorf("apply execute: unsupported action scope %q", action.Scope)
	}
}

func (e *Executor) publishProgress(workspace string, action plan.Action, status, message string) error {
	if e == nil || e.Publisher == nil {
		return nil
	}
	resource := ""
	if action.Scope == plan.ScopeResource {
		resource = action.Target
	}
	_, err := e.Publisher.Publish(events.ApplyProgress(workspace, resource, string(action.Scope), action.Target, action.RuntimeName, string(action.Kind), status, message))
	return err
}

func applyResource(resource *ResourcePayload) runtimepkg.AppliedResource {
	return runtimepkg.AppliedResource{
		Key:         resource.Key,
		LogicalHost: resource.LogicalHost,
		RuntimeName: resource.RuntimeName,
		Spec: runtimepkg.ResourceSpec{
			Image:         resource.Image,
			Build:         runtimeBuild(resource.Build),
			Command:       cloneStringSlice(resource.Command),
			Entrypoint:    cloneStringSlice(resource.Entrypoint),
			WorkingDir:    resource.WorkingDir,
			Env:           cloneEnvMap(resource.Env),
			Ports:         runtimePorts(resource.Ports),
			Volumes:       runtimeVolumes(resource.Volumes),
			Health:        cloneHealth(resource.Health),
			ProjectSource: cloneProjectSource(resource.ProjectSource),
			DevelopWatch:  cloneWatchRules(resource.DevelopWatch),
			Labels:        cloneStringMap(resource.Labels),
		},
	}
}

func runtimeBuild(build *BuildPayload) *runtimepkg.BuildSpec {
	if build == nil {
		return nil
	}
	return &runtimepkg.BuildSpec{Context: build.Context, Dockerfile: build.Dockerfile, Target: build.Target, Args: cloneEnvMap(build.Args)}
}

func runtimePorts(values []PortPayload) []runtimepkg.PortSpec {
	if len(values) == 0 {
		return nil
	}
	ports := make([]runtimepkg.PortSpec, len(values))
	for i := range values {
		ports[i] = runtimepkg.PortSpec(values[i])
	}
	return ports
}

func runtimeVolumes(values []VolumePayload) []runtimepkg.VolumeSpec {
	if len(values) == 0 {
		return nil
	}
	volumes := make([]runtimepkg.VolumeSpec, len(values))
	for i := range values {
		volumes[i] = runtimepkg.VolumeSpec(values[i])
	}
	return volumes
}

func cacheOperations(values []Operation) []cachepkg.OperationRecord {
	if len(values) == 0 {
		return nil
	}
	operations := make([]cachepkg.OperationRecord, len(values))
	for i := range values {
		operations[i] = cachepkg.OperationRecord{Scope: string(values[i].Scope), Target: values[i].Target, RuntimeName: values[i].RuntimeName, Kind: string(values[i].Kind), Status: values[i].Status, Message: values[i].Message}
	}
	return operations
}

func desiredBoundaryFromPayload(payload *Payload) *runtimepkg.DesiredWorkspace {
	desired := &runtimepkg.DesiredWorkspace{Name: payload.Workspace, Provider: payload.Provider, Resources: make([]*runtimepkg.DesiredResource, 0, len(payload.Resources))}
	if payload.Network != nil {
		desired.Network = &runtimepkg.DesiredNetwork{Name: payload.Network.Name, Labels: cloneStringMap(payload.Network.Labels)}
	}
	for _, resource := range payload.Resources {
		if resource == nil {
			continue
		}
		desired.Resources = append(desired.Resources, &runtimepkg.DesiredResource{Key: resource.Key, RuntimeName: resource.RuntimeName})
	}
	return desired
}

func networkName(payload *Payload) string {
	if payload == nil || payload.Network == nil {
		return ""
	}
	return payload.Network.Name
}
