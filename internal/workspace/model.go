package workspace

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workspace is the normalized in-memory representation of a DevArch V2
// workspace manifest. ManifestPath and ManifestDir are internal metadata and are
// intentionally omitted from serialized output.
type Workspace struct {
	APIVersion string               `yaml:"apiVersion" json:"apiVersion,omitempty"`
	Kind       string               `yaml:"kind" json:"kind,omitempty"`
	Metadata   Metadata             `yaml:"metadata" json:"metadata"`
	Runtime    RuntimePreferences   `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Catalog    Catalog              `yaml:"catalog,omitempty" json:"catalog,omitempty"`
	Policies   Policies             `yaml:"policies,omitempty" json:"policies,omitempty"`
	Secrets    map[string]any       `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Profiles   map[string]any       `yaml:"profiles,omitempty" json:"profiles,omitempty"`
	Resources  map[string]*Resource `yaml:"resources" json:"resources"`

	ManifestPath string `yaml:"-" json:"-"`
	ManifestDir  string `yaml:"-" json:"-"`
}

type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type RuntimePreferences struct {
	Provider        string `yaml:"provider,omitempty" json:"provider,omitempty"`
	IsolatedNetwork bool   `yaml:"isolatedNetwork,omitempty" json:"isolatedNetwork,omitempty"`
	NamingStrategy  string `yaml:"namingStrategy,omitempty" json:"namingStrategy,omitempty"`
}

type Catalog struct {
	Sources         []string `yaml:"sources,omitempty" json:"sources,omitempty"`
	ResolvedSources []string `yaml:"-" json:"-"`
}

type Policies struct {
	AutoWire     bool   `yaml:"autoWire,omitempty" json:"autoWire,omitempty"`
	SecretSource string `yaml:"secretSource,omitempty" json:"secretSource,omitempty"`
}

type Resource struct {
	Template  string              `yaml:"template,omitempty" json:"template,omitempty"`
	Source    *Source             `yaml:"source,omitempty" json:"source,omitempty"`
	Enabled   *bool               `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Env       map[string]EnvValue `yaml:"env,omitempty" json:"env,omitempty"`
	Ports     []Port              `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes   []Volume            `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	DependsOn []string            `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
	Imports   []Import            `yaml:"imports,omitempty" json:"imports,omitempty"`
	Exports   []Export            `yaml:"exports,omitempty" json:"exports,omitempty"`
	Health    *Health             `yaml:"health,omitempty" json:"health,omitempty"`
	Domains   []string            `yaml:"domains,omitempty" json:"domains,omitempty"`
	Develop   map[string]any      `yaml:"develop,omitempty" json:"develop,omitempty"`
	Overrides map[string]any      `yaml:"overrides,omitempty" json:"overrides,omitempty"`
}

type Source struct {
	Type         string `yaml:"type" json:"type"`
	Path         string `yaml:"path" json:"path"`
	Service      string `yaml:"service,omitempty" json:"service,omitempty"`
	ResolvedPath string `yaml:"-" json:"-"`
}

type Port struct {
	Host      int    `yaml:"host,omitempty" json:"host,omitempty"`
	Container int    `yaml:"container" json:"container"`
	Protocol  string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	HostIP    string `yaml:"hostIP,omitempty" json:"hostIP,omitempty"`
}

type Volume struct {
	Source   string `yaml:"source,omitempty" json:"source,omitempty"`
	Target   string `yaml:"target" json:"target"`
	ReadOnly bool   `yaml:"readOnly,omitempty" json:"readOnly,omitempty"`
	Kind     string `yaml:"kind,omitempty" json:"kind,omitempty"`
}

type Import struct {
	Contract string `yaml:"contract" json:"contract"`
	From     string `yaml:"from,omitempty" json:"from,omitempty"`
	Alias    string `yaml:"alias,omitempty" json:"alias,omitempty"`
}

type Export struct {
	Contract string            `yaml:"contract" json:"contract"`
	Env      map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

func (e *Export) UnmarshalYAML(node *yaml.Node) error {
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
		return fmt.Errorf("decode export: unsupported YAML node kind %d", node.Kind)
	}
}

