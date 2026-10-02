// Package root locates the DevArch checkout.
package root

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// How describes where the root came from.
type How string

const (
	FromEnv    How = "DEVARCH_ROOT"
	FromConfig How = "config"
	FromWalk   How = "walk"
)

// IsRoot reports whether dir contains services-library/.
func IsRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "services-library"))
	return err == nil && info.IsDir()
}

// Find resolves the root: $DEVARCH_ROOT, then the configured root, then a
// walk up from start looking for services-library/.
func Find(configured, start string) (string, How, error) {
	if env := os.Getenv("DEVARCH_ROOT"); env != "" {
		return check(env, FromEnv)
	}
	if configured != "" {
		return check(configured, FromConfig)
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", "", err
	}
	for {
		if IsRoot(dir) {
			return physical(dir), FromWalk, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", errors.New("DevArch checkout not found: run devarch inside it once, or set DEVARCH_ROOT")
		}
		dir = parent
	}
}

func check(dir string, how How) (string, How, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	if !IsRoot(abs) {
		return "", "", fmt.Errorf("%s (from %s) is not a DevArch checkout: services-library/ is missing", abs, how)
	}
	return physical(abs), how, nil
}

func physical(dir string) string {
	if p, err := filepath.EvalSymlinks(dir); err == nil {
		return p
	}
	return dir
}
