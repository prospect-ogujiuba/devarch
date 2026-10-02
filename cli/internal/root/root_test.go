package root

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	deep := filepath.Join(repo, "apps", "site", "src")
	os.MkdirAll(filepath.Join(repo, "services-library"), 0o755)
	os.MkdirAll(deep, 0o755)
	repo, _ = filepath.EvalSymlinks(repo)

	t.Setenv("DEVARCH_ROOT", "")
	if got, how, err := Find("", deep); err != nil || got != repo || how != FromWalk {
		t.Fatalf("walk: %s %s %v", got, how, err)
	}
	if got, how, err := Find(repo, base); err != nil || got != repo || how != FromConfig {
		t.Fatalf("config: %s %s %v", got, how, err)
	}
	if _, _, err := Find(base, base); err == nil {
		t.Fatal("accepted configured non-root")
	}
	if _, _, err := Find("", base); err == nil {
		t.Fatal("walk outside a checkout succeeded")
	}
	t.Setenv("DEVARCH_ROOT", repo)
	if _, how, err := Find(base, base); err != nil || how != FromEnv {
		t.Fatalf("env: %s %v", how, err)
	}
}
