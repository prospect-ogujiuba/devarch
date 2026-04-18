package events

import (
	"encoding/json"
	"time"
)

type Kind string

const (
	KindApplyStarted   Kind = "apply.started"
	KindApplyProgress  Kind = "apply.progress"
	KindApplyCompleted Kind = "apply.completed"
	KindLogsStarted    Kind = "logs.started"
	KindLogsChunk      Kind = "logs.chunk"
	KindLogsCompleted  Kind = "logs.completed"
	KindExecStarted    Kind = "exec.started"
	KindExecCompleted  Kind = "exec.completed"
)

type Envelope struct {
	Sequence  uint64          `json:"sequence"`
	Workspace string          `json:"workspace"`
	Resource  string          `json:"resource,omitempty"`
	Kind      Kind            `json:"kind"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type Spec struct {
	Workspace string
	Resource  string
	Kind      Kind
	Timestamp time.Time
	Payload   any
}

type Publisher interface {
	Publish(spec Spec) (Envelope, error)
}

type ApplyStartedPayload struct {
	TotalActions int `json:"totalActions"`
}

type ApplyProgressPayload struct {
	Scope       string `json:"scope"`
	Target      string `json:"target"`
	RuntimeName string `json:"runtimeName,omitempty"`
	Action      string `json:"action"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
}

type ApplyCompletedPayload struct {
	Succeeded      bool `json:"succeeded"`
	OperationCount int  `json:"operationCount"`
}

type LogsSessionPayload struct {
	Tail   int  `json:"tail,omitempty"`
	Follow bool `json:"follow,omitempty"`
}

type LogsChunkPayload struct {
	Stream    string     `json:"stream,omitempty"`
	Line      string     `json:"line"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type ExecStartedPayload struct {
	Command []string `json:"command"`
}

type ExecCompletedPayload struct {
	ExitCode int `json:"exitCode"`
}

func ApplyStarted(workspace string, totalActions int) Spec {
	return Spec{Workspace: workspace, Kind: KindApplyStarted, Payload: ApplyStartedPayload{TotalActions: totalActions}}
}

func ApplyProgress(workspace, resource, scope, target, runtimeName, action, status, message string) Spec {
	return Spec{
		Workspace: workspace,
		Resource:  resource,
		Kind:      KindApplyProgress,
		Payload: ApplyProgressPayload{
			Scope:       scope,
			Target:      target,
			RuntimeName: runtimeName,
			Action:      action,
			Status:      status,
			Message:     message,
		},
	}
}

func ApplyCompleted(workspace string, succeeded bool, operationCount int) Spec {
	return Spec{Workspace: workspace, Kind: KindApplyCompleted, Payload: ApplyCompletedPayload{Succeeded: succeeded, OperationCount: operationCount}}
}

func LogsStarted(workspace, resource string, tail int, follow bool) Spec {
	return Spec{Workspace: workspace, Resource: resource, Kind: KindLogsStarted, Payload: LogsSessionPayload{Tail: tail, Follow: follow}}
}

func LogsChunk(workspace, resource, stream, line string, timestamp *time.Time) Spec {
	return Spec{Workspace: workspace, Resource: resource, Kind: KindLogsChunk, Payload: LogsChunkPayload{Stream: stream, Line: line, Timestamp: timestamp}}
}

func LogsCompleted(workspace, resource string, tail int, follow bool) Spec {
	return Spec{Workspace: workspace, Resource: resource, Kind: KindLogsCompleted, Payload: LogsSessionPayload{Tail: tail, Follow: follow}}
}

func ExecStarted(workspace, resource string, command []string) Spec {
	return Spec{Workspace: workspace, Resource: resource, Kind: KindExecStarted, Payload: ExecStartedPayload{Command: append([]string(nil), command...)}}
}

func ExecCompleted(workspace, resource string, exitCode int) Spec {
	return Spec{Workspace: workspace, Resource: resource, Kind: KindExecCompleted, Payload: ExecCompletedPayload{ExitCode: exitCode}}
}
