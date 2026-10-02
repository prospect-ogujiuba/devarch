package state

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is config.yml. Every key is optional; Resolve applies the defaults.
type Config struct {
	Root    string      `yaml:"root,omitempty"`
	Runtime string      `yaml:"runtime,omitempty"`
	AppsDir string      `yaml:"apps_dir,omitempty"`
	Network string      `yaml:"network,omitempty"`
	Hosts   HostsConfig `yaml:"hosts,omitempty"`
	Editor  string      `yaml:"editor,omitempty"`
}

// HostsConfig is the hosts: section of config.yml.
type HostsConfig struct {
	Manage  *bool  `yaml:"manage,omitempty"`
	Address string `yaml:"address,omitempty"`
}

// Defaults for the settable keys: today's behavior.
const (
	DefaultRuntime      = "podman"
	DefaultNetwork      = "microservices-net"
	DefaultHostsAddress = "127.0.0.1"
	DefaultEditor       = "code"
)

// Keys are the settable config.yml keys, in display order.
var Keys = []string{"runtime", "apps_dir", "network", "hosts.manage", "hosts.address", "editor"}

// Settings are the effective values of the settable keys.
type Settings struct {
	Runtime string
	// AppsDir is absolute.
	AppsDir      string
	Network      string
	HostsManage  bool
	HostsAddress string
	// Editor is a command; the directory is appended as its last argument.
	Editor string
	// Sources maps each key to "default" or "config.yml".
	Sources map[string]string
}

// Defaults returns the settings used when config.yml sets nothing.
func Defaults(root string) Settings {
	s, _ := Config{}.Resolve(root)
	return s
}

var networkRE = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)

// Resolve applies defaults and validates the settable keys. A relative
// apps_dir is relative to the checkout root.
func (c Config) Resolve(root string) (Settings, error) {
	s := Settings{Sources: map[string]string{}}
	pick := func(key, val, def string) string {
		if val == "" {
			s.Sources[key] = "default"
			return def
		}
		s.Sources[key] = "config.yml"
		return val
	}
	s.Runtime = pick("runtime", c.Runtime, DefaultRuntime)
	s.AppsDir = pick("apps_dir", c.AppsDir, "apps")
	s.Network = pick("network", c.Network, DefaultNetwork)
	s.HostsAddress = pick("hosts.address", c.Hosts.Address, DefaultHostsAddress)
	s.Editor = pick("editor", c.Editor, DefaultEditor)
	s.HostsManage = true
	s.Sources["hosts.manage"] = "default"
	if c.Hosts.Manage != nil {
		s.HostsManage = *c.Hosts.Manage
		s.Sources["hosts.manage"] = "config.yml"
	}
	if strings.HasPrefix(s.AppsDir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			s.AppsDir = filepath.Join(home, s.AppsDir[2:])
		}
	}
	if !filepath.IsAbs(s.AppsDir) {
		s.AppsDir = filepath.Join(root, s.AppsDir)
	}
	s.AppsDir = filepath.Clean(s.AppsDir)

	switch s.Runtime {
	case "podman":
	case "docker":
		return s, errors.New("runtime: docker is not supported yet")
	default:
		return s, fmt.Errorf("runtime must be podman or docker, not %q", s.Runtime)
	}
	if !networkRE.MatchString(s.Network) {
		return s, fmt.Errorf("network %q is not a valid network name", s.Network)
	}
	if ip := net.ParseIP(s.HostsAddress); ip == nil || ip.To4() == nil || strings.Count(s.HostsAddress, ".") != 3 {
		return s, fmt.Errorf("hosts.address %q must be an IPv4 address", s.HostsAddress)
	}
	if len(strings.Fields(s.Editor)) == 0 {
		return s, errors.New("editor must name a command")
	}
	return s, nil
}

// Get returns the effective value of a settable key as text.
func (s Settings) Get(key string) (string, error) {
	switch key {
	case "runtime":
		return s.Runtime, nil
	case "apps_dir":
		return s.AppsDir, nil
	case "network":
		return s.Network, nil
	case "hosts.manage":
		return strconv.FormatBool(s.HostsManage), nil
	case "hosts.address":
		return s.HostsAddress, nil
	case "editor":
		return s.Editor, nil
	}
	return "", unknownKey(key)
}

func unknownKey(key string) error {
	return fmt.Errorf("unknown key %q; settable keys are %s", key, strings.Join(Keys, ", "))
}

func LoadConfig(dir string) (Config, error) {
	var c Config
	path := filepath.Join(dir, "config.yml")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil && !errors.Is(err, io.EOF) {
		return c, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

// SaveConfig writes c, replacing config.yml.
func SaveConfig(dir string, c Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "config.yml"), data)
}

// SetConfig sets one settable key in config.yml, keeping the file's other
// keys and comments, after checking that the result is valid.
func SetConfig(dir, key, value string) error {
	if !isKey(key) {
		return unknownKey(key)
	}
	return setKey(dir, key, value)
}

func setKey(dir, key, value string) error {
	path := filepath.Join(dir, "config.yml")
	var doc yaml.Node
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	}
	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: value}
	if key == "hosts.manage" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("hosts.manage must be true or false, not %q", value)
		}
		scalar.Tag, scalar.Value = "!!bool", strconv.FormatBool(b)
	} else {
		scalar.Tag = "!!str"
	}
	node := doc.Content[0]
	parts := strings.Split(key, ".")
	for i, p := range parts {
		if node.Kind != yaml.MappingNode {
			return fmt.Errorf("%s: %s is not a mapping", path, strings.Join(parts[:i], "."))
		}
		var next *yaml.Node
		for j := 0; j+1 < len(node.Content); j += 2 {
			if node.Content[j].Value == p {
				next = node.Content[j+1]
				if i == len(parts)-1 {
					scalar.HeadComment, scalar.LineComment = next.HeadComment, next.LineComment
					node.Content[j+1] = scalar
				}
				break
			}
		}
		if next == nil {
			next = scalar
			if i < len(parts)-1 {
				next = &yaml.Node{Kind: yaml.MappingNode}
			}
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: p}, next)
		}
		node = next
	}
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return err
	}
	out := []byte(buf.String())
	var c Config
	dec := yaml.NewDecoder(strings.NewReader(string(out)))
	dec.KnownFields(true)
	if err := dec.Decode(&c); err != nil {
		return err
	}
	if _, err := c.Resolve(""); err != nil {
		return err
	}
	return writeFile(path, out)
}

func isKey(key string) bool {
	for _, k := range Keys {
		if k == key {
			return true
		}
	}
	return false
}

// SetRoot records the checkout location, keeping the rest of config.yml.
func SetRoot(dir, root string) error {
	return setKey(dir, "root", root)
}
