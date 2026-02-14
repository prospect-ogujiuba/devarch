# Implementation Plan

**Request:** Hook config file import into project scan/import pipeline so Dockerfiles, nginx.conf, php.ini, env files are automatically read from disk, stored in DB, and linked to service mounts.
**Discovery Level:** 1 — Quick Scan (infrastructure exists, needs wiring)
**Overall Risk:** medium
**Files affected:** 1 (api/internal/project/import.go)
**Context budget:** 3 tasks (~50%)

## Must-Haves

### Observable Truths
- When `ensureStack()` completes, `service_config_files` contains Dockerfile content for services with build directives
- When `ensureStack()` completes, `service_config_files` contains content for mounted config files (nginx.conf, php.ini, supervisord.conf)
- `service_config_mounts.config_file_id` is resolved (not NULL) for mounts with matching config files
- `env_file` paths referenced in compose.yml have their contents stored in `service_config_files`
- Compose generation finds config files in DB and materializes them to `.runtime/compose/preview/`

### Required Artifacts
- `api/internal/project/import.go` — `importBuildContextFiles()`, `importEnvFiles()`, `importProjectConfigFiles()`, `resolveConfigMountLinks()`, `isBinaryContent()` functions
- `ensureStack()` calls all import functions after template import loop

### Key Links
- `ensureStack()` → `importBuildContextFiles()` (after template import, reads Dockerfiles from build contexts)
- `ensureStack()` → `importEnvFiles()` (reads env_file content)
- `ensureStack()` → `importProjectConfigFiles()` (scans config dirs)
- `ensureStack()` → `resolveConfigMountLinks()` (links config_file_id FKs)
- All import functions → `service_config_files` table (INSERT)

## Tasks

### Task 1: Add config file import functions to project import pipeline
- **File:** `api/internal/project/import.go`
- **Risk:** medium — Changes project import pipeline; parse errors could break existing project scans. Build context parsing adds file I/O which could fail if paths don't exist.
- **Action:**
  1. Add `importBuildContextFiles(tx, composePath, templateIDs)` — reparses compose, extracts build contexts, reads Dockerfiles, stores in `service_config_files`
  2. Add `importEnvFiles(tx, composePath, templateIDs)` — reparses compose, reads `env_file` contents, stores in `service_config_files`
  3. Add `importProjectConfigFiles(tx, composePath, templateIDs)` — scans `{composeDir}/config/{serviceName}/` dirs, reads non-binary files, stores in `service_config_files`
  4. Wire all three into `ensureStack()` after template import loop (after line ~96)
  5. Use `os.Stat()` before reads, skip missing files silently. Use `isBinaryContent()` to filter binaries.
- **Verify:** `go build ./cmd/server` succeeds, new functions exist and are called in `ensureStack()`
- **Done:** After `ensureStack()`, `service_config_files` populated with Dockerfiles, env files, and config files
- **Depends on:** Task 3 (needs `isBinaryContent()`)

### Task 2: Add resolveConfigMountLinks helper for projects
- **File:** `api/internal/project/import.go`
- **Risk:** low — New helper, self-contained SQL, no changes to existing code paths
- **Action:**
  1. Add `resolveConfigMountLinks(tx *sql.Tx) (int, error)` — mirrors `compose/importer.go:456-540`
  2. Query `service_config_mounts` with NULL `config_file_id`
  3. Parse `source_path` to extract service name and relative path
  4. JOIN against `service_config_files` to find matching file
  5. UPDATE `config_file_id` for matches
  6. Call from `ensureStack()` after all config imports
- **Verify:** Function compiles, signature correct
- **Done:** `service_config_mounts.config_file_id` populated for mounts with matching config files
- **Depends on:** Task 1

### Task 3: Add isBinaryContent helper to project package
- **File:** `api/internal/project/import.go`
- **Risk:** low — Pure helper, no side effects
- **Action:**
  1. Copy `isBinaryContent()` from `compose/importer.go:429-454` to `import.go`
  2. Detects binary via `http.DetectContentType()` and null byte scanning
  3. Add `"net/http"` to import block if not present
- **Verify:** `go build ./cmd/server` succeeds
- **Done:** `isBinaryContent()` available in project package
- **Depends on:** none

## Verification Plan