type Health struct {
	Test        StringList `yaml:"test,omitempty" json:"test,omitempty"`
	Interval    string     `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout     string     `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries     int        `yaml:"retries,omitempty" json:"retries,omitempty"`
	StartPeriod string     `yaml:"startPeriod,omitempty" json:"startPeriod,omitempty"`
}

// StringList accepts either a scalar string or a string array and normalizes the
// result to a deterministic string slice.
type StringList []string

func (s *StringList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var value string
		if err := node.Decode(&value); err != nil {
			return err
		}
		*s = StringList{value}
		return nil
	case yaml.SequenceNode:
		var values []string
		if err := node.Decode(&values); err != nil {
			return err
		}
		*s = StringList(values)
		return nil
	default:
		return fmt.Errorf("decode string list: unsupported YAML node kind %d", node.Kind)
	}
}

type EnvValueKind string

const (
	EnvValueString    EnvValueKind = "string"
	EnvValueNumber    EnvValueKind = "number"
	EnvValueBool      EnvValueKind = "bool"
	EnvValueSecretRef EnvValueKind = "secretRef"
)

// EnvValue preserves the workspace/template env union while remaining easy to
// serialize in the scalar-or-secretRef form required by fixtures and goldens.
type EnvValue struct {
	kind      EnvValueKind
	stringVal string
	numberVal string
	boolVal   bool
	secretRef string
}

func StringEnvValue(value string) EnvValue {
	return EnvValue{kind: EnvValueString, stringVal: value}
}

func NumberEnvValue(value string) EnvValue {
	return EnvValue{kind: EnvValueNumber, numberVal: value}
}

func BoolEnvValue(value bool) EnvValue {
	return EnvValue{kind: EnvValueBool, boolVal: value}
}

func SecretRefEnvValue(secretRef string) EnvValue {
	return EnvValue{kind: EnvValueSecretRef, secretRef: secretRef}
}

func (v *EnvValue) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		var value struct {
			SecretRef string `yaml:"secretRef"`
		}
		if err := node.Decode(&value); err != nil {
			return err
		}
		*v = SecretRefEnvValue(value.SecretRef)
		return nil
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			var value string
			if err := node.Decode(&value); err != nil {
				return err
			}
			*v = StringEnvValue(value)
			return nil
		case "!!int", "!!float":
			var value any
			if err := node.Decode(&value); err != nil {
				return err
			}
			converted, err := EnvValueFromAny(value)
			if err != nil {
				return err
			}
			*v = converted
			return nil
		case "!!bool":
			var value bool
			if err := node.Decode(&value); err != nil {
				return err
			}
			*v = BoolEnvValue(value)
			return nil
		default:
			var value any
			if err := node.Decode(&value); err != nil {
				return err
			}
			converted, err := EnvValueFromAny(value)
			if err != nil {
				return err
			}
			*v = converted
			return nil
		}
	default:
		return fmt.Errorf("decode env value: unsupported YAML node kind %d", node.Kind)
	}
}

func (v EnvValue) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case EnvValueString:
		return json.Marshal(v.stringVal)
	case EnvValueNumber:
		return []byte(v.numberVal), nil
	case EnvValueBool:
		return json.Marshal(v.boolVal)
	case EnvValueSecretRef:
		return json.Marshal(map[string]string{"secretRef": v.secretRef})
	default:
		return []byte("null"), nil
	}
}

func (v EnvValue) MarshalYAML() (any, error) {
	switch v.kind {
	case EnvValueString:
		return v.stringVal, nil
	case EnvValueNumber:
		return v.numberVal, nil
	case EnvValueBool:
		return v.boolVal, nil
	case EnvValueSecretRef:
		return map[string]string{"secretRef": v.secretRef}, nil
	default:
		return nil, nil
	}
}

func (v EnvValue) Kind() EnvValueKind {
	return v.kind
}

func (v EnvValue) Clone() EnvValue {
	return v
}

func (v EnvValue) SecretRef() (string, bool) {
	if v.kind != EnvValueSecretRef {
		return "", false
	}
	return v.secretRef, true
}

func (v EnvValue) Text() string {
	switch v.kind {
	case EnvValueString:
		return v.stringVal
	case EnvValueNumber:
		return v.numberVal
	case EnvValueBool:
		return strconv.FormatBool(v.boolVal)
	case EnvValueSecretRef:
		return v.secretRef
	default:
		return ""
	}
}

