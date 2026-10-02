package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	v, err := LoadVersions(dir)
	if err != nil || len(v) != 0 {
		t.Fatalf("missing file: %v %v", v, err)
	}
	if err := SaveVersions(dir, Versions{"PHP_VERSION": "8.3", "MARIADB_VERSION": "11.8"}); err != nil {
		t.Fatal(err)
	}
	v, err = LoadVersions(dir)
	if err != nil || v["PHP_VERSION"] != "8.3" || v["MARIADB_VERSION"] != "11.8" {
		t.Fatalf("got %v %v", v, err)
	}
}

func TestVersionsRejectsGarbage(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "versions.env"), []byte("export X=1\n"), 0o644)
	if _, err := LoadVersions(dir); err == nil {
		t.Fatal("accepted malformed line")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	if err := SaveConfig(dir, Config{Root: "/x"}); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(dir)
	if err != nil || c.Root != "/x" {
		t.Fatalf("got %+v %v", c, err)
	}
}

func TestDirPrecedence(t *testing.T) {
	t.Setenv("DEVARCH_CONFIG_HOME", "/a")
	if d, _ := Dir(); d != "/a" {
		t.Fatal(d)
	}
	t.Setenv("DEVARCH_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "/b")
	if d, _ := Dir(); d != "/b/devarch" {
		t.Fatal(d)
	}
}
