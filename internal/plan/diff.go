package plan

import (
	"fmt"
	"reflect"
	"sort"

	runtimepkg "github.com/prospect-ogujiuba/devarch/internal/runtime"
)

func Diff(desired *runtimepkg.DesiredWorkspace, snapshot *runtimepkg.Snapshot) (*Result, error) {
	if desired == nil {
		return nil, fmt.Errorf("plan diff: nil desired workspace")
	}
	if snapshot == nil {
		snapshot = &runtimepkg.Snapshot{Workspace: runtimepkg.SnapshotWorkspace{Name: desired.Name, Provider: desired.Provider}}
	}

	result := &Result{
		Workspace:   desired.Name,
		Provider:    desired.Provider,
		Blocked:     desired.Blocked(),
		Diagnostics: append([]runtimepkg.Diagnostic(nil), desired.Diagnostics...),
	}
	for _, resource := range desired.Resources {
		if resource != nil && len(resource.Diagnostics) > 0 {
			result.Diagnostics = append(result.Diagnostics, resource.Diagnostics...)
		}
	}
	if result.Blocked {
		return result, nil
	}

	actions := make([]Action, 0, len(desired.Resources)+1)
	if desired.Network != nil || snapshot.Workspace.Network != nil {
		actions = append(actions, diffWorkspaceNetwork(desired, snapshot))
	}

	desiredByKey := make(map[string]*runtimepkg.DesiredResource, len(desired.Resources))
	for _, resource := range desired.Resources {
		if resource == nil {
			continue
		}
		desiredByKey[resource.Key] = resource
		actions = append(actions, diffResource(resource, snapshot.Resource(resource.Key)))
	}

	for _, resource := range snapshot.Resources {
		if resource == nil {
			continue
		}
		if _, ok := desiredByKey[resource.Key]; ok {
			continue
		}
		actions = append(actions, Action{
			Scope:       ScopeResource,
			Target:      resource.Key,
			RuntimeName: resource.RuntimeName,
			Kind:        ActionRemove,
			Reasons:     resourceRemoveReasons(false),
		})
	}

	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Scope != actions[j].Scope {
			if actions[i].Scope == ScopeWorkspace {
				return true
			}
			if actions[j].Scope == ScopeWorkspace {
				return false
			}
			return actions[i].Scope < actions[j].Scope
		}
		if actions[i].Target != actions[j].Target {
			return actions[i].Target < actions[j].Target
		}
		if actions[i].Kind != actions[j].Kind {
			return actions[i].Kind < actions[j].Kind
		}
		return actions[i].RuntimeName < actions[j].RuntimeName
	})
	result.Actions = actions
	return result, nil
}

func diffWorkspaceNetwork(desired *runtimepkg.DesiredWorkspace, snapshot *runtimepkg.Snapshot) Action {
	if desired.Network != nil && snapshot.Workspace.Network == nil {
		return Action{Scope: ScopeWorkspace, Target: "network", RuntimeName: desired.Network.Name, Kind: ActionAdd, Reasons: workspaceNetworkAddReasons()}
	}
	if desired.Network == nil && snapshot.Workspace.Network != nil {
		return Action{Scope: ScopeWorkspace, Target: "network", RuntimeName: snapshot.Workspace.Network.Name, Kind: ActionRemove, Reasons: workspaceNetworkRemoveReasons()}
	}
	if desired.Network != nil && snapshot.Workspace.Network != nil && desired.Network.Name != snapshot.Workspace.Network.Name {
		return Action{Scope: ScopeWorkspace, Target: "network", RuntimeName: desired.Network.Name, Kind: ActionModify, Reasons: []string{"workspace network identity changed"}}
	}
	name := ""
	if desired.Network != nil {
		name = desired.Network.Name
	} else if snapshot.Workspace.Network != nil {
		name = snapshot.Workspace.Network.Name
	}
	return Action{Scope: ScopeWorkspace, Target: "network", RuntimeName: name, Kind: ActionNoop, Reasons: workspaceNetworkNoopReasons()}
}

func diffResource(desired *runtimepkg.DesiredResource, snapshot *runtimepkg.SnapshotResource) Action {
	if desired == nil {
		return Action{}
	}
	if !desired.Enabled {
		if snapshot == nil {
			return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: desired.RuntimeName, Kind: ActionNoop, Reasons: resourceDisabledNoopReasons()}
		}
		return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: snapshot.RuntimeName, Kind: ActionRemove, Reasons: resourceRemoveReasons(true)}
	}
	if snapshot == nil {
		return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: desired.RuntimeName, Kind: ActionAdd, Reasons: resourceAddReasons()}
	}

	fields := changedFields(desired, snapshot)
	if len(fields) > 0 {
		return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: desired.RuntimeName, Kind: ActionModify, Reasons: modifyReasons(fields)}
	}
	if requiresRestart(snapshot) {
		return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: desired.RuntimeName, Kind: ActionRestart, Reasons: resourceRestartReasons(snapshot.State.Running, snapshot.State.Health)}
	}
	return Action{Scope: ScopeResource, Target: desired.Key, RuntimeName: desired.RuntimeName, Kind: ActionNoop, Reasons: resourceNoopReasons()}
}

func changedFields(desired *runtimepkg.DesiredResource, snapshot *runtimepkg.SnapshotResource) []string {
	fields := make([]string, 0)
	if desired == nil || snapshot == nil {
		return fields
	}
	if desired.Spec.Image != snapshot.Spec.Image {
		fields = append(fields, "image")
	}
	if !reflect.DeepEqual(desired.Spec.Build, snapshot.Spec.Build) {
		fields = append(fields, "build")
	}
	if !reflect.DeepEqual(desired.Spec.Command, snapshot.Spec.Command) {
		fields = append(fields, "command")
	}
	if !reflect.DeepEqual(desired.Spec.Entrypoint, snapshot.Spec.Entrypoint) {
		fields = append(fields, "entrypoint")
	}
	if desired.Spec.WorkingDir != snapshot.Spec.WorkingDir {
		fields = append(fields, "workingDir")
	}
	if !reflect.DeepEqual(desired.Spec.Env, snapshot.Spec.Env) {
		fields = append(fields, "env")
	}
	if !reflect.DeepEqual(desired.Spec.Ports, snapshot.Spec.Ports) {
		fields = append(fields, "ports")
	}
	if !reflect.DeepEqual(desired.Spec.Volumes, snapshot.Spec.Volumes) {
		fields = append(fields, "volumes")
	}
	if !reflect.DeepEqual(desired.Spec.Health, snapshot.Spec.Health) {
		fields = append(fields, "health")
	}
	if !reflect.DeepEqual(desired.Spec.ProjectSource, snapshot.Spec.ProjectSource) {
		fields = append(fields, "projectSource")
	}
	if !reflect.DeepEqual(desired.Spec.DevelopWatch, snapshot.Spec.DevelopWatch) {
		fields = append(fields, "developWatch")
	}
	if !reflect.DeepEqual(desired.Spec.Labels, snapshot.Spec.Labels) {
		fields = append(fields, "labels")
	}
	sort.Strings(fields)
	return fields
}

func requiresRestart(snapshot *runtimepkg.SnapshotResource) bool {
	if snapshot == nil {
		return false
	}
	if !snapshot.State.Running {
		return true
	}
	return snapshot.State.Health == "unhealthy"
}
