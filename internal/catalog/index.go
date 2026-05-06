package catalog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"gopkg.in/yaml.v3"
)

// Template is the minimal in-memory catalog model used by discovery and lookup.
type Template struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   TemplateMetadata `yaml:"metadata"`
	Spec       TemplateSpec     `yaml:"spec"`
	Path       string           `yaml:"-"`
}

type TemplateMetadata struct {
	Name        string   `yaml:"name"`
	Tags        []string `yaml:"tags,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type TemplateSpec struct {
	Runtime map[string]any   `yaml:"runtime"`
	Env     map[string]any   `yaml:"env,omitempty"`
	Ports   []TemplatePort   `yaml:"ports,omitempty"`
	Volumes []TemplateVolume `yaml:"volumes,omitempty"`
	Imports []TemplateImport `yaml:"imports,omitempty"`
	Exports []TemplateExport `yaml:"exports,omitempty"`
	Health  map[string]any   `yaml:"health,omitempty"`
	Develop map[string]any   `yaml:"develop,omitempty"`
}

type TemplatePort struct {
	Host      int    `yaml:"host,omitempty"`
	Container int    `yaml:"container"`
	Protocol  string `yaml:"protocol,omitempty"`
	HostIP    string `yaml:"hostIP,omitempty"`
}

type TemplateVolume struct {
	Source   string `yaml:"source,omitempty"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"readOnly,omitempty"`
	Kind     string `yaml:"kind,omitempty"`
}

type TemplateImport struct {
	Contract string `yaml:"contract"`
	From     string `yaml:"from,omitempty"`
	Alias    string `yaml:"alias,omitempty"`
}

type TemplateExport struct {
	Contract string            `yaml:"contract"`
	Env      map[string]string `yaml:"env,omitempty"`
}

func (e *TemplateExport) UnmarshalYAML(node *yaml.Node) error {
	type exportObject struct {
		Contract string            `yaml:"contract"`
		Env      map[string]string `yaml:"env,omitempty"`
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var contract string
		if err := node.Decode(&contract); err != nil {
			return err
		}
		e.Contract = contract
		e.Env = nil
		return nil
	case yaml.MappingNode:
		var value exportObject
		if err := node.Decode(&value); err != nil {
			return err
		}
		e.Contract = value.Contract
		e.Env = value.Env
		return nil
	default:
		return fmt.Errorf("decode template export: unsupported YAML node kind %d", node.Kind)
	}
}

// DuplicateTemplateNameError reports an ambiguous template name across two files.
type DuplicateTemplateNameError struct {
	Name       string
	FirstPath  string
	SecondPath string
}

func (e *DuplicateTemplateNameError) Error() string {
	return fmt.Sprintf("duplicate template name %q in %s and %s", e.Name, e.FirstPath, e.SecondPath)
}

// Index provides deterministic catalog template lookups.
type Index struct {
	templates []*Template
	byName    map[string]*Template
}

// LoadIndex loads, validates, decodes, and indexes template files.
func LoadIndex(paths []string) (*Index, error) {
	cleanPaths := uniqueSortedPaths(paths)
	index := &Index{
		templates: make([]*Template, 0, len(cleanPaths)),
		byName:    make(map[string]*Template, len(cleanPaths)),
	}

	for _, path := range cleanPaths {
		template, err := loadTemplate(path)
		if err != nil {
			return nil, err
		}

		if existing, ok := index.byName[template.Metadata.Name]; ok {
			return nil, &DuplicateTemplateNameError{
				Name:       template.Metadata.Name,
				FirstPath:  existing.Path,
				SecondPath: template.Path,
			}
		}

		index.templates = append(index.templates, template)
		index.byName[template.Metadata.Name] = template

	}

	sortTemplates(index.templates)

	return index, nil
}

// Templates returns all indexed templates in deterministic order.
func (i *Index) Templates() []*Template {
	if i == nil || len(i.templates) == 0 {
		return nil
	}
	return append([]*Template(nil), i.templates...)
}

// ByName returns one template by its canonical metadata.name.
func (i *Index) ByName(name string) (*Template, bool) {
	if i == nil {
		return nil, false
	}
	template, ok := i.byName[name]
	return template, ok
}

func loadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %s: %w", path, err)
	}
	if err := spec.ValidateTemplateBytes(data); err != nil {
		return nil, fmt.Errorf("validate template %s: %w", path, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("decode template %s: %w", path, err)
	}
	template.Path = filepath.Clean(path)
	return &template, nil
}

func uniqueSortedPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		if _, ok := seen[cleanPath]; ok {
			continue
		}
		seen[cleanPath] = struct{}{}
		unique = append(unique, cleanPath)
	}

	sort.Strings(unique)
	return unique
}

func sortTemplates(templates []*Template) {
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].Metadata.Name != templates[j].Metadata.Name {
			return templates[i].Metadata.Name < templates[j].Metadata.Name
		}
		return templates[i].Path < templates[j].Path
	})
}
