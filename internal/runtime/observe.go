package runtime

import (
	"context"

	"github.com/prospect-ogujiuba/devarch/internal/events"
)

func StreamLogsWithEvents(ctx context.Context, adapter Adapter, publisher events.Publisher, resource ResourceRef, request LogsRequest) error {
	if publisher != nil {
		if _, err := publisher.Publish(events.LogsStarted(resource.Workspace, resource.Key, request.Tail, request.Follow)); err != nil {
			return err
		}
	}
	err := adapter.StreamLogs(ctx, resource, request, func(chunk LogChunk) error {
		if publisher != nil {
			_, err := publisher.Publish(events.LogsChunk(resource.Workspace, resource.Key, chunk.Stream, chunk.Line, chunk.Timestamp))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if publisher != nil {
		_, err = publisher.Publish(events.LogsCompleted(resource.Workspace, resource.Key, request.Tail, request.Follow))
		if err != nil {
			return err
		}
	}
	return nil
}

func ExecWithEvents(ctx context.Context, adapter Adapter, publisher events.Publisher, resource ResourceRef, request ExecRequest) (*ExecResult, error) {
	if publisher != nil {
		if _, err := publisher.Publish(events.ExecStarted(resource.Workspace, resource.Key, request.Command)); err != nil {
			return nil, err
		}
	}
	result, err := adapter.Exec(ctx, resource, request)
	if err != nil {
		return nil, err
	}
	if publisher != nil {
		if _, err := publisher.Publish(events.ExecCompleted(resource.Workspace, resource.Key, result.ExitCode)); err != nil {
			return nil, err
		}
	}
	return result, nil
}
