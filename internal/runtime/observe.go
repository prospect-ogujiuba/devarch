package runtime

import (
	"context"

	"github.com/prospect-ogujiuba/devarch/internal/events"
)

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
