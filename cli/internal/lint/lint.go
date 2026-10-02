// Package lint validates the catalog contract: compose structure, x-devarch
// metadata, container names, host-port collisions and image pinning.
package lint

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

// Finding is one lint result.
type Finding struct {
	Severity Severity `json:"severity"`
	Service  string   `json:"service"`
	Message  string   `json:"message"`
}

var (
	varRE   = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	labelRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	tagRE   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// Options controls Run.
type Options struct {
	// Native also runs `podman compose config --quiet` for every service.
	Native bool
	Runner runner.Runner
}

// Run lints the catalog. Findings are sorted by service then message.
func Run(ctx context.Context, cat *catalog.Catalog, problems []catalog.Problem, opts Options) []Finding {
	var out []Finding
	add := func(sev Severity, svc, format string, args ...any) {
		out = append(out, Finding{sev, svc, fmt.Sprintf(format, args...)})
	}
	for _, p := range problems {
		add(Error, p.Path, "%v", p.Err)
	}

	type owner struct{ svc, ip string }
	ports := map[string][]owner{} // "port/proto" -> owners
	names := map[string][]string{}

	for _, s := range cat.Services {
		lintMeta(cat, s, add)
		for _, c := range s.Containers {
			if c.ContainerName != "" {
				if !labelRE.MatchString(c.ContainerName) {
					add(Error, s.ID, "container_name %q is not a valid DNS label (hosts sync rejects it)", c.ContainerName)
				}
				names[c.ContainerName] = append(names[c.ContainerName], s.ID)
			}
			if c.Image != "" && !c.Build && unpinned(c.Image) {
				add(Warning, s.ID, "image %s is not pinned to a version", c.Image)
			}
			for _, p := range c.Ports {
				if p.Published == 0 {
					continue
				}
				key := fmt.Sprintf("%d/%s", p.Published, p.Protocol)
				for _, o := range ports[key] {
					if o.svc != s.ID && (o.ip == p.HostIP || o.ip == "" || p.HostIP == "" || o.ip == "0.0.0.0" || p.HostIP == "0.0.0.0") {
						add(Error, s.ID, "host port %s is also published by %s", key, o.svc)
					}
				}
				ports[key] = append(ports[key], owner{s.ID, p.HostIP})
			}
		}
		if opts.Native && opts.Runner != nil {
			cmd := runner.Cmd{Name: "podman", Args: []string{"compose", "config", "--quiet"}, Dir: s.Dir}
			if _, err := opts.Runner.Output(ctx, cmd); err != nil {
				add(Error, s.ID, "podman compose config failed: %v", err)
			}
		}
	}
	for name, ids := range names {
		if len(ids) > 1 {
			add(Warning, strings.Join(ids, ", "), "container_name %q is used by more than one service; they cannot run together", name)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Service != out[j].Service {
			return out[i].Service < out[j].Service
		}
		return out[i].Message < out[j].Message
	})
	return slices.Compact(out)
}

func lintMeta(cat *catalog.Catalog, s catalog.Service, add func(Severity, string, string, ...any)) {
	m := s.Meta
	if v := m.Version; v != nil {
		switch {
		case !varRE.MatchString(v.Var):
			add(Error, s.ID, "version.var %q must be an upper-case environment variable name", v.Var)
		case v.Default == "":
			add(Error, s.ID, "version.default is required")
		default:
			data, err := os.ReadFile(s.ComposeFile)
			want := "${" + v.Var + ":-" + v.Default + "}"
			if err == nil && !strings.Contains(string(data), want) {
				add(Error, s.ID, "compose.yml must reference %s so plain podman compose uses the default", want)
			}
		}
		if len(v.Choices) > 0 && !slices.Contains(v.Choices, v.Default) {
			add(Error, s.ID, "version.default %q is not in version.choices", v.Default)
		}
		switch v.Data {
		case "", "upgrade-only", "same-major":
		default:
			add(Error, s.ID, "version.data must be upgrade-only or same-major, not %q", v.Data)
		}
	}
	for _, r := range m.Requires {
		if _, err := cat.Resolve(r); err != nil {
			add(Error, s.ID, "requires %s: %v", r, err)
		}
	}
	if _, err := cat.WithRequires([]catalog.Service{s}); err != nil && strings.Contains(err.Error(), "cycle") {
		add(Error, s.ID, "%v", err)
	}
	for _, u := range m.URLs {
		parsed, err := url.Parse(u)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			add(Error, s.ID, "url %q must be an absolute http(s) URL", u)
		}
	}
	if m.Ready != nil && len(m.Ready.Exec) == 0 {
		add(Error, s.ID, "ready.exec must not be empty")
	}
	for _, t := range m.Tags {
		if !tagRE.MatchString(t) {
			add(Warning, s.ID, "tag %q should be lower-case kebab-case", t)
		}
	}
}

func unpinned(image string) bool {
	if strings.Contains(image, "${") || strings.Contains(image, "@sha256:") {
		return false
	}
	name := image[strings.LastIndexByte(image, '/')+1:]
	i := strings.IndexByte(name, ':')
	return i < 0 || name[i+1:] == "latest"
}

// Count returns the number of errors and warnings.
func Count(fs []Finding) (errors, warnings int) {
	for _, f := range fs {
		if f.Severity == Error {
			errors++
		} else {
			warnings++
		}
	}
	return
}
