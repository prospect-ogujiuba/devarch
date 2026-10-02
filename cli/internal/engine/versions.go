package engine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

// UseOptions controls Use.
type UseOptions struct {
	// Force allows versions outside choices and data-incompatible switches.
	Force bool
	// NoRecreate records the version without rebuilding or recreating.
	NoRecreate bool
}

// UseResult reports what Use did.
type UseResult struct {
	From, To   string
	Rebuilt    bool
	Recreated  bool
	WasRunning bool
}

// DataConflictError explains why a switch could make existing data unreadable.
type DataConflictError struct {
	Service, From, To, Rule string
	Volumes                 []string
}

func (e *DataConflictError) Error() string {
	why := "its data volume cannot be downgraded"
	if e.Rule == "same-major" {
		why = "its data directory is tied to the major version"
	}
	return fmt.Sprintf("refusing to switch %s from %s to %s: %s (%s).\n"+
		"Dump and restore the data, or delete it with `devarch down %s --volumes`, then retry; --force skips this check",
		e.Service, e.From, e.To, why, strings.Join(e.Volumes, ", "), e.Service)
}

// Use selects a version for svc, records it in versions.env, rebuilds
// Dockerfile-based services, and recreates the service if it is running.
func (e *Engine) Use(ctx context.Context, svc catalog.Service, version string, opts UseOptions) (UseResult, error) {
	v := svc.Meta.Version
	if v == nil || v.Var == "" {
		return UseResult{}, fmt.Errorf("%s has no switchable version (add x-devarch.version to its compose.yml)", svc.ID)
	}
	res := UseResult{From: e.SelectedVersion(svc), To: version}
	if !opts.Force && len(v.Choices) > 0 && !slices.Contains(v.Choices, version) {
		return res, fmt.Errorf("%s is not a listed %s version (choices: %s); --force uses it anyway",
			version, svc.ID, strings.Join(v.Choices, ", "))
	}
	if !opts.Force && version != res.From {
		if err := e.checkData(ctx, svc, res.From, version); err != nil {
			return res, err
		}
	}

	if e.Versions == nil {
		e.Versions = map[string]string{}
	}
	if version == v.Default {
		delete(e.Versions, v.Var)
	} else {
		e.Versions[v.Var] = version
	}
	if e.DryRun {
		e.logf("would record %s=%s in versions.env", v.Var, version)
	} else if e.SaveVersions != nil {
		if err := e.SaveVersions(e.Versions); err != nil {
			return res, err
		}
	}
	if opts.NoRecreate {
		return res, nil
	}

	cs, err := e.Containers(ctx)
	if err == nil {
		st := States([]catalog.Service{svc}, cs)[svc.ID].State
		res.WasRunning = st == "running" || st == "partial"
	}
	if v.Rebuild {
		e.logf("building %s %s", svc.ID, version)
		if err := e.Runner.Run(ctx, e.Compose(svc, "build")); err != nil {
			return res, fmt.Errorf("build %s %s: %w", svc.ID, version, err)
		}
		res.Rebuilt = true
	}
	if res.WasRunning {
		e.logf("recreating %s", svc.ID)
		if err := e.Runner.Run(ctx, e.Compose(svc, "up", "-d", "--force-recreate")); err != nil {
			return res, fmt.Errorf("recreate %s: %w", svc.ID, err)
		}
		res.Recreated = true
	}
	return res, nil
}

func (e *Engine) checkData(ctx context.Context, svc catalog.Service, from, to string) error {
	rule := svc.Meta.Version.Data
	switch rule {
	case "upgrade-only":
		if CompareVersions(to, from) >= 0 {
			return nil
		}
	case "same-major":
		if Major(to) == Major(from) {
			return nil
		}
	default:
		return nil
	}
	var existing []string
	for _, vol := range svc.Volumes {
		_, err := e.Runner.Output(ctx, e.Cmd("volume", "exists", vol))
		var exit *runner.ExitError
		switch {
		case err == nil:
			existing = append(existing, vol)
		case !errors.As(err, &exit):
			return fmt.Errorf("cannot check volume %s: %w", vol, err)
		}
	}
	if len(existing) == 0 {
		return nil
	}
	return &DataConflictError{Service: svc.Name, From: from, To: to, Rule: rule, Volumes: existing}
}

// Major returns the leading numeric component of a version ("v1.22" -> "1").
func Major(v string) string {
	parts := versionParts(v)
	if len(parts) == 0 {
		return v
	}
	return strconv.Itoa(parts[0])
}

// CompareVersions compares dotted numeric versions, ignoring a leading "v",
// over their common components only: a floating tag such as "12.3" equals
// "12.3.2" because it resolves to the newest 12.3.x.
func CompareVersions(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	var out []int
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		out = append(out, n)
	}
	return out
}