func (v EnvValue) IsZero() bool {
	return v.kind == ""
}

// EnvValueFromAny converts schema-valid env values from generic decoded maps.
func EnvValueFromAny(value any) (EnvValue, error) {
	switch typed := value.(type) {
	case EnvValue:
		return typed.Clone(), nil
	case string:
		return StringEnvValue(typed), nil
	case bool:
		return BoolEnvValue(typed), nil
	case int:
		return NumberEnvValue(strconv.Itoa(typed)), nil
	case int8:
		return NumberEnvValue(strconv.FormatInt(int64(typed), 10)), nil
	case int16:
		return NumberEnvValue(strconv.FormatInt(int64(typed), 10)), nil
	case int32:
		return NumberEnvValue(strconv.FormatInt(int64(typed), 10)), nil
	case int64:
		return NumberEnvValue(strconv.FormatInt(typed, 10)), nil
	case uint:
		return NumberEnvValue(strconv.FormatUint(uint64(typed), 10)), nil
	case uint8:
		return NumberEnvValue(strconv.FormatUint(uint64(typed), 10)), nil
	case uint16:
		return NumberEnvValue(strconv.FormatUint(uint64(typed), 10)), nil
	case uint32:
		return NumberEnvValue(strconv.FormatUint(uint64(typed), 10)), nil
	case uint64:
		return NumberEnvValue(strconv.FormatUint(typed, 10)), nil
	case float32:
		return NumberEnvValue(strconv.FormatFloat(float64(typed), 'f', -1, 32)), nil
	case float64:
		return NumberEnvValue(strconv.FormatFloat(typed, 'f', -1, 64)), nil
	case map[string]any:
		secretRef, ok := typed["secretRef"].(string)
		if !ok {
			return EnvValue{}, fmt.Errorf("decode env value: invalid secretRef object")
		}
		return SecretRefEnvValue(secretRef), nil
	case map[any]any:
		secretRaw, ok := typed["secretRef"]
		if !ok {
			return EnvValue{}, fmt.Errorf("decode env value: invalid secretRef object")
		}
		secretRef, ok := secretRaw.(string)
		if !ok {
			return EnvValue{}, fmt.Errorf("decode env value: invalid secretRef object")
		}
		return SecretRefEnvValue(secretRef), nil
	default:
		return EnvValue{}, fmt.Errorf("decode env value: unsupported type %T", value)
	}
}

func (w *Workspace) SortedResourceKeys() []string {
	if w == nil || len(w.Resources) == 0 {
		return nil
	}

	keys := make([]string, 0, len(w.Resources))
	for key := range w.Resources {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (w *Workspace) SortedResources() []*Resource {
	keys := w.SortedResourceKeys()
	if len(keys) == 0 {
		return nil
	}

	resources := make([]*Resource, 0, len(keys))
	for _, key := range keys {
		resources = append(resources, w.Resources[key])
	}
	return resources
}

func (w *Workspace) ResolvedCatalogSources() []string {
	if w == nil || len(w.Catalog.ResolvedSources) == 0 {
		return nil
	}
	return append([]string(nil), w.Catalog.ResolvedSources...)
}

func (r *Resource) EnabledValue() bool {
	if r == nil || r.Enabled == nil {
		return true
	}
	return *r.Enabled
}

func (r *Resource) SetEnabled(value bool) {
	r.Enabled = boolPtr(value)
}

func boolPtr(value bool) *bool {
	return &value
}

func cloneEnvMap(values map[string]EnvValue) map[string]EnvValue {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]EnvValue, len(values))
	for key, value := range values {
		cloned[key] = value.Clone()
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneRawMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = cloneRawValue(value)
	}
	return cloned
}

func cloneRawValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneRawMap(typed)
	case map[any]any:
		converted := make(map[string]any, len(typed))
		for key, nested := range typed {
			converted[fmt.Sprint(key)] = cloneRawValue(nested)
		}
		return converted
	case []any:
		cloned := make([]any, len(typed))
		for i := range typed {
			cloned[i] = cloneRawValue(typed[i])
		}
		return cloned
	case []string:
		return append([]string(nil), typed...)
	default:
		return typed
	}
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	sort.Strings(normalized)
	return normalized
}
