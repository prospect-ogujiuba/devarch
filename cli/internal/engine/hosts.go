package engine

import (
	"context"
	"sort"
	"strings"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
)

// HostLabels returns the .test labels a service answers on: its container
// names plus .test hosts in its x-devarch urls.
func HostLabels(svc catalog.Service) []string {
	set := map[string]bool{}
	for _, c := range svc.Containers {
		if c.ContainerName != "" && hosts.ValidateHostname(c.ContainerName+".test") == nil {
			set[c.ContainerName] = true
		}
	}
	for _, u := range svc.Meta.URLs {
		if l, ok := hosts.TestLabel(u); ok {
			set[l] = true
		}
	}
	out := make([]string, 0, len(set))
	for l := range set {
		out = append(out, l)
	}
	sort.Strings(out)
	return out
}

// HostsBlock discovers every DevArch domain and renders the managed block.
func (e *Engine) HostsBlock(address string) (string, hosts.Domains, error) {
	d, err := hosts.Discover(e.Catalog, e.Settings.AppsDir)
	if err != nil {
		return "", d, err
	}
	return hosts.Block(d, address), d, nil
}

// EnsureHosts syncs the managed block when any of svcs' hostnames is unmapped.
// Failures are reported as warnings: the services are already running.
// Nothing happens when config.yml sets hosts.manage: false.
func (e *Engine) EnsureHosts(ctx context.Context, svcs []catalog.Service) {
	if e.Hosts == nil || !e.Settings.HostsManage {
		return
	}
	content, err := e.Hosts.Read()
	if err != nil {
		e.logf("warning: cannot read %s: %v", e.Hosts.Path(), err)
		return
	}
	mapped := hosts.Mapped(content)
	var missing []string
	for _, s := range svcs {
		for _, l := range HostLabels(s) {
			if _, ok := mapped[l+".test"]; !ok {
				missing = append(missing, l+".test")
			}
		}
	}
	if len(missing) == 0 {
		return
	}
	block, _, err := e.HostsBlock(e.Settings.HostsAddress)
	if err != nil {
		e.logf("warning: hosts not updated: %v", err)
		return
	}
	if e.DryRun {
		e.logf("would sync the DevArch hosts block in %s (missing %s)", e.Hosts.Path(), strings.Join(missing, ", "))
		return
	}
	e.logf("registering %s in %s", strings.Join(missing, ", "), e.Hosts.Path())
	if _, err := e.Hosts.Sync(ctx, block); err != nil {
		e.logf("warning: hosts not updated: %v (run `devarch hosts sync` to retry)", err)
	}
}
