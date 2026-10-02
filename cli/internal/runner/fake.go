package runner

import (
	"context"
	"strings"
	"sync"
)

// Fake records commands and answers queries from scripted responses. It is
// exported for tests in other packages.
type Fake struct {
	mu   sync.Mutex
	Cmds []Cmd
	// Responses maps a command-line prefix (Cmd.String without Dir) to output.
	Responses map[string]Response
	// RunErrors maps a command-line prefix to errors returned by Run, consumed in order.
	RunErrors map[string][]error
}

type Response struct {
	Out []byte
	Err error
}

func plain(c Cmd) string {
	c.Dir = ""
	c.Env = nil
	return c.String()
}

func (f *Fake) record(c Cmd) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Cmds = append(f.Cmds, c)
}

func (f *Fake) Run(_ context.Context, c Cmd) error {
	f.record(c)
	f.mu.Lock()
	defer f.mu.Unlock()
	line := plain(c)
	for prefix, errs := range f.RunErrors {
		if strings.HasPrefix(line, prefix) && len(errs) > 0 {
			f.RunErrors[prefix] = errs[1:]
			return errs[0]
		}
	}
	return nil
}

func (f *Fake) Output(_ context.Context, c Cmd) ([]byte, error) {
	f.record(c)
	line := plain(c)
	best := ""
	for prefix := range f.Responses {
		if strings.HasPrefix(line, prefix) && len(prefix) > len(best) {
			best = prefix
		}
	}
	if best == "" {
		return nil, nil
	}
	r := f.Responses[best]
	return r.Out, r.Err
}

func (f *Fake) Exec(c Cmd) error {
	f.record(c)
	return nil
}

// Lines returns every recorded command as a string, including Dir and Env.
func (f *Fake) Lines() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.Cmds))
	for i, c := range f.Cmds {
		out[i] = c.String()
	}
	return out
}
