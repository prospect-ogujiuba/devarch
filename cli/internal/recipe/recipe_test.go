package recipe

import (
	"bytes"
	"context"
	"fmt"
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

func TestPreviewChoice(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		os.MkdirAll(filepath.Dir(filepath.Join(root, rel)), 0o755)
		os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644)
	}
	write("profiles/base.profile", "# Base set.\nwp-plugin query-monitor\n")
	write("profiles/full.profile", "# Everything.\ninclude base.profile\ngithub-plugin migrate inactive   # trailing note\n\nwp-plugin debug-bar\n")
	write("profiles/bad.profile", "teleport now\n")
	write("profiles/loop.profile", "include loop.profile\n")
	arg := Arg{Name: "profile", ChoicesFrom: "profiles/*.profile", Preview: &Preview{Include: "include", Directives: []Directive{
		{Kind: "github-plugin", Label: "GitHub plugins"}, {Kind: "wp-plugin", Label: "WordPress.org plugins"}}}}
	r := Recipe{Root: root}
	p, err := r.PreviewChoice(arg, nil, "full")
	if err != nil {
		t.Fatal(err)
	}
	want := ProfilePreview{Description: "Everything.", Groups: []Group{
		{Label: "GitHub plugins", Items: []string{"migrate inactive"}},
		{Label: "WordPress.org plugins", Items: []string{"query-monitor", "debug-bar"}},
	}}
	if fmt.Sprint(p) != fmt.Sprint(want) {
		t.Fatalf("got %+v", p)
	}
	for _, choice := range []string{"bad", "loop", "missing", "../etc/passwd"} {
		if _, err := r.PreviewChoice(arg, nil, choice); err == nil {
			t.Errorf("%s: no error", choice)
		}
	}
}

func TestRepositoryProfilesPreview(t *testing.T) {
	root, _ := filepath.Abs("../../..")
	rs, err := Load(root, filepath.Join(root, "recipes"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, r := range rs {
		for _, a := range r.Args {
			if a.Preview == nil {
				continue
			}
			for _, c := range r.ChoicesFor(a, map[string]string{}) {
				if _, err := r.PreviewChoice(a, nil, c); err != nil {
					t.Errorf("%s %s: %v", r.Name, c, err)
				}
				checked++
			}
		}
	}
	if checked < 7 {
		t.Fatalf("only %d profiles previewed", checked)
	}
}
