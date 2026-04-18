package appsvc

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/prospect-ogujiuba/devarch/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/internal/spec"
	"github.com/prospect-ogujiuba/devarch/internal/workspace"
)

// DiscoverWorkspaces recursively scans workspace roots for canonical manifest
// files, loads them through workspace.Load, sorts them by workspace name, and
// fails fast on duplicate metadata.name values.
func DiscoverWorkspaces(roots []string) ([]*workspace.Workspace, error) {
	manifestPaths, err := discoverWorkspaceManifestPaths(roots)
	if err != nil {
		return nil, err
	}

	workspaces := make([]*workspace.Workspace, 0, len(manifestPaths))
	seenByName := make(map[string]string, len(manifestPaths))
	for _, manifestPath := range manifestPaths {
		ws, err := workspace.Load(manifestPath)
		if err != nil {
			return nil, err
		}
		if firstPath, ok := seenByName[ws.Metadata.Name]; ok {
			return nil, &DuplicateWorkspaceNameError{
				Name:       ws.Metadata.Name,
				FirstPath:  firstPath,
				SecondPath: ws.ManifestPath,
			}
		}
		seenByName[ws.Metadata.Name] = ws.ManifestPath
		workspaces = append(workspaces, ws)
	}

	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Metadata.Name != workspaces[j].Metadata.Name {
			return workspaces[i].Metadata.Name < workspaces[j].Metadata.Name
		}
		return workspaces[i].ManifestPath < workspaces[j].ManifestPath
	})
	return workspaces, nil
}

// LoadCatalogIndex loads the daemon-configured catalog roots used by the shared
// catalog read endpoints.
func LoadCatalogIndex(roots []string) (*catalog.Index, error) {
	paths, err := catalog.DiscoverTemplateFiles(roots)
	if err != nil {
		return nil, err
	}
	return catalog.LoadIndex(paths)
}

func discoverWorkspaceManifestPaths(roots []string) ([]string, error) {
	if len(roots) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	paths := make([]string, 0)
	for _, root := range roots {
		cleanRoot := filepath.Clean(root)
		info, err := os.Stat(cleanRoot)
		if err != nil {
			return nil, fmt.Errorf("stat workspace root %s: %w", cleanRoot, err)
		}

		switch {
		case info.IsDir():
			if err := filepath.WalkDir(cleanRoot, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if filepath.Base(path) != spec.ManifestFilename {
					return nil
				}
				absolutePath, err := filepath.Abs(path)
				if err != nil {
					return fmt.Errorf("resolve workspace manifest %s: %w", path, err)
				}
				absolutePath = filepath.Clean(absolutePath)
				if _, ok := seen[absolutePath]; ok {
					return nil
				}
				seen[absolutePath] = struct{}{}
				paths = append(paths, absolutePath)
				return nil
			}); err != nil {
				return nil, fmt.Errorf("walk workspace root %s: %w", cleanRoot, err)
			}
		case filepath.Base(cleanRoot) == spec.ManifestFilename:
			absolutePath, err := filepath.Abs(cleanRoot)
			if err != nil {
				return nil, fmt.Errorf("resolve workspace manifest %s: %w", cleanRoot, err)
			}
			absolutePath = filepath.Clean(absolutePath)
			if _, ok := seen[absolutePath]; ok {
				continue
			}
			seen[absolutePath] = struct{}{}
			paths = append(paths, absolutePath)
		default:
			return nil, fmt.Errorf("workspace root %s: expected directory or %s", cleanRoot, spec.ManifestFilename)
		}
	}

	sort.Strings(paths)
	return paths, nil
}
