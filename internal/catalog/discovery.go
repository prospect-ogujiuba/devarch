package catalog

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const TemplateFilename = "template.yaml"

// DiscoverTemplateFiles walks the provided catalog roots and returns the
// canonical template documents in deterministic path order.
func DiscoverTemplateFiles(roots []string) ([]string, error) {
	if len(roots) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	paths := make([]string, 0)
	for _, root := range roots {
		cleanRoot := filepath.Clean(root)

		info, err := os.Stat(cleanRoot)
		if err != nil {
			return nil, fmt.Errorf("stat catalog root %s: %w", cleanRoot, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("catalog root %s: not a directory", cleanRoot)
		}

		if err := filepath.WalkDir(cleanRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Base(path) != TemplateFilename {
				return nil
			}

			cleanPath := filepath.Clean(path)
			if _, ok := seen[cleanPath]; ok {
				return nil
			}

			seen[cleanPath] = struct{}{}
			paths = append(paths, cleanPath)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("walk catalog root %s: %w", cleanRoot, err)
		}
	}

	sort.Strings(paths)
	return paths, nil
}
