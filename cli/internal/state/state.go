// Package state owns DevArch's only persistent files, under ~/.config/devarch.
package state

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Dir returns $DEVARCH_CONFIG_HOME, else $XDG_CONFIG_HOME/devarch, else ~/.config/devarch.
func Dir() (string, error) {
	if d := os.Getenv("DEVARCH_CONFIG_HOME"); d != "" {
		return d, nil
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "devarch"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "devarch"), nil
}

// Config is config.yml.
type Config struct {
	Root string `yaml:"root,omitempty"`
}

func LoadConfig(dir string) (Config, error) {
	var c Config
	data, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("%s: %w", filepath.Join(dir, "config.yml"), err)
	}
	return c, nil
}

func SaveConfig(dir string, c Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "config.yml"), data)
}

var varRE = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// Versions is versions.env: selected service versions passed to compose as
// process environment.
type Versions map[string]string

func LoadVersions(dir string) (Versions, error) {
	v := Versions{}
	f, err := os.Open(filepath.Join(dir, "versions.env"))
	if errors.Is(err, os.ErrNotExist) {
		return v, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, val, ok := strings.Cut(line, "=")
		if !ok || !varRE.MatchString(k) {
			return nil, fmt.Errorf("versions.env:%d: expected NAME=value", n)
		}
		v[k] = val
	}
	return v, sc.Err()
}

func SaveVersions(dir string, v Versions) error {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("# Managed by `devarch use`. Values are passed to podman compose as environment.\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, v[k])
	}
	return writeFile(filepath.Join(dir, "versions.env"), []byte(b.String()))
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
