# Research Report: Config File Import Gaps

**Topic:** Why config files (Dockerfiles, nginx.conf, php.ini, supervisord.conf, env files) aren't processed during project scanning — and what it would take to hook them up.

**Files analyzed:** 12

## Summary

DevArch has **partial infrastructure** for config file handling but **critical gaps** prevent proper processing during project scans. The compose parser captures `build` contexts and stores them in `compose_overrides` JSONB, and `env_file` references are parsed but not materialized. Config file content storage exists (`service_config_files` table, migration 003) but **no automatic ingestion** happens during project scanning — config files must be manually imported via a separate `ImportAllConfigFiles()` call that's **never invoked** in the scan pipeline. Services with standard images (meilisearch, redis, postgres) work fine; services needing custom builds (app, reverb, queue) fail because paths are captured but file contents aren't available.

## Key Findings

### 1. Project Scan Does NOT Import Config Files
- `api/internal/scanner/scanner.go:761-831` — `scanCompose()` only extracts metadata from docker-compose.yml (service names, images, ports). Does NOT read Dockerfiles, nginx.conf, php.ini, or other referenced config files.

### 2. Parser Captures Build Context Path but Not Content
- `api/internal/compose/parser.go:28-42,183-197` — `ComposeService` struct parses `build` directives into the `Overrides` map as opaque key-value pairs. Build context paths stored in `compose_overrides` JSONB, but **no files are read** from those paths.

### 3. Config File Schema Exists But Unused During Scan
- `api/migrations/003_service_config_model.up.sql:1-34` — Tables exist: `service_config_files` (path, content, mode), `service_config_mounts` (FK links), `service_config_versions` (change tracking). **Never populated** during project scanning.

### 4. Manual Import Function Exists But Never Called for Projects
- `api/internal/compose/importer.go:348-427` — `ImportAllConfigFiles()` and `ResolveConfigMountLinks()` exist. They scan a config directory, read file contents, store in DB, and link to mounts. **Only used for services-library import, NOT project scans.**

### 5. Project Import Pipeline Skips Config Files
- `api/internal/project/import.go:43-305` — `ensureStack()` parses compose.yml, resolves bind mount paths, resolves build context paths, imports service templates, creates instances. **Missing:** no config file content ingestion, no Dockerfile storage, no nginx.conf/php.ini reading. `env_file` paths stored in `service_env_files` but file content not read.

### 6. Compose Generation Tries to Materialize Non-Existent Files
- `api/internal/compose/generator.go:158-162,378-463` — `loadConfigMounts()` resolves paths to `.runtime/compose/preview/{category}/{service}/...`. `MaterializeConfigFiles()` writes from DB to disk. **If files never imported → nothing materializes → builds fail.**

### 7. Env File References Parsed But Not Stored
- `api/internal/compose/parser.go:37,166,473-490` — Parser extracts `env_file` paths into `ParsedService.EnvFiles`. Paths stored in `service_env_files` table but **file contents never read or stored**.

### 8. Build Context Path Resolution Fragile
- `api/internal/compose/generator.go:508-532` — `resolveBuildContextPath()` converts relative paths to `/workspace` container-internal paths. If config files aren't mounted into API container or path doesn't exist, builds fail silently.

### 9. Dashboard Shows Config Files But None Exist After Scan
- `dashboard/src/components/services/config-files-panel.tsx:15-17` — UI queries `/api/v1/services/{name}/config-files` which returns empty after scan because no import occurred.

### 10. ConfigMounts Show "Unresolved" State
- `dashboard/src/components/services/editable-config-mounts.tsx:91-99` — All mounts show "unresolved" badge because `ResolveConfigMountLinks()` never runs for projects.

## Patterns

| Pattern | Implementation | File Reference |
|---------|---------------|----------------|
| Selective parsing | Parser extracts known compose keys, rest → `compose_overrides` JSONB | `parser.go:183-197` |
| Path resolution layers | Relative paths resolved at parse (import) and generation (compose output) time | `importer.go:156-185`, `generator.go:534-557` |
| Two-phase import | Services imported, then config files via separate call (never happens for projects) | `importer.go:348-540` |
| Materialization | Config files stored in DB, written to `.runtime/compose/preview/` on demand | `generator.go:425-463` |
| FK-based resolution | `service_config_mounts.config_file_id` links mounts to files after import | `importer.go:456-540` |

## Dependencies

| Dependency | Usage | File |
|------------|-------|------|
| gopkg.in/yaml.v3 | Parse docker-compose.yml structure | `parser.go:12`, `scanner.go:16` |
| filepath (stdlib) | Path resolution and normalization | `importer.go:7-8`, `generator.go:7-8` |
| service_config_files (DB) | Stores config file content, mode, timestamps | `003_service_config_model.up.sql:1-11` |
| service_config_mounts (DB) | Links volume mounts to config files | `003_service_config_model.up.sql:15-23` |
| compose_overrides (JSONB) | Stores build context, non-standard compose keys | `001_categories_services.up.sql:24` |

## Risks & Gaps

| Risk | Severity | Details |
|------|----------|---------|
| Silent failure on build | **High** | Services with custom Dockerfiles fail during `podman-compose up` — build context doesn't exist |
| Missing env file content | **High** | env_file paths stored but files not read; runtime expects `.env` on host |
| Config mount FK always NULL | Medium | `ResolveConfigMountLinks()` never called for projects → `config_file_id` stays NULL |
| Manual fix required | Medium | Users must manually upload config files via dashboard after scan |
| No Dockerfile content | **High** | Build directives reference Dockerfiles that aren't in DB or API container |
| Migration 012 hardcoded fix | Medium | Specific patch for flowstate paths suggests broader problem |

## Recommendations

1. **Integrate Config Import into Project Scan** — Extend `scanner.upsert()` / `project/import.go` to call `ImportAllConfigFiles()` after stack creation. Scan the compose_path parent directory for Dockerfiles, *.conf, *.ini files. Store in `service_config_files` and link via `ResolveConfigMountLinks()`.

2. **Build Context Content Storage** — Parse `build.dockerfile` path during import, read content, store in `service_config_files`. Materialize to `.runtime/compose/preview/{category}/{service}/Dockerfile` at generation time.

3. **Env File Content Ingestion** — Read env_file paths during import, store content in DB. Materialize alongside compose output.

4. **Scanner Hook for Config Detection** — Add `scanConfigFiles()` to scanner that walks `deploy/`, `config/`, `.docker/` directories. Auto-import Dockerfiles, nginx.conf, php.ini, supervisord.conf.

5. **Dashboard Import Button** — Add "Import Config Files" button on project detail page, triggers `/api/v1/projects/{name}/import-config-files`, shows progress and imported file count.

6. **Validation Pass Post-Import** — After import, validate all `compose_overrides` build contexts have corresponding Dockerfiles. Flag services with unresolved config mounts in UI with actionable warnings.
