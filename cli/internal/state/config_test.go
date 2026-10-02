package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDefaults(t *testing.T) {
	s, err := Config{}.Resolve("/repo")
	if err != nil {
		t.Fatal(err)
	}
	if s.Runtime != "podman" || s.AppsDir != "/repo/apps" || s.Network != "microservices-net" ||
		!s.HostsManage || s.HostsAddress != "127.0.0.1" || s.Editor != "code" {
		t.Fatalf("defaults: %+v", s)
	}
	for _, k := range Keys {
		if s.Sources[k] != "default" {
			t.Fatalf("%s source = %q", k, s.Sources[k])
		}
	}
}

func TestResolveConfigured(t *testing.T) {
	off := false
	c := Config{AppsDir: "sites", Network: "dev-net", Editor: "nvim", Hosts: HostsConfig{Manage: &off, Address: "127.0.0.2"}}
	s, err := c.Resolve("/repo")
	if err != nil {
		t.Fatal(err)
	}
	if s.AppsDir != "/repo/sites" || s.Network != "dev-net" || s.HostsManage || s.HostsAddress != "127.0.0.2" || s.Editor != "nvim" {
		t.Fatalf("got %+v", s)
	}
	if s.Sources["hosts.manage"] != "config.yml" || s.Sources["runtime"] != "default" {
		t.Fatalf("sources %v", s.Sources)
	}
	if s, _ := (Config{AppsDir: "/srv/apps"}).Resolve("/repo"); s.AppsDir != "/srv/apps" {
		t.Fatalf("absolute apps_dir: %s", s.AppsDir)
	}
}

func TestResolveRejectsInvalid(t *testing.T) {
	for _, c := range []Config{
		{Runtime: "lxc"},
		{Network: "bad name"},
		{Hosts: HostsConfig{Address: "::1"}},
		{Editor: "   "},
	} {
		if _, err := c.Resolve("/repo"); err == nil {
			t.Errorf("accepted %+v", c)
		}
	}
}

func TestLoadConfigRejectsUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yml"), []byte("runtme: docker\n"), 0o644)
	if _, err := LoadConfig(dir); err == nil || !strings.Contains(err.Error(), "runtme") {
		t.Fatalf("err = %v", err)
	}
	os.WriteFile(filepath.Join(dir, "config.yml"), nil, 0o644)
	if _, err := LoadConfig(dir); err != nil {
		t.Fatalf("empty file: %v", err)
	}
}

func TestSetConfigKeepsCommentsAndOtherKeys(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yml"), []byte("# mine\nroot: /repo\neditor: vim # terminal\n"), 0o644)
	if err := SetConfig(dir, "editor", "code -n"); err != nil {
		t.Fatal(err)
	}
	if err := SetConfig(dir, "hosts.manage", "no"); err == nil {
		t.Fatal("accepted a non-boolean hosts.manage")
	}
	if err := SetConfig(dir, "hosts.manage", "false"); err != nil {
		t.Fatal(err)
	}
	if err := SetConfig(dir, "hosts.address", "localhost"); err == nil {
		t.Fatal("accepted a hostname as hosts.address")
	}
	if err := SetConfig(dir, "root", "/elsewhere"); err == nil {
		t.Fatal("root is not settable through config set")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "config.yml"))
	for _, want := range []string{"# mine", "root: /repo", "editor: code -n # terminal", "hosts:\n  manage: false"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q in:\n%s", want, data)
		}
	}
	c, err := LoadConfig(dir)
	if err != nil || c.Root != "/repo" || c.Editor != "code -n" || c.Hosts.Manage == nil || *c.Hosts.Manage {
		t.Fatalf("reloaded %+v %v", c, err)
	}
}

func TestSetRootCreatesFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	if err := SetRoot(dir, "/x"); err != nil {
		t.Fatal(err)
	}
	if c, err := LoadConfig(dir); err != nil || c.Root != "/x" {
		t.Fatalf("got %+v %v", c, err)
	}
}
