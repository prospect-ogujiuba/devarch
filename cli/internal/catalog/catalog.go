// Package catalog discovers services-library entries and their x-devarch metadata.
package catalog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var segmentRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// Version describes a switchable image or build version.
type Version struct {
	Var     string   `yaml:"var" json:"var"`
	Default string   `yaml:"default" json:"default"`
	Choices []string `yaml:"choices" json:"choices,omitempty"`
	// Rebuild marks services built from a Dockerfile that takes Var as a build arg.
	Rebuild bool `yaml:"rebuild" json:"rebuild,omitempty"`
	// Data is the data-volume compatibility rule: "", "upgrade-only" or "same-major".
	Data string `yaml:"data" json:"data,omitempty"`
}

// Ready is an optional readiness probe run inside the primary container.
type Ready struct {
	Exec []string `yaml:"exec" json:"exec,omitempty"`
}

// Meta is the optional top-level x-devarch block of a compose file.
type Meta struct {
	Title       string   `yaml:"title" json:"title,omitempty"`
	Description string   `yaml:"description" json:"description,omitempty"`
	Tags        []string `yaml:"tags" json:"tags,omitempty"`
	URLs        []string `yaml:"urls" json:"urls,omitempty"`
	Notes       string   `yaml:"notes" json:"notes,omitempty"`
	Version     *Version `yaml:"version" json:"version,omitempty"`
	Ready       *Ready   `yaml:"ready" json:"ready,omitempty"`
	Requires    []string `yaml:"requires" json:"requires,omitempty"`
}

// Container is one compose service inside a catalog entry.
type Container struct {
	Service       string   `json:"service"`
	ContainerName string   `json:"container_name,omitempty"`
	Image         string   `json:"image,omitempty"`
	Build         bool     `json:"build,omitempty"`
	Ports         []string `json:"ports,omitempty"`
}

// Service is one catalog entry: a directory holding compose.yml.
type Service struct {
	ID          string      `json:"id"`
	Category    string      `json:"category"`
	Name        string      `json:"name"`
	Dir         string      `json:"dir"`
	ComposeFile string      `json:"compose_file"`
	Source      string      `json:"source"`
	Overrides   bool        `json:"overrides,omitempty"`
	Meta        Meta        `json:"meta"`
	Containers  []Container `json:"containers"`
	// Volumes are the podman volume names `compose down --volumes` deletes:
	// explicit names, else <project>_<key>; external volumes are excluded.
	Volumes []string `json:"volumes,omitempty"`
}

// Title returns the display title, falling back to the service name.
func (s Service) Title() string {
	if s.Meta.Title != "" {
		return s.Meta.Title
	}
	return s.Name
}

// Primary returns the compose service named like the directory, else the first.
func (s Service) Primary() Container {
	for _, c := range s.Containers {
		if c.Service == s.Name {
			return c
		}
	}
	if len(s.Containers) > 0 {
		return s.Containers[0]
	}
	return Container{}
}

// HasTag reports whether the service carries tag.
func (s Service) HasTag(tag string) bool {
	for _, t := range s.Meta.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Path is one catalog search location. Earlier paths override later ones.
type Path struct {
	Dir    string
	Source string
}

// Catalog is the merged set of services from every search path.
type Catalog struct {
	Services []Service
	byID     map[string]int
}

// Problem is a non-fatal discovery issue; lint treats them as failures.
type Problem struct {
	Path string
	Err  error
}

func (p Problem) Error() string { return p.Path + ": " + p.Err.Error() }

// Load scans each path for <category>/<name>/compose.yml. Missing search paths
// are skipped; malformed entries are reported as problems and left out.
func Load(paths []Path) (*Catalog, []Problem) {
	c := &Catalog{byID: map[string]int{}}
	var problems []Problem
	for _, p := range paths {
		entries, err := os.ReadDir(p.Dir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			problems = append(problems, Problem{p.Dir, err})
			continue
		}
		for _, cat := range entries {
			catDir := filepath.Join(p.Dir, cat.Name())
			if !cat.IsDir() {
				if cat.Type()&os.ModeSymlink != 0 {
					problems = append(problems, Problem{catDir, errors.New("catalog category may not be a symbolic link")})
				}
				continue
			}
			if !segmentRE.MatchString(cat.Name()) {
				problems = append(problems, Problem{catDir, errors.New("malformed catalog category name")})
				continue
			}
			svcs, err := os.ReadDir(catDir)
			if err != nil {
				problems = append(problems, Problem{catDir, err})
				continue
			}
			for _, s := range svcs {
				dir := filepath.Join(catDir, s.Name())
				if !s.IsDir() {
					if s.Type()&os.ModeSymlink != 0 {
						problems = append(problems, Problem{dir, errors.New("service path may not be a symbolic link")})
					}
					continue
				}
				svc, err := loadService(p, cat.Name(), s.Name(), dir)
				if err != nil {
					problems = append(problems, Problem{dir, err})
					continue
				}
				if _, seen := c.byID[svc.ID]; seen {
					// An earlier search path already provides this ID.
					c.Services[c.byID[svc.ID]].Overrides = true
					continue
				}
				c.byID[svc.ID] = len(c.Services)
				c.Services = append(c.Services, svc)
			}
		}
	}
	sort.Slice(c.Services, func(i, j int) bool { return c.Services[i].ID < c.Services[j].ID })
	for i, s := range c.Services {
		c.byID[s.ID] = i
	}
	return c, problems
}

type composeFile struct {
	XDevarch *Meta `yaml:"x-devarch"`
	Services map[string]struct {
		ContainerName string    `yaml:"container_name"`
		Image         string    `yaml:"image"`
		Build         yaml.Node `yaml:"build"`
		Ports         []any     `yaml:"ports"`
	} `yaml:"services"`
	Volumes map[string]*struct {
		Name     string `yaml:"name"`
		External bool   `yaml:"external"`
	} `yaml:"volumes"`
}

func loadService(p Path, category, name, dir string) (Service, error) {
	if !segmentRE.MatchString(name) {
		return Service{}, errors.New("malformed catalog service name")
	}
	file := filepath.Join(dir, "compose.yml")
	info, err := os.Lstat(file)
	if err != nil {
		return Service{}, fmt.Errorf("compose.yml is missing")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Service{}, errors.New("compose.yml may not be a symbolic link")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return Service{}, err
	}
	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return Service{}, fmt.Errorf("malformed compose file: %w", err)
	}
	if len(cf.Services) == 0 {
		return Service{}, errors.New("malformed compose file: top-level services mapping is missing")
	}
	svc := Service{
		ID:          category + "/" + name,
		Category:    category,
		Name:        name,
		Dir:         dir,
		ComposeFile: file,
		Source:      p.Source,
	}
	if cf.XDevarch != nil {
		svc.Meta = *cf.XDevarch
	}
	for sn, s := range cf.Services {
		c := Container{Service: sn, ContainerName: s.ContainerName, Image: s.Image, Build: !s.Build.IsZero()}
		for _, port := range s.Ports {
			c.Ports = append(c.Ports, fmt.Sprint(port))
		}
		svc.Containers = append(svc.Containers, c)
	}
	sort.Slice(svc.Containers, func(i, j int) bool { return svc.Containers[i].Service < svc.Containers[j].Service })
	for key, v := range cf.Volumes {
		switch {
		case v != nil && v.External:
			continue
		case v != nil && v.Name != "":
			svc.Volumes = append(svc.Volumes, v.Name)
		default:
			svc.Volumes = append(svc.Volumes, name+"_"+key)
		}
	}
	sort.Strings(svc.Volumes)
	return svc, nil
}

