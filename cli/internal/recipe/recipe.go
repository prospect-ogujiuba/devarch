// Package recipe runs project bootstraps (recipes) and reads their progress.
//
// A recipe is recipes/<name>/recipe.yml pointing at an executable entry
// script. The script stays the authority for its own arguments; the manifest
// describes them so the CLI can validate early and the TUI can build a form.
// Scripts may report progress as JSON lines on the file descriptor named by
// DEVARCH_PROGRESS_FD (see scripts/devarch/lib/platform.sh).
package recipe

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Arg describes one recipe argument.
type Arg struct {
	Name       string `yaml:"name" json:"name"`
	Flag       string `yaml:"flag" json:"flag,omitempty"`
	Positional bool   `yaml:"positional" json:"positional,omitempty"`
	Required   bool   `yaml:"required" json:"required,omitempty"`
	Help       string `yaml:"help" json:"help,omitempty"`
	// Type is "string" (default), "bool" or "path".
	Type    string `yaml:"type" json:"type,omitempty"`
	Pattern string `yaml:"pattern" json:"pattern,omitempty"`
	Default string `yaml:"default" json:"default,omitempty"`
	// Choices lists allowed values; ChoicesFrom globs files relative to the
	// checkout instead, and may reference other args as {name}.
	Choices     []string `yaml:"choices" json:"choices,omitempty"`
	ChoicesFrom string   `yaml:"choices_from" json:"choices_from,omitempty"`
	// ChoiceIs selects what a ChoicesFrom match contributes: "stem" (file name
	// without extension, default) or "dir" (parent directory name).
	ChoiceIs string `yaml:"choice_is" json:"choice_is,omitempty"`
	Confirm  string `yaml:"confirm" json:"confirm,omitempty"`
}

// Recipe is a parsed recipe.yml.
type Recipe struct {
	Name        string   `yaml:"name" json:"name"`
	Title       string   `yaml:"title" json:"title"`
	Description string   `yaml:"description" json:"description,omitempty"`
	Entry       string   `yaml:"entry" json:"entry"`
	Requires    []string `yaml:"requires" json:"requires,omitempty"`
	Args        []Arg    `yaml:"args" json:"args,omitempty"`
	// Dir is the recipe directory; Root is the checkout entry paths resolve from.
	Dir  string `yaml:"-" json:"dir"`
	Root string `yaml:"-" json:"-"`
}

// EntryPath is the absolute entry script path.
func (r Recipe) EntryPath() string { return filepath.Join(r.Root, r.Entry) }

