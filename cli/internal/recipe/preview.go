package recipe

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Preview declares that an argument's choices are line-oriented profile
// files, one "directive value [option]" per line, so the TUI can show what a
// choice installs before running. The script stays the authority: Go only
// groups lines for display and never interprets them.
type Preview struct {
	// Include names the directive that pulls in another profile from the
	// same directory (by file name).
	Include string `yaml:"include" json:"include,omitempty"`
	// Directives lists the known directives in display order.
	Directives []Directive `yaml:"directives" json:"directives"`
}

// Directive labels one kind of profile line.
type Directive struct {
	Kind  string `yaml:"kind" json:"kind"`
	Label string `yaml:"label" json:"label"`
}

// Group is the items of one directive kind.
type Group struct {
	Label string   `json:"label"`
	Items []string `json:"items"`
}

// ProfilePreview is what a profile choice contains.
type ProfilePreview struct {
	// Description is the file's first "# " comment.
	Description string  `json:"description,omitempty"`
	Groups      []Group `json:"groups,omitempty"`
}

// PreviewChoice reads the profile file behind a choice of arg a.
func (r Recipe) PreviewChoice(a Arg, values map[string]string, choice string) (ProfilePreview, error) {
	if a.Preview == nil || a.ChoicesFrom == "" || choice == "" {
		return ProfilePreview{}, fmt.Errorf("%s has no preview", a.Name)
	}
	pattern := a.ChoicesFrom
	for k, v := range values {
		pattern = strings.ReplaceAll(pattern, "{"+k+"}", v)
	}
	if strings.Count(pattern, "*") != 1 || strings.ContainsAny(choice, `/\*`) {
		return ProfilePreview{}, fmt.Errorf("cannot locate profile %q", choice)
	}
	path := filepath.Join(r.Root, strings.Replace(pattern, "*", choice, 1))
	var p ProfilePreview
	items := map[string][]string{}
	if err := a.Preview.read(path, items, &p, map[string]bool{}); err != nil {
		return p, err
	}
	for _, d := range a.Preview.Directives {
		if len(items[d.Kind]) > 0 {
			p.Groups = append(p.Groups, Group{Label: d.Label, Items: items[d.Kind]})
		}
	}
	return p, nil
}

func (pv *Preview) read(path string, items map[string][]string, p *ProfilePreview, seen map[string]bool) error {
	if seen[path] {
		return fmt.Errorf("%s includes itself", filepath.Base(path))
	}
	seen[path] = true
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	known := map[string]bool{}
	for _, d := range pv.Directives {
		known[d.Kind] = true
	}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "# ") && p.Description == "" && n == 1 {
			p.Description = strings.TrimPrefix(line, "# ")
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		kind, rest, _ := strings.Cut(line, " ")
		rest = strings.Join(strings.Fields(rest), " ")
		switch {
		case pv.Include != "" && kind == pv.Include:
			sub := ProfilePreview{Description: "-"}
			if err := pv.read(filepath.Join(filepath.Dir(path), rest), items, &sub, seen); err != nil {
				return fmt.Errorf("%s:%d: %w", filepath.Base(path), n, err)
			}
		case known[kind]:
			items[kind] = append(items[kind], rest)
		default:
			return fmt.Errorf("%s:%d: unknown directive %q", filepath.Base(path), n, kind)
		}
	}
	return sc.Err()
}