| Check | Command | Covers Tasks |
|-------|---------|-------------|
| Go build | `cd api && go build ./cmd/server` | 1, 2, 3 |
| Go test | `cd api && go test ./internal/project/...` | 1, 2, 3 |
| Manual: scan project | `POST /api/v1/projects/scan` then start flowstate | 1, 2 |
| Manual: verify DB | `SELECT COUNT(*) FROM service_config_files WHERE file_path = 'Dockerfile'` | 1 |
| Manual: verify mounts | `SELECT COUNT(*) FROM service_config_mounts WHERE config_file_id IS NOT NULL` | 2 |
| Must-haves | Code inspection: all 5 truths verified | all |

## Design Decisions

1. **Reuse existing import logic** — Don't rewrite `ImportAllConfigFiles()`, create project-specific wrappers that scan the project compose directory
2. **Three-phase import** — Build contexts (Dockerfiles) → env files → config dirs. Each independent.
3. **Transaction safety** — All imports within existing `ensureStack()` transaction, failures roll back cleanly
4. **Silent fallback** — Missing config dirs skipped (not all services have config files)
5. **Binary filtering** — Reuse `isBinaryContent()` to avoid storing binaries in TEXT columns

---

```xml
<plan>
  <objective>Hook config file import into project import pipeline</objective>
  <discovery_level>1</discovery_level>
  <overall_risk>medium</overall_risk>

  <must_haves>
    <truths>
      <truth>ensureStack() populates service_config_files with Dockerfiles for build-context services</truth>
      <truth>ensureStack() stores env_file contents in service_config_files</truth>
      <truth>ensureStack() stores mounted config files (nginx.conf, php.ini) in service_config_files</truth>
      <truth>service_config_mounts.config_file_id resolved for matching files</truth>
      <truth>Compose generation materializes stored config files to .runtime/</truth>
    </truths>
    <artifacts>
      <artifact path="api/internal/project/import.go" provides="importBuildContextFiles, importEnvFiles, importProjectConfigFiles, resolveConfigMountLinks, isBinaryContent"/>
    </artifacts>
    <key_links>
      <link from="ensureStack()" to="importBuildContextFiles()" via="function call after template import"/>
      <link from="ensureStack()" to="importEnvFiles()" via="function call"/>
      <link from="ensureStack()" to="importProjectConfigFiles()" via="function call"/>
      <link from="ensureStack()" to="resolveConfigMountLinks()" via="function call after all imports"/>
    </key_links>
  </must_haves>

  <tasks>
    <task type="auto">
      <name>Add isBinaryContent helper to project package</name>
      <files>api/internal/project/import.go</files>
      <action>Copy isBinaryContent() from compose/importer.go:429-454. Add net/http import.</action>
      <verify>go build ./cmd/server</verify>
      <done>isBinaryContent() available in project package</done>
      <risk level="low">Pure helper, no side effects</risk>
      <needs>compose/importer.go reference</needs>
      <creates>isBinaryContent() in project/import.go</creates>
    </task>

    <task type="auto">
      <name>Add config file import functions and wire into ensureStack</name>
      <files>api/internal/project/import.go</files>
      <action>Add importBuildContextFiles, importEnvFiles, importProjectConfigFiles. Wire into ensureStack after template import loop. Each reads files from disk relative to composePath, stores in service_config_files via INSERT ON CONFLICT.</action>
      <verify>go build ./cmd/server; functions called in ensureStack</verify>
      <done>service_config_files populated after ensureStack completes</done>
      <risk level="medium">Changes import pipeline, file I/O could fail</risk>
      <needs>isBinaryContent(), parser.go ParseFileAll()</needs>
      <creates>importBuildContextFiles(), importEnvFiles(), importProjectConfigFiles()</creates>
    </task>

    <task type="auto">
      <name>Add resolveConfigMountLinks helper</name>
      <files>api/internal/project/import.go</files>
      <action>Mirror compose/importer.go:456-540 logic. Query NULL config_file_id mounts, resolve via JOIN, UPDATE FK. Call from ensureStack after imports.</action>
      <verify>go build ./cmd/server; function signature correct</verify>
      <done>config_file_id populated for matching mounts</done>
      <risk level="low">Self-contained SQL helper</risk>
      <needs>Task 2 (config files must exist first)</needs>
      <creates>resolveConfigMountLinks()</creates>
    </task>
  </tasks>
</plan>
```
