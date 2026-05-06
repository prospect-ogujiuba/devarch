package workflows

import "context"

type FakeRunner struct {
	Results []CommandResult
	Calls   []CommandResult
}

func (f *FakeRunner) Run(ctx context.Context, command string, args ...string) CommandResult {
	_ = ctx
	call := CommandResult{Command: command, Args: append([]string(nil), args...)}
	f.Calls = append(f.Calls, call)
	if len(f.Results) == 0 {
		return CommandResult{Command: command, Args: append([]string(nil), args...), Status: StatusFail, ExitCode: -1, Error: "fake result missing"}
	}
	result := f.Results[0]
	f.Results = f.Results[1:]
	if result.Command == "" {
		result.Command = command
	}
	if result.Args == nil {
		result.Args = append([]string(nil), args...)
	}
	return result
}
