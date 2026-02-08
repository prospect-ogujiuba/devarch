# Phase 7: Export/Import & Bootstrap - Context

**Gathered:** 2026-02-08
**Status:** Ready for planning

<domain>
## Phase Boundary

Users export stacks to devarch.yml (with resolved specifics) for sharing, import with create-update reconciliation, generate devarch.lock for deterministic reproduction, and bootstrap environments with `devarch init` / `devarch doctor`. Wires are NOT included yet (Phase 8 adds them). Secrets are redacted via heuristic (Phase 9 adds encryption).

</domain>

<decisions>
## Implementation Decisions

### YAML Schema Design
- One stack per file (devarch.yml = one stack definition)
- Version field: simple integer (`version: 1`)
- Override representation: merged effective config (self-contained, no template dependency needed to interpret)
- Instance keys: map keyed by instance name (`instances:\n  my-postgres:\n    template: postgresql`)
- Stack metadata in yml: name + description + network_name (enabled state is runtime, not config)
- Reserved `wires: []` optional field for Phase 8 forward-compatibility (avoids version bump)

*Forward: Phase 8 (WIRE-08) populates wires field without schema version change*

### Resolved Specifics Scope
- Image pinning: tag in yml (`image: postgres:16`), digest in lock (`digest: sha256:abc...`)
- Host ports: configured values in yml, resolved (actual bound) values in lock
- Template version: name in yml, version hash in lock for drift detection
- Lock enforcement on apply: warn only (non-blocking), user decides to proceed or refresh

*Forward: Phase 8 (WIRE-01/02) template version hash enables wire validity drift detection*

### Import Reconciliation
- Match strategy: by instance name (stable key between export/import)
- Missing template: fail with clear error ("Template X not found in catalog. Import the template first.")
- Update behavior: overwrite silently (import is intentional, create-update mode per requirements)
- Stack creation: auto-create if stack name doesn't exist (full bootstrap from file)

*Forward: Phase 8 wires also match by instance name, fail-fast on missing template is correct for wire resolution too*

### Secret Redaction
- Placeholder syntax: `${SECRET:VAR_NAME}` (parseable, distinct from shell variables)
- Detection: keyword heuristic matching *PASSWORD*, *SECRET*, *KEY*, *TOKEN* in env var names
- Import handling: leave placeholder as value, user fills via override editor (no special prompts)
- No unredacted export option — always redact, always safe to share

*Forward: Phase 9 (SECR-01) upgrades detection from heuristic to DB-marked secrets; heuristic becomes fallback. Same placeholder format preserved.*

### Lockfile Boundaries
- Location: same directory as devarch.yml, sibling file (devarch.lock)
- Generation: auto-generated on successful apply (reflects last known-good state)
- VCS: recommended to commit (like package-lock.json — reproducibility for teammates)
- Format: mirrors yml structure with resolved values added (digests, actual ports, template version hashes)
- Integrity: SHA256 hash of devarch.yml embedded in lock (detects yml changes without lock refresh)

*Forward: Phase 8 apply with wires regenerates lock — wires become part of resolved state naturally*

### CLI Bootstrap Flow
- `devarch init`: import devarch.yml + pull images + create networks + apply (one command to running environment)
- `devarch doctor`: Pass/Warn/Fail severity model with exit code 1 on any fail
- Doctor checks: runtime running, configured ports available, disk space >1GB, required CLI tools present
- Implementation: bash scripts in scripts/ extending existing devarch CLI (calls API endpoints)

*Forward: Phase 8 adds "template exports check" to doctor. Phase 9 adds "encryption key present" check. Extensible model.*

### Dashboard Export/Import UX
- Export/Import buttons in stack detail page toolbar (alongside existing action buttons)
- Import: file upload picker (select devarch.yml file)
- Export: browser download (same pattern as existing compose download)
- No import preview — immediate import with result summary (consistent with overwrite-silently decision)

### Claude's Discretion
- Exact devarch.yml field naming and nesting beyond documented decisions
- Lock file internal structure details
- Import result summary format in dashboard
- Doctor check implementation details (disk thresholds, port scan method)
- Error message wording for import failures
- API endpoint paths for export/import

</decisions>

<specifics>
## Specific Ideas

- devarch.yml + devarch.lock pattern modeled after package.json + package-lock.json
- Export always safe to share (redacted) — personal backup uses DB backup, not unredacted export
- `devarch init` is the "clone repo, run one command" experience for teammate onboarding
- Lock mirrors yml structure for easy diffing (git diff devarch.lock shows what changed)
- Compose download button already exists — export button follows same blob download pattern

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

## Forward-Compatibility Decisions

Decisions shaped by Phase 8/9 requirements (flagged for traceability):

1. **Reserved `wires: []` field** — avoids version bump when Phase 8 adds wires to export (WIRE-08)
2. **Template version hash in lock** — enables Phase 8 wire drift detection (WIRE-06 diagnostics)
3. **`${SECRET:VAR_NAME}` placeholder** — Phase 9 upgrades detection from heuristic to DB-marked; same format
4. **Keyword heuristic for secrets** — bridge until Phase 9 explicit secret marking; becomes fallback
5. **Lock generated on apply** — Phase 8 wires become part of resolved state without flow change
6. **Fail on missing template** — correct for Phase 8 too (missing template = wires can't resolve)
7. **Diff format reuse** — Phase 6 decided import reconciliation reuses plan diff format

---

*Phase: 07-export-import-bootstrap*
*Context gathered: 2026-02-08*
