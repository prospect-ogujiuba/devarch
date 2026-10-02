// Package testutil builds temporary catalog fixtures.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// WriteFiles writes path→content under dir.
func WriteFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Compose returns a minimal compose file with an optional x-devarch block.
func Compose(service, meta string) string {
	return meta + "services:\n  " + service + ":\n    image: example/" + service + "\n    container_name: " + service + "\nvolumes:\n  " + service + "_data: null\n"
}
