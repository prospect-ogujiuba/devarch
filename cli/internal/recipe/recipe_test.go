package recipe

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func fixture(t *testing.T) (root string, rs []Recipe) {
	t.Helper()
	root = t.TempDir()
	testutil.WriteFiles(t, root, map[string]string{
		"recipes/js/recipe.yml": `name: js
title: JS
entry: scripts/js.sh
args:
  - {name: name, positional: true, pattern: '^[a-z]+$'}
  - {name: framework, flag: --framework, choices_from: profiles/*/framework.conf, choice_is: dir}
  - {name: profile, flag: --profile, choices_from: 'profiles/{framework}/*.profile'}
  - {name: start, flag: --start, type: bool}
`,
		"profiles/next/framework.conf":    "",
		"profiles/next/api.profile":       "",
		"profiles/next/fullstack.profile": "",
		"profiles/astro/framework.conf":   "",
		"profiles/astro/blog.profile":     "",
		"recipes/notarecipe/README.md":    "",
		"scripts/js.sh": `#!/usr/bin/env bash
printf '{"step":"scaffold","state":"start"}\n' >&3
echo "args: $*"
printf 'not json\n' >&3
printf '{"step":"scaffold","state":"done","message":"made %s"}\n' "$1" >&3
`,
	})
	os.Chmod(filepath.Join(root, "scripts/js.sh"), 0o755)
	user := filepath.Join(t.TempDir(), "recipes")
	testutil.WriteFiles(t, user, map[string]string{"js/recipe.yml": "name: js\ntitle: My JS\nentry: scripts/js.sh\n"})
	var err error
	rs, err = Load(root, filepath.Join(root, "recipes"))
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := Load(root, user, filepath.Join(root, "recipes"))
	if err != nil || len(overlay) != 1 || overlay[0].Title != "My JS" {
		t.Fatalf("user overlay: %+v %v", overlay, err)
	}
	return root, rs
}

func TestChoicesAndValidate(t *testing.T) {
	_, rs := fixture(t)
	r, err := Find(rs, "js")
	if err != nil {
		t.Fatal(err)
	}
	fw := r.Args[1]
	if got := strings.Join(r.ChoicesFor(fw, nil), ","); got != "astro,next" {
		t.Fatalf("framework choices = %s", got)
	}
	if got := r.ChoicesFor(r.Args[2], map[string]string{}); got != nil {
		t.Fatalf("profile choices without a framework = %v", got)
	}
	if got := strings.Join(r.ChoicesFor(r.Args[2], map[string]string{"framework": "next"}), ","); got != "api,fullstack" {
		t.Fatalf("profile choices = %s", got)
	}
	values := r.Parse([]string{"store", "--framework=next", "--profile", "api", "--start", "--unknown", "x"})
	if values["name"] != "store" || values["framework"] != "next" || values["profile"] != "api" || values["start"] != "true" {
		t.Fatalf("parse = %v", values)
	}
	if err := r.Validate(values); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []map[string]string{{"name": "Store"}, {"framework": "rails"}, {"framework": "astro", "profile": "api"}} {
		if err := r.Validate(bad); err == nil {
			t.Errorf("accepted %v", bad)
		}
	}
	if got := strings.Join(r.Argv(values), " "); got != "store --framework next --profile api --start" {
		t.Fatalf("argv = %s", got)
	}
	if _, err := Find(rs, "nope"); err == nil || !strings.Contains(err.Error(), "available: js") {
		t.Fatalf("find error = %v", err)
	}
}

func TestRunWithProgress(t *testing.T) {
	_, rs := fixture(t)
	var out bytes.Buffer
	var events []Event
	err := rs[0].RunWithProgress(context.Background(), []string{"store"}, nil, &out, func(e Event) { events = append(events, e) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "args: store") {
		t.Fatalf("output = %q", out.String())
	}
	if len(events) != 2 || events[1].State != "done" || events[1].Message != "made store" {
		t.Fatalf("events = %+v", events)
	}
}

func TestLoadRejectsBadManifests(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFiles(t, root, map[string]string{"recipes/x/recipe.yml": "name: x\n"})
	if _, err := Load(root, filepath.Join(root, "recipes")); err == nil {
		t.Fatal("missing entry accepted")
	}
	testutil.WriteFiles(t, root, map[string]string{"recipes/x/recipe.yml": "name: x\nentry: a\nargs: [{name: n, pattern: '('}]\n"})
	if _, err := Load(root, filepath.Join(root, "recipes")); err == nil {
		t.Fatal("invalid pattern accepted")
	}
}
