package projectscan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScanNodeProjectDerivesTemplatesAndComposeServices(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{
  "name": "shop-api",
  "version": "1.2.3",
  "description": "Shop API",
  "main": "server.js",
  "dependencies": {
    "express": "^4.19.0"
  }
}`)
	writeFile(t, filepath.Join(root, "compose.yml"), `services:
  db:
    image: postgres:16
    ports:
      - "5432:5432"
  cache:
    image: redis:7
    depends_on:
      - db
`)

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if got, want := result.ProjectType, "node"; got != want {
		t.Fatalf("ProjectType = %q, want %q", got, want)
	}
	if got, want := result.Framework, "Express"; got != want {
		t.Fatalf("Framework = %q, want %q", got, want)
	}
	if got, want := result.PackageManager, "npm"; got != want {
		t.Fatalf("PackageManager = %q, want %q", got, want)
	}
	if got, want := result.EntryPoint, "server.js"; got != want {
		t.Fatalf("EntryPoint = %q, want %q", got, want)
	}
	if got, want := result.SuggestedTemplates, []string{"node-api", "postgres", "redis"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedTemplates = %v, want %v", got, want)
	}
	if got, want := result.ServiceCount, 2; got != want {
		t.Fatalf("ServiceCount = %d, want %d", got, want)
	}
	if got, want := result.Services[0].Name, "cache"; got != want {
		t.Fatalf("Services[0].Name = %q, want %q", got, want)
	}
}

func TestScanLaravelProjectSuggestsLaravelTemplate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "artisan"), "#!/usr/bin/env php\n")
	writeFile(t, filepath.Join(root, "composer.json"), `{
  "description": "Laravel shop",
  "require": {
    "php": "^8.3",
    "laravel/framework": "^11.0"
  }
}`)

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if got, want := result.ProjectType, "laravel"; got != want {
		t.Fatalf("ProjectType = %q, want %q", got, want)
	}
	if got, want := result.SuggestedTemplates, []string{"laravel-app"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedTemplates = %v, want %v", got, want)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", path, err)
	}
}