// Load reads every recipe under dirs; earlier dirs override later ones by name.
func Load(root string, dirs ...string) ([]Recipe, error) {
	byName := map[string]Recipe{}
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			r, err := LoadFile(root, filepath.Join(dirs[i], e.Name(), "recipe.yml"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			byName[r.Name] = r
		}
	}
	out := make([]Recipe, 0, len(byName))
	for _, r := range byName {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// LoadFile parses one recipe.yml.
func LoadFile(root, path string) (Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, err
	}
	var r Recipe
	if err := yaml.Unmarshal(data, &r); err != nil {
		return Recipe{}, fmt.Errorf("%s: %w", path, err)
	}
	r.Dir, r.Root = filepath.Dir(path), root
	if r.Name == "" {
		r.Name = filepath.Base(r.Dir)
	}
	if r.Entry == "" {
		return Recipe{}, fmt.Errorf("%s: entry is required", path)
	}
	for _, a := range r.Args {
		if a.Pattern != "" {
			if _, err := regexp.Compile(a.Pattern); err != nil {
				return Recipe{}, fmt.Errorf("%s: arg %s: %w", path, a.Name, err)
			}
		}
	}
	return r, nil
}

// Find returns the named recipe.
func Find(recipes []Recipe, name string) (Recipe, error) {
	for _, r := range recipes {
		if r.Name == name {
			return r, nil
		}
	}
	names := make([]string, len(recipes))
	for i, r := range recipes {
		names[i] = r.Name
	}
	return Recipe{}, fmt.Errorf("unknown recipe %q (available: %s)", name, strings.Join(names, ", "))
}

// ChoicesFor resolves an argument's allowed values given other values.
func (r Recipe) ChoicesFor(a Arg, values map[string]string) []string {
	if len(a.Choices) > 0 || a.ChoicesFrom == "" {
		return a.Choices
	}
	pattern := a.ChoicesFrom
	for k, v := range values {
		pattern = strings.ReplaceAll(pattern, "{"+k+"}", v)
	}
	if strings.Contains(pattern, "{") {
		return nil // depends on a value that is not set yet
	}
	matches, _ := filepath.Glob(filepath.Join(r.Root, pattern))
	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		c := strings.TrimSuffix(filepath.Base(m), filepath.Ext(m))
		if a.ChoiceIs == "dir" {
			c = filepath.Base(filepath.Dir(m))
		}
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// Parse maps command-line arguments onto declared args. Unknown flags are
// left for the script, which stays the authority on its interface.
func (r Recipe) Parse(argv []string) map[string]string {
	values := map[string]string{}
	byFlag := map[string]Arg{}
	var positional []Arg
	for _, a := range r.Args {
		if a.Flag != "" {
			byFlag[a.Flag] = a
		}
		if a.Positional {
			positional = append(positional, a)
		}
	}
	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		if strings.HasPrefix(tok, "-") {
			flag, val, hasVal := strings.Cut(tok, "=")
			a, ok := byFlag[flag]
			if !ok {
				continue
			}
			switch {
			case a.Type == "bool":
				values[a.Name] = "true"
			case hasVal:
				values[a.Name] = val
			case i+1 < len(argv):
				i++
				values[a.Name] = argv[i]
			}
			continue
		}
		if len(positional) > 0 {
			values[positional[0].Name] = tok
			positional = positional[1:]
		}
	}
	return values
}

// Validate checks declared patterns and choices. Missing required values are
// left to the script, which may infer them (WordPress discovers its site name
// from the current directory).
func (r Recipe) Validate(values map[string]string) error {
	for _, a := range r.Args {
		v, ok := values[a.Name]
		if !ok || v == "" {
			continue
		}
		if a.Pattern != "" && !regexp.MustCompile(a.Pattern).MatchString(v) {
			return fmt.Errorf("%s %q must match %s", a.Name, v, a.Pattern)
		}
		if choices := r.ChoicesFor(a, values); len(choices) > 0 && !contains(choices, v) {
			return fmt.Errorf("%s %q is not one of: %s", a.Name, v, strings.Join(choices, ", "))
		}
	}
	return nil
}

// Argv renders values as script arguments in manifest order.
func (r Recipe) Argv(values map[string]string) []string {
	var argv []string
	for _, a := range r.Args {
		v, ok := values[a.Name]
		if !ok || v == "" {
			continue
		}
		switch {
		case a.Positional:
			argv = append(argv, v)
		case a.Type == "bool":
			if v == "true" {
				argv = append(argv, a.Flag)
			}
		default:
			argv = append(argv, a.Flag, v)
		}
	}
	return argv
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Event is one progress line from a recipe script.
type Event struct {
	Step    string `json:"step"`
	State   string `json:"state"` // start, done, warn, fail
	Message string `json:"message,omitempty"`
}

// Command builds the process for a recipe run. env is layered over the
// current environment.
func (r Recipe) Command(ctx context.Context, argv []string, env []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, r.EntryPath(), argv...)
	cmd.Dir = r.Root
	cmd.Env = append(os.Environ(), env...)
	return cmd
}

// RunWithProgress runs the recipe with a progress pipe on fd 3, streaming
// combined output to out and events to onEvent.
func (r Recipe) RunWithProgress(ctx context.Context, argv, env []string, out io.Writer, onEvent func(Event)) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	cmd := r.Command(ctx, argv, append(env, "DEVARCH_PROGRESS_FD=3"))
	cmd.Stdout, cmd.Stderr = out, out
	cmd.ExtraFiles = []*os.File{pw}
	if err := cmd.Start(); err != nil {
		pr.Close()
		pw.Close()
		return err
	}
	pw.Close() // the child holds its own copy
	done := make(chan struct{})
	go func() {
		defer close(done)
		ReadEvents(pr, onEvent)
	}()
	err = cmd.Wait()
	<-done
	pr.Close()
	return err
}

// ReadEvents parses JSON progress lines, ignoring malformed ones.
func ReadEvents(rd io.Reader, onEvent func(Event)) {
	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.Step != "" && onEvent != nil {
			onEvent(e)
		}
	}
}