// Get returns the service with the exact canonical ID.
func (c *Catalog) Get(id string) (Service, bool) {
	i, ok := c.byID[id]
	if !ok {
		return Service{}, false
	}
	return c.Services[i], true
}

// AmbiguousError lists candidates for a short name that matches several IDs.
type AmbiguousError struct {
	Query      string
	Candidates []string
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("ambiguous service name %q; candidates: %s", e.Query, strings.Join(e.Candidates, ", "))
}

// Resolve finds an exact canonical ID (category/name) or a unique short name.
// Exact IDs are never broadened to fuzzy matches.
func (c *Catalog) Resolve(query string) (Service, error) {
	if query == "" {
		return Service{}, errors.New("a service ID is required")
	}
	if strings.Contains(query, "/") {
		if strings.Count(query, "/") != 1 {
			return Service{}, fmt.Errorf("invalid canonical service ID: %s", query)
		}
		if s, ok := c.Get(query); ok {
			return s, nil
		}
		return Service{}, fmt.Errorf("unknown service: %s", query)
	}
	if !segmentRE.MatchString(query) {
		return Service{}, fmt.Errorf("invalid service name: %s", query)
	}
	var matches []Service
	for _, s := range c.Services {
		if s.Name == query {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return Service{}, fmt.Errorf("unknown service: %s (try `devarch ls %s`)", query, query)
	default:
		ids := make([]string, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		return Service{}, &AmbiguousError{Query: query, Candidates: ids}
	}
}

// Search returns services matching a free-text query (ID, title, description,
// tags) and, when tag is set, carrying that tag.
func (c *Catalog) Search(query, tag string) []Service {
	q := strings.ToLower(query)
	var out []Service
	for _, s := range c.Services {
		if tag != "" && !s.HasTag(tag) {
			continue
		}
		if q != "" {
			hay := strings.ToLower(strings.Join(append([]string{s.ID, s.Meta.Title, s.Meta.Description}, s.Meta.Tags...), " "))
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, s)
	}
	return out
}

// WithRequires expands services with their x-devarch requires, dependencies
// first, without duplicates. It fails on unknown IDs and cycles.
func (c *Catalog) WithRequires(svcs []Service) ([]Service, error) {
	var out []Service
	state := map[string]int{} // 1 visiting, 2 done
	var visit func(s Service, chain []string) error
	visit = func(s Service, chain []string) error {
		switch state[s.ID] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("requires cycle: %s", strings.Join(append(chain, s.ID), " -> "))
		}
		state[s.ID] = 1
		for _, r := range s.Meta.Requires {
			dep, err := c.Resolve(r)
			if err != nil {
				return fmt.Errorf("%s requires %s: %w", s.ID, r, err)
			}
			if err := visit(dep, append(chain, s.ID)); err != nil {
				return err
			}
		}
		state[s.ID] = 2
		out = append(out, s)
		return nil
	}
	for _, s := range svcs {
		if err := visit(s, nil); err != nil {
			return nil, err
		}
	}
	return out, nil
}
