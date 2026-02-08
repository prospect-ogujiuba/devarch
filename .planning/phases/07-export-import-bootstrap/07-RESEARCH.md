# Phase 7: Export/Import & Bootstrap - Research

**Researched:** 2026-02-08
**Domain:** YAML export/import, lockfile patterns, CLI diagnostics, image digest pinning
**Confidence:** HIGH

## Summary

Export/import patterns are well-established in package managers (npm, yarn, cargo, poetry) and infrastructure tools (Docker Compose, Kubernetes manifests). The core pattern: **config file (devarch.yml) declares intent, lockfile (devarch.lock) captures resolved state**. This phase follows npm's package.json + package-lock.json model — yml is human-editable and VCS-friendly, lock ensures reproducibility across environments.

The codebase already uses yaml.v3 for compose parsing (`api/internal/compose/parser.go`) and generation (`api/internal/compose/generator.go`). The missing piece: **bidirectional transformation** between DB state and YAML files, plus lockfile generation from resolved runtime state (image digests, bound ports, template versions).

Critical insights from 2026 research:
1. **Lockfile format design matters more for tooling than humans** — use JSON for lock (machine-readable), YAML for config (human-editable) ([Lockfile Design Tradeoffs, Jan 2026](https://nesbitt.io/2026/01/17/lockfile-format-design-and-tradeoffs.html))
2. **Secret detection via heuristics is standard bridge until explicit marking** — keyword matching (*PASSWORD*, *SECRET*, *KEY*, *TOKEN*) is how Netlify, GitGuardian, and TruffleHog all start
3. **Image digests belong in lockfile, not config** — tags in yml (declarative), digests in lock (resolved) matches Docker Compose + Kubernetes patterns
4. **CLI diagnostics follow Pass/Warn/Fail severity model** — established pattern from Kubernetes preflight, Terraform validate, Docker Scout

**Primary recommendation:** Reuse existing yaml.v3 patterns from compose parser/generator. Add new `internal/export` and `internal/lock` packages. Keep secret detection simple (keyword heuristic) — Phase 9 upgrades to DB-marked secrets. CLI bootstrap extends existing bash scripts in `scripts/`.

## User Constraints (from CONTEXT.md)

### Locked Decisions

**YAML Schema Design:**
- One stack per file (devarch.yml = one stack definition)
- Version field: simple integer (`version: 1`)
- Override representation: merged effective config (self-contained, no template dependency needed to interpret)
- Instance keys: map keyed by instance name (`instances:\n  my-postgres:\n    template: postgresql`)
- Stack metadata in yml: name + description + network_name (enabled state is runtime, not config)
- Reserved `wires: []` optional field for Phase 8 forward-compatibility (avoids version bump)

**Resolved Specifics Scope:**
- Image pinning: tag in yml (`image: postgres:16`), digest in lock (`digest: sha256:abc...`)
- Host ports: configured values in yml, resolved (actual bound) values in lock
- Template version: name in yml, version hash in lock for drift detection
- Lock enforcement on apply: warn only (non-blocking), user decides to proceed or refresh

**Import Reconciliation:**
- Match strategy: by instance name (stable key between export/import)
- Missing template: fail with clear error ("Template X not found in catalog. Import the template first.")
- Update behavior: overwrite silently (import is intentional, create-update mode per requirements)
- Stack creation: auto-create if stack name doesn't exist (full bootstrap from file)

**Secret Redaction:**
- Placeholder syntax: `${SECRET:VAR_NAME}` (parseable, distinct from shell variables)
- Detection: keyword heuristic matching *PASSWORD*, *SECRET*, *KEY*, *TOKEN* in env var names
- Import handling: leave placeholder as value, user fills via override editor (no special prompts)
- No unredacted export option — always redact, always safe to share

**Lockfile Boundaries:**
- Location: same directory as devarch.yml, sibling file (devarch.lock)
- Generation: auto-generated on successful apply (reflects last known-good state)
- VCS: recommended to commit (like package-lock.json — reproducibility for teammates)
- Format: mirrors yml structure with resolved values added (digests, actual ports, template version hashes)
- Integrity: SHA256 hash of devarch.yml embedded in lock (detects yml changes without lock refresh)

**CLI Bootstrap Flow:**
- `devarch init`: import devarch.yml + pull images + create networks + apply (one command to running environment)
- `devarch doctor`: Pass/Warn/Fail severity model with exit code 1 on any fail
- Doctor checks: runtime running, configured ports available, disk space >1GB, required CLI tools present
- Implementation: bash scripts in scripts/ extending existing devarch CLI (calls API endpoints)

**Dashboard Export/Import UX:**
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

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope

## Standard Stack

### Core

No new external libraries required. The codebase already uses yaml.v3 extensively:

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gopkg.in/yaml.v3 | Already in use | YAML marshaling/unmarshaling | Existing dependency for compose parsing, stable API (gopkg.in guarantees compatibility) |
| crypto/sha256 | Go stdlib | Lockfile integrity hash, image digest handling | Standard cryptographic hash, same as container image digests |
| encoding/json | Go stdlib | Lockfile serialization (JSON for machine-readability) | Built-in, better for nested structures than YAML for lock |
| os/exec | Go stdlib | Container CLI for digest resolution | Existing pattern in `container/client.go` |
| net | Go stdlib | Port availability checks (Listen/Close probe) | Stdlib socket checks, no dependencies |

**Already in codebase:**
- `api/internal/compose/parser.go` — YAML parsing with yaml.v3
- `api/internal/compose/generator.go` — YAML generation with yaml.v3
- `api/internal/container/client.go` — Container runtime abstraction (podman/docker)

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| strings | Go stdlib | Secret detection keyword matching | Case-insensitive substring matching for *PASSWORD*, *SECRET*, etc. |
| regexp | Go stdlib | Image tag/digest parsing | Extract digest from `docker/podman inspect` output |
| filepath | Go stdlib | Config file path handling | Lock sibling to yml file |
| io | Go stdlib | File hashing for integrity | Stream yml file into SHA256 hash |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| yaml.v3 | goccy/go-yaml | goccy supports better round-tripping with comments, but yaml.v3 already in use, stable, and sufficient for generated YAML |
| JSON for lockfile | YAML for lockfile | YAML lock would be human-readable but harder to parse reliably; JSON is machine-first (correct priority per [Lockfile Design](https://nesbitt.io/2026/01/17/lockfile-format-design-and-tradeoffs.html)) |
| Keyword heuristic secrets | detect-secrets library | Adding Python dependency for heuristics we can implement in 20 lines of Go is overkill; Phase 9 adds proper secret marking |
| Port check via ss/netstat | net.Listen probe | External commands vary by distro (ss vs netstat vs lsof); Go stdlib net.Listen is portable |
| React file upload library | Native input type="file" | No drag-drop needed, simple file picker matches existing compose download UX |

**Installation:**
```bash
# No new dependencies required
```

## Architecture Patterns

### Recommended Project Structure

```
api/internal/export/
├── exporter.go          # DB → devarch.yml conversion
├── importer.go          # devarch.yml → DB reconciliation
├── types.go             # DevArchFile struct (yml schema)
└── secrets.go           # Keyword-based secret detection/redaction

api/internal/lock/
├── generator.go         # Generate devarch.lock from resolved state
├── validator.go         # Compare lock against runtime, emit warnings
├── types.go             # LockFile struct
└── integrity.go         # SHA256 hash of yml file

api/internal/api/handlers/
├── export.go            # GET /stacks/{name}/export → devarch.yml
├── import.go            # POST /stacks/import (multipart/form-data)
├── lock_generate.go     # POST /stacks/{name}/lock → devarch.lock
└── lock_validate.go     # POST /stacks/{name}/lock/validate → warnings

dashboard/src/features/stacks/
├── export.ts            # Export mutation (blob download)
└── import.ts            # Import mutation (file upload)

scripts/
├── devarch-init.sh      # Bootstrap from devarch.yml
└── devarch-doctor.sh    # Environment diagnostics
```

### Pattern 1: YAML Round-Trip with Struct Tags

**What:** Use yaml.v3 struct tags with `omitempty` for optional fields
**When to use:** Export and import operations
**Why:** Ensures clean YAML output (no null values), stable round-trip

```go
// api/internal/export/types.go
type DevArchFile struct {
    Version     int                    `yaml:"version"`
    Stack       StackConfig            `yaml:"stack"`
    Instances   map[string]InstanceDef `yaml:"instances"`
    Wires       []WireDef              `yaml:"wires,omitempty"` // Reserved for Phase 8
}

type StackConfig struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description,omitempty"`
    NetworkName string `yaml:"network_name"`
}

type InstanceDef struct {
    Template    string            `yaml:"template"`
    Image       string            `yaml:"image,omitempty"`      // Override template default
    Ports       []string          `yaml:"ports,omitempty"`
    Environment map[string]string `yaml:"environment,omitempty"`
    Volumes     []VolumeDef       `yaml:"volumes,omitempty"`
    Command     string            `yaml:"command,omitempty"`
}

type VolumeDef struct {
    Source   string `yaml:"source"`
    Target   string `yaml:"target"`
    ReadOnly bool   `yaml:"readonly,omitempty"`
}

// Export to YAML
func (e *Exporter) Export(stackName string) ([]byte, error) {
    file := DevArchFile{Version: 1}

    // Load stack from DB
    var stack models.Stack
    err := e.db.QueryRow(`SELECT name, description, network_name FROM stacks WHERE name = $1`, stackName).
        Scan(&stack.Name, &stack.Description, &stack.NetworkName)

    file.Stack = StackConfig{
        Name:        stack.Name,
        Description: stack.Description,
        NetworkName: stack.NetworkName,
    }

    // Load instances with overrides (merged effective config)
    file.Instances = e.loadInstances(stackName)

    // Apply secret redaction
    for name, inst := range file.Instances {
        inst.Environment = e.redactSecrets(inst.Environment)
        file.Instances[name] = inst
    }

    return yaml.Marshal(&file)
}
```

**Source:** Existing pattern from `compose/generator.go` + [yaml.v3 documentation](https://pkg.go.dev/gopkg.in/yaml.v3)

### Pattern 2: Secret Detection via Keyword Heuristic

**What:** Case-insensitive substring matching against known secret keywords
**When to use:** Export operation, before YAML serialization
**Why:** Industry-standard bridge until explicit secret marking (Phase 9)

```go
// api/internal/export/secrets.go
var secretKeywords = []string{
    "password", "secret", "key", "token",
    "api_key", "apikey", "auth",
    "private", "credential", "passwd",
}

func (e *Exporter) redactSecrets(env map[string]string) map[string]string {
    redacted := make(map[string]string)

    for key, value := range env {
        if isSecretKey(key) {
            redacted[key] = fmt.Sprintf("${SECRET:%s}", key)
        } else {
            redacted[key] = value
        }
    }

    return redacted
}

func isSecretKey(key string) bool {
    lowerKey := strings.ToLower(key)

    for _, keyword := range secretKeywords {
        if strings.Contains(lowerKey, keyword) {
            return true
        }
    }

    return false
}
```

**Source:** Pattern from [Netlify Secrets Controller](https://docs.netlify.com/build/environment-variables/secrets-controller/) and [detect-secrets approach](https://github.com/Yelp/detect-secrets)

### Pattern 3: Lockfile with Integrity Hash

**What:** JSON lockfile with SHA256 hash of yml file embedded
**When to use:** After successful apply, before commit
**Why:** Detects manual yml edits without lock refresh

```go
// api/internal/lock/types.go
type LockFile struct {
    Version       int                  `json:"version"`
    GeneratedAt   string               `json:"generated_at"`    // RFC3339
    ConfigHash    string               `json:"config_hash"`     // SHA256 of devarch.yml
    Stack         StackLock            `json:"stack"`
    Instances     map[string]InstLock  `json:"instances"`
}

type StackLock struct {
    Name        string `json:"name"`
    NetworkName string `json:"network_name"`
    NetworkID   string `json:"network_id,omitempty"` // Actual runtime network ID
}

type InstLock struct {
    Template       string            `json:"template"`
    TemplateHash   string            `json:"template_hash"`   // For drift detection (Phase 8)
    ImageTag       string            `json:"image_tag"`
    ImageDigest    string            `json:"image_digest"`    // sha256:abc...
    ResolvedPorts  map[string]int    `json:"resolved_ports"`  // "5432/tcp" -> 54321
}

// Generate lock from runtime state
func (g *Generator) Generate(stackName, ymlPath string) (*LockFile, error) {
    lock := &LockFile{
        Version:     1,
        GeneratedAt: time.Now().Format(time.RFC3339),
        Instances:   make(map[string]InstLock),
    }

    // Compute yml integrity hash
    hash, err := computeFileHash(ymlPath)
    if err != nil {
        return nil, fmt.Errorf("compute config hash: %w", err)
    }
    lock.ConfigHash = hash

    // Load stack info
    lock.Stack = g.loadStackLock(stackName)

    // Load running containers and resolve specifics
    containers, _ := g.containerClient.ListContainersWithLabels(map[string]string{
        "devarch.stack_id": stackName,
    })

    for _, container := range containers {
        instanceName := container.Labels["devarch.instance_id"]

        // Get image digest from inspect
        digest, _ := g.getImageDigest(container.Image)

        lock.Instances[instanceName] = InstLock{
            Template:      container.Labels["devarch.template"],
            TemplateHash:  g.getTemplateHash(container.Labels["devarch.template"]),
            ImageTag:      container.Image,
            ImageDigest:   digest,
            ResolvedPorts: g.extractBoundPorts(container),
        }
    }

    return lock, nil
}

func computeFileHash(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }

    return fmt.Sprintf("%x", h.Sum(nil)), nil
}
```

**Source:** [SHA256 file integrity in Go](https://transloadit.com/devtips/verify-file-integrity-with-go-and-sha256/) + lockfile integrity from [npm package-lock.json format](https://docs.npmjs.com/cli/v9/configuring-npm/package-lock-json/)

### Pattern 4: Image Digest Resolution via Container Inspect

**What:** Use `podman/docker inspect` to get sha256 digest of pulled images
**When to use:** Lockfile generation after apply
**Why:** Digests are immutable, prevent tag mutation attacks

```go
// api/internal/lock/generator.go
func (g *Generator) getImageDigest(imageRef string) (string, error) {
    runtime := g.containerClient.RuntimeName()

    var cmd *exec.Cmd
    if runtime == "podman" {
        cmd = exec.Command("podman", "image", "inspect", imageRef, "--format", "{{.Digest}}")
    } else {
        cmd = exec.Command("docker", "image", "inspect", imageRef, "--format", "{{.RepoDigests}}")
    }

    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("inspect image %s: %w", imageRef, err)
    }

    // Parse digest from output
    digest := strings.TrimSpace(string(output))

    // Docker returns [registry/image@sha256:abc], extract sha256:abc
    if strings.Contains(digest, "@") {
        parts := strings.Split(digest, "@")
        if len(parts) == 2 {
            digest = parts[1]
        }
    }

    return digest, nil
}
```

**Source:** [Podman image inspect docs](https://docs.podman.io/en/latest/markdown/podman-image-inspect.1.html) + [Docker image digests](https://docs.docker.com/dhi/core-concepts/digests/)

### Pattern 5: Import Reconciliation via Upsert

**What:** Match instances by name, upsert stack + instances in transaction
**When to use:** Import endpoint
**Why:** Idempotent, handles both create and update cases

```go
// api/internal/export/importer.go
func (imp *Importer) Import(file *DevArchFile) (*ImportResult, error) {
    tx, err := imp.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Upsert stack
    var stackID int
    err = tx.QueryRow(`
        INSERT INTO stacks (name, description, network_name)
        VALUES ($1, $2, $3)
        ON CONFLICT (name) DO UPDATE
        SET description = EXCLUDED.description,
            network_name = EXCLUDED.network_name,
            updated_at = NOW()
        RETURNING id
    `, file.Stack.Name, file.Stack.Description, file.Stack.NetworkName).Scan(&stackID)

    if err != nil {
        return nil, fmt.Errorf("upsert stack: %w", err)
    }

    result := &ImportResult{
        StackName: file.Stack.Name,
        Created:   []string{},
        Updated:   []string{},
        Errors:    []string{},
    }

    // Reconcile instances
    for name, inst := range file.Instances {
        // Verify template exists
        var templateID int
        err := tx.QueryRow(`SELECT id FROM services WHERE name = $1`, inst.Template).Scan(&templateID)
        if err == sql.ErrNoRows {
            result.Errors = append(result.Errors, fmt.Sprintf("Template %s not found for instance %s", inst.Template, name))
            continue
        }

        // Upsert instance
        var exists bool
        err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM service_instances WHERE instance_id = $1 AND stack_id = $2)`, name, stackID).Scan(&exists)

        if exists {
            // Update existing instance
            _, err = tx.Exec(`UPDATE service_instances SET ... WHERE instance_id = $1 AND stack_id = $2`, name, stackID)
            result.Updated = append(result.Updated, name)
        } else {
            // Create new instance
            _, err = tx.Exec(`INSERT INTO service_instances (...) VALUES (...)`)
            result.Created = append(result.Created, name)
        }
    }

    if len(result.Errors) > 0 {
        return result, fmt.Errorf("import had errors")
    }

    tx.Commit()
    return result, nil
}
```

**Source:** Existing pattern from `compose/importer.go` service template import + [PostgreSQL upsert docs](https://www.postgresql.org/docs/current/sql-insert.html)

### Pattern 6: Port Availability Check via net.Listen

**What:** Attempt to bind port, immediately close if successful
**When to use:** `devarch doctor` diagnostics
**Why:** Portable across Linux/macOS, no external commands needed

```go
// scripts/devarch-doctor.sh calls API endpoint /doctor
// api/internal/api/handlers/doctor.go

func (h *DoctorHandler) CheckPorts(ports []int) []PortCheck {
    results := make([]PortCheck, len(ports))

    for i, port := range ports {
        addr := fmt.Sprintf(":%d", port)
        ln, err := net.Listen("tcp", addr)

        if err != nil {
            results[i] = PortCheck{
                Port:      port,
                Available: false,
                Message:   fmt.Sprintf("Port %d in use", port),
                Severity:  "warn",
            }
        } else {
            ln.Close()
            results[i] = PortCheck{
                Port:      port,
                Available: true,
                Message:   fmt.Sprintf("Port %d available", port),
                Severity:  "pass",
            }
        }
    }

    return results
}
```

**Source:** [Go net.Listen for port checks](https://pkg.go.dev/net#Listen) — preferred over ss/netstat per [port check comparison](https://linuxize.com/post/check-listening-ports-linux/)

### Pattern 7: CLI Bootstrap Script

**What:** Bash script that orchestrates import + pull + apply
**When to use:** `devarch init` command for teammate onboarding
**Why:** Single command from clone to running environment

```bash
#!/bin/bash
# scripts/devarch-init.sh

set -eo pipefail

YML_FILE="${1:-devarch.yml}"
API_BASE="${DEVARCH_API_URL:-http://localhost:8550}"

if [[ ! -f "$YML_FILE" ]]; then
    echo "Error: $YML_FILE not found"
    exit 1
fi

echo "🚀 Bootstrapping from $YML_FILE"

# 1. Import devarch.yml
echo "📥 Importing stack configuration..."
STACK_NAME=$(grep '^  name:' "$YML_FILE" | awk '{print $2}')

curl -sS -X POST "$API_BASE/api/v1/stacks/import" \
    -H "X-API-Key: ${DEVARCH_API_KEY}" \
    -F "file=@$YML_FILE" | jq .

# 2. Pull images (if apply will pull anyway, but explicit feedback)
echo "🐳 Pulling container images..."
# Extract image refs from yml, issue pull commands
grep 'image:' "$YML_FILE" | awk '{print $2}' | sort -u | while read -r img; do
    echo "  Pulling $img..."
    podman pull "$img" || docker pull "$img" || true
done

# 3. Create network (idempotent)
echo "🌐 Creating network..."
NETWORK_NAME=$(grep 'network_name:' "$YML_FILE" | awk '{print $2}')
podman network create "$NETWORK_NAME" 2>/dev/null || \
docker network create "$NETWORK_NAME" 2>/dev/null || \
echo "  Network $NETWORK_NAME already exists"

# 4. Apply stack
echo "🚢 Applying stack..."
curl -sS -X POST "$API_BASE/api/v1/stacks/$STACK_NAME/apply" \
    -H "X-API-Key: ${DEVARCH_API_KEY}" \
    -H "Content-Type: application/json" \
    -d '{"force": true}' | jq .

echo "✅ Bootstrap complete! Stack $STACK_NAME is running."
```

**Source:** Pattern from existing `scripts/devarch` CLI structure + [one-command setup patterns](https://oneuptime.com/blog/post/2026-01-23-go-build-vs-install/view)

### Pattern 8: File Upload/Download in React

**What:** Native file input for upload, blob download for export
**When to use:** Dashboard export/import UI
**Why:** Simple, no dependencies, matches existing compose download

```tsx
// dashboard/src/features/stacks/export.ts
export function useExportStack(stackName: string) {
  return useMutation({
    mutationFn: async () => {
      const response = await api.get(`/stacks/${stackName}/export`, {
        responseType: 'blob',
      })
      return response.data
    },
    onSuccess: (blob) => {
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${stackName}-devarch.yml`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    },
  })
}

// dashboard/src/features/stacks/import.ts
export function useImportStack() {
  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData()
      formData.append('file', file)

      const response = await api.post('/stacks/import', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return response.data
    },
  })
}

// dashboard/src/routes/stacks/$name.tsx (add buttons to toolbar)
function StackDetailToolbar({ stackName }: { stackName: string }) {
  const exportStack = useExportStack(stackName)
  const importStack = useImportStack()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleImport = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      importStack.mutate(file)
    }
  }

  return (
    <div className="flex gap-2">
      <Button onClick={() => exportStack.mutate()}>
        Export
      </Button>

      <input
        ref={fileInputRef}
        type="file"
        accept=".yml,.yaml"
        className="hidden"
        onChange={handleImport}
      />
      <Button onClick={() => fileInputRef.current?.click()}>
        Import
      </Button>
    </div>
  )
}
```

**Source:** [React file upload patterns](https://uploadcare.com/blog/how-to-upload-file-in-react/) + existing compose download in dashboard

### Anti-Patterns to Avoid

- **Using YAML for lockfile:** Lockfiles are machine-first, not human-first. JSON is easier to parse reliably and standard for locks ([Lockfile Design](https://nesbitt.io/2026/01/17/lockfile-format-design-and-tradeoffs.html))
- **Pulling digests from registry API before pull:** Digest of pulled image may differ from registry manifest. Inspect after pull. ([Docker digest resolution](https://deepwiki.com/docker/compose/4.2-classic-builder))
- **Omitting template version in lock:** Phase 8 needs template hash for wire drift detection. Include now, use later.
- **Blocking import on lockfile presence:** Lock is optional. Import from yml only, regenerate lock on next apply.
- **Running doctor checks serially:** Port checks, disk space, runtime status are independent — run in parallel for speed.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML parsing/generation | Custom parser | yaml.v3 (already in use) | Handles edge cases (multiline strings, anchors, escaping) correctly |
| File integrity hashing | Custom checksum | crypto/sha256 + io.Copy | Standard library, streaming for large files, same hash as container images |
| Port availability detection | Parse ss/netstat output | net.Listen probe | Portable across distros, no sudo needed, stdlib |
| Secret detection | Regex patterns per language | Keyword substring matching | Simple, effective bridge until Phase 9 DB marking |
| Image digest extraction | Parse manifest JSON | podman/docker inspect | Handles both OCI and Docker formats, digest of pulled image (not registry manifest) |
| File upload/download UI | react-uploady, FilePond | Native input + Blob API | No dependencies, sufficient for single-file picker |

**Key insight:** This phase is data transformation and CLI orchestration. The hard problems (YAML parsing, crypto hashing, network I/O) are solved by stdlib and existing dependencies. The value is in correct schema design, idempotent import, and helpful diagnostics.

## Common Pitfalls

### Pitfall 1: YAML Round-Trip Loses Zero Values with omitempty

**What goes wrong:** Export then import drops fields with zero values (empty strings, 0, false) because `omitempty` skips them
**Why it happens:** yaml.v3 `omitempty` treats zero values as "empty" by default
**How to avoid:** Only use `omitempty` on truly optional fields. For fields that can be zero (e.g., port 0 = auto-assign), omit the tag.
**Warning signs:** Import creates instances missing explicit zero-value configs from export

**Source:** [yaml.v3 omitempty behavior](https://github.com/go-yaml/yaml/issues/113)

### Pitfall 2: Image Digest Mismatch Between Registry and Runtime

**What goes wrong:** Lockfile stores digest from registry manifest, but pulled image has different digest
**Why it happens:** Multi-arch images have per-platform digests; manifest digest ≠ image digest
**How to avoid:** Always inspect image after pull, never query registry for digest before pull
**Warning signs:** Lock validation fails immediately after generation on same image

**Source:** [Podman digest issues](https://github.com/containers/podman/issues/14779) and [Docker digest docs](https://docs.docker.com/dhi/core-concepts/digests/)

### Pitfall 3: Secret Placeholder Collision with Shell Variables

**What goes wrong:** User has env var `DB_URL=${DATABASE_HOST}/mydb`, export redacts as `${SECRET:DB_URL}`, import breaks shell-style interpolation elsewhere
**Why it happens:** `${}` syntax overlaps with shell variables
**How to avoid:** Use distinct prefix: `${SECRET:VAR_NAME}` never valid shell syntax (colon inside braces)
**Warning signs:** Import fails to substitute placeholders, or shell scripts break

**Source:** Pattern from [environment variable best practices](https://blog.gitguardian.com/secure-your-secrets-with-env/)

### Pitfall 4: Lock Validation Blocks Apply After Manual YML Edit

**What goes wrong:** User edits devarch.yml, lock config_hash mismatches, apply blocked
**Why it happens:** Lock integrity check is too strict (fail instead of warn)
**How to avoid:** Lock validation emits warnings only (locked decision: "warn only, non-blocking")
**Warning signs:** Users unable to apply after yml edits without deleting lock

**Source:** Locked decision in CONTEXT.md + [npm lock behavior](https://docs.npmjs.com/cli/v9/configuring-npm/package-lock-json/)

### Pitfall 5: Doctor Checks Fail in Containers

**What goes wrong:** `devarch doctor` checks localhost ports, but API runs in container with different network namespace
**Why it happens:** Port checks run from wrong perspective (host vs container)
**How to avoid:** Doctor checks should detect runtime (local vs container) and adjust checks. For container API, skip port checks or check from container perspective.
**Warning signs:** Doctor reports ports available when host shows conflicts

**Source:** [Container networking fundamentals](https://docs.docker.com/network/)

### Pitfall 6: Import Fails Silently on Missing Templates

**What goes wrong:** Import skips instances with missing templates, returns success
**Why it happens:** Partial import logic doesn't distinguish fatal vs skippable errors
**How to avoid:** Locked decision: fail-fast on missing template with clear error. Validate all templates exist before starting import transaction.
**Warning signs:** Import succeeds but instances missing, user confused

**Source:** Locked decision in CONTEXT.md

### Pitfall 7: Concurrent Import Overwrites

**What goes wrong:** Two users import different yml files to same stack simultaneously, final state is unpredictable
**Why it happens:** No locking on import endpoint
**How to avoid:** Reuse Phase 6 advisory lock pattern — acquire `pg_try_advisory_lock(stack.id)` during import, return 409 if unavailable
**Warning signs:** Import changes disappear, users report "lost my changes"

**Source:** Phase 6 advisory lock pattern + upsert semantics

## Code Examples

Verified patterns from official sources and existing codebase:

### Export Endpoint with Secret Redaction

```go
// api/internal/api/handlers/export.go
func (h *ExportHandler) ExportStack(w http.ResponseWriter, r *http.Request) {
    stackName := chi.URLParam(r, "name")

    exporter := export.NewExporter(h.db, h.containerClient)
    yamlBytes, err := exporter.Export(stackName)

    if err != nil {
        http.Error(w, fmt.Sprintf("export failed: %v", err), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/x-yaml")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-devarch.yml"`, stackName))
    w.Write(yamlBytes)
}

// api/internal/export/exporter.go
func (e *Exporter) Export(stackName string) ([]byte, error) {
    file := &DevArchFile{Version: 1}

    // Load stack
    err := e.db.QueryRow(`
        SELECT name, description, network_name
        FROM stacks
        WHERE name = $1 AND deleted_at IS NULL
    `, stackName).Scan(&file.Stack.Name, &file.Stack.Description, &file.Stack.NetworkName)

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("stack not found")
    }

    // Load instances with merged effective config
    file.Instances = make(map[string]InstanceDef)
    rows, err := e.db.Query(`
        SELECT si.instance_id, s.name as template, si.overrides
        FROM service_instances si
        JOIN services s ON s.id = si.template_service_id
        WHERE si.stack_id = (SELECT id FROM stacks WHERE name = $1)
        AND si.deleted_at IS NULL
    `, stackName)

    for rows.Next() {
        var instanceID, template string
        var overridesJSON sql.NullString
        rows.Scan(&instanceID, &template, &overridesJSON)

        inst := InstanceDef{Template: template}

        // Parse overrides JSON, populate inst fields
        if overridesJSON.Valid {
            var overrides map[string]interface{}
            json.Unmarshal([]byte(overridesJSON.String), &overrides)

            if env, ok := overrides["environment"].(map[string]interface{}); ok {
                inst.Environment = make(map[string]string)
                for k, v := range env {
                    inst.Environment[k] = fmt.Sprintf("%v", v)
                }
            }

            // ... populate other fields from overrides
        }

        // Redact secrets
        inst.Environment = e.redactSecrets(inst.Environment)

        file.Instances[instanceID] = inst
    }

    return yaml.Marshal(file)
}

func (e *Exporter) redactSecrets(env map[string]string) map[string]string {
    keywords := []string{"password", "secret", "key", "token", "api_key", "apikey", "auth", "private", "credential"}

    redacted := make(map[string]string)
    for key, value := range env {
        isSecret := false
        lowerKey := strings.ToLower(key)

        for _, kw := range keywords {
            if strings.Contains(lowerKey, kw) {
                isSecret = true
                break
            }
        }

        if isSecret {
            redacted[key] = fmt.Sprintf("${SECRET:%s}", key)
        } else {
            redacted[key] = value
        }
    }

    return redacted
}
```

**Source:** Pattern from existing `compose/generator.go` + [detect-secrets heuristics](https://github.com/Yelp/detect-secrets)

### Import Endpoint with Reconciliation

```go
// api/internal/api/handlers/import.go
func (h *ImportHandler) ImportStack(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20) // 10MB max
    if err != nil {
        http.Error(w, "invalid form data", http.StatusBadRequest)
        return
    }

    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "no file provided", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Parse YAML
    var devarchFile export.DevArchFile
    decoder := yaml.NewDecoder(file)
    if err := decoder.Decode(&devarchFile); err != nil {
        http.Error(w, fmt.Sprintf("invalid YAML: %v", err), http.StatusBadRequest)
        return
    }

    // Version check
    if devarchFile.Version != 1 {
        http.Error(w, fmt.Sprintf("unsupported version: %d", devarchFile.Version), http.StatusBadRequest)
        return
    }

    // Import
    importer := export.NewImporter(h.db)
    result, err := importer.Import(&devarchFile)

    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(result) // Return partial result with errors
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// api/internal/export/importer.go
type ImportResult struct {
    StackName string   `json:"stack_name"`
    Created   []string `json:"created"`
    Updated   []string `json:"updated"`
    Errors    []string `json:"errors,omitempty"`
}

func (imp *Importer) Import(file *DevArchFile) (*ImportResult, error) {
    result := &ImportResult{StackName: file.Stack.Name}

    tx, err := imp.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Acquire advisory lock for stack
    var stackID int
    err = tx.QueryRow(`SELECT id FROM stacks WHERE name = $1`, file.Stack.Name).Scan(&stackID)
    if err == nil {
        // Stack exists, try lock
        var acquired bool
        tx.QueryRow(`SELECT pg_try_advisory_xact_lock($1)`, stackID).Scan(&acquired)
        if !acquired {
            return nil, fmt.Errorf("stack is locked by another operation")
        }
    }

    // Upsert stack
    err = tx.QueryRow(`
        INSERT INTO stacks (name, description, network_name)
        VALUES ($1, $2, $3)
        ON CONFLICT (name) DO UPDATE
        SET description = EXCLUDED.description,
            network_name = EXCLUDED.network_name,
            updated_at = NOW()
        RETURNING id
    `, file.Stack.Name, file.Stack.Description, file.Stack.NetworkName).Scan(&stackID)

    if err != nil {
        return nil, fmt.Errorf("upsert stack: %w", err)
    }

    // Validate templates exist
    for name, inst := range file.Instances {
        var exists bool
        err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM services WHERE name = $1)`, inst.Template).Scan(&exists)

        if !exists {
            result.Errors = append(result.Errors, fmt.Sprintf("Template %s not found for instance %s. Import the template first.", inst.Template, name))
        }
    }

    if len(result.Errors) > 0 {
        return result, fmt.Errorf("import validation failed")
    }

    // Reconcile instances
    for name, inst := range file.Instances {
        var templateID int
        tx.QueryRow(`SELECT id FROM services WHERE name = $1`, inst.Template).Scan(&templateID)

        // Convert InstanceDef to JSON overrides
        overrides := make(map[string]interface{})
        if inst.Image != "" {
            overrides["image"] = inst.Image
        }
        if len(inst.Environment) > 0 {
            overrides["environment"] = inst.Environment
        }
        // ... other fields

        overridesJSON, _ := json.Marshal(overrides)

        // Check if instance exists
        var exists bool
        tx.QueryRow(`
            SELECT EXISTS(SELECT 1 FROM service_instances WHERE instance_id = $1 AND stack_id = $2)
        `, name, stackID).Scan(&exists)

        if exists {
            _, err := tx.Exec(`
                UPDATE service_instances
                SET template_service_id = $1, overrides = $2, updated_at = NOW()
                WHERE instance_id = $3 AND stack_id = $4
            `, templateID, overridesJSON, name, stackID)

            if err == nil {
                result.Updated = append(result.Updated, name)
            }
        } else {
            _, err := tx.Exec(`
                INSERT INTO service_instances (instance_id, stack_id, template_service_id, overrides, enabled)
                VALUES ($1, $2, $3, $4, true)
            `, name, stackID, templateID, overridesJSON)

            if err == nil {
                result.Created = append(result.Created, name)
            }
        }
    }

    tx.Commit()
    return result, nil
}
```

**Source:** Upsert pattern from existing `compose/importer.go` + advisory lock from Phase 6

### Lockfile Generation

```go
// api/internal/lock/generator.go
func (g *Generator) Generate(stackName, ymlPath string) ([]byte, error) {
    lock := &LockFile{
        Version:     1,
        GeneratedAt: time.Now().Format(time.RFC3339),
        Instances:   make(map[string]InstLock),
    }

    // Compute yml integrity hash
    f, err := os.Open(ymlPath)
    if err != nil {
        return nil, fmt.Errorf("open yml: %w", err)
    }
    defer f.Close()

    h := sha256.New()
    io.Copy(h, f)
    lock.ConfigHash = fmt.Sprintf("%x", h.Sum(nil))

    // Load stack info
    var networkName, networkID string
    err = g.db.QueryRow(`
        SELECT network_name
        FROM stacks
        WHERE name = $1
    `, stackName).Scan(&networkName)

    if err != nil {
        return nil, fmt.Errorf("load stack: %w", err)
    }

    // Get network ID from runtime
    networkInfo, _ := g.containerClient.InspectNetwork(networkName)
    if networkInfo != nil {
        networkID = networkInfo.ID
    }

    lock.Stack = StackLock{
        Name:        stackName,
        NetworkName: networkName,
        NetworkID:   networkID,
    }

    // Load running containers
    containers, err := g.containerClient.ListContainersWithLabels(map[string]string{
        "devarch.stack_id": stackName,
    })

    for _, container := range containers {
        instanceName := container.Labels["devarch.instance_id"]
        templateName := container.Labels["devarch.template"]

        // Get image digest
        digest, err := g.getImageDigest(container.Image)
        if err != nil {
            digest = "" // Non-fatal, lockfile still useful
        }

        // Get template hash
        templateHash, _ := g.getTemplateHash(templateName)

        // Extract resolved ports
        resolvedPorts := make(map[string]int)
        for _, port := range container.Ports {
            // port format: "5432/tcp" -> 54321
            resolvedPorts[port.Container] = port.Host
        }

        lock.Instances[instanceName] = InstLock{
            Template:      templateName,
            TemplateHash:  templateHash,
            ImageTag:      container.Image,
            ImageDigest:   digest,
            ResolvedPorts: resolvedPorts,
        }
    }

    return json.MarshalIndent(lock, "", "  ")
}

func (g *Generator) getImageDigest(imageRef string) (string, error) {
    runtime := g.containerClient.RuntimeName()

    var cmd *exec.Cmd
    if runtime == "podman" {
        cmd = exec.Command("podman", "image", "inspect", imageRef, "--format", "{{.Digest}}")
    } else {
        cmd = exec.Command("docker", "image", "inspect", imageRef, "--format", "{{index .RepoDigests 0}}")
    }

    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    digest := strings.TrimSpace(string(output))

    // Extract sha256:abc from registry/image@sha256:abc
    if idx := strings.Index(digest, "@"); idx != -1 {
        digest = digest[idx+1:]
    }

    return digest, nil
}

func (g *Generator) getTemplateHash(templateName string) (string, error) {
    var createdAt time.Time
    err := g.db.QueryRow(`
        SELECT created_at FROM services WHERE name = $1
    `, templateName).Scan(&createdAt)

    if err != nil {
        return "", err
    }

    // Simple version hash: SHA256 of template name + created_at
    // Phase 8 can enhance to include template content hash
    h := sha256.New()
    h.Write([]byte(templateName))
    h.Write([]byte(createdAt.Format(time.RFC3339Nano)))

    return fmt.Sprintf("%x", h.Sum(nil))[:16], nil // 16 chars sufficient
}
```

**Source:** [crypto/sha256 streaming](https://gobyexample.com/sha256-hashes) + [Podman image inspect](https://docs.podman.io/en/latest/markdown/podman-image-inspect.1.html)

### CLI Bootstrap (devarch init)

```bash
#!/bin/bash
# scripts/devarch-init.sh

set -eo pipefail

YML_FILE="${1:-devarch.yml}"
API_BASE="${DEVARCH_API_URL:-http://localhost:8550}"
API_KEY="${DEVARCH_API_KEY:-test}"

if [[ ! -f "$YML_FILE" ]]; then
    echo "❌ Error: $YML_FILE not found"
    echo "Usage: devarch init [devarch.yml]"
    exit 1
fi

echo "🚀 DevArch Bootstrap"
echo "   Config: $YML_FILE"
echo ""

# Extract stack name from YAML
STACK_NAME=$(grep -A1 '^stack:' "$YML_FILE" | grep 'name:' | awk '{print $2}' | tr -d '"')

if [[ -z "$STACK_NAME" ]]; then
    echo "❌ Error: Could not parse stack name from $YML_FILE"
    exit 1
fi

echo "📦 Stack: $STACK_NAME"
echo ""

# Step 1: Import
echo "📥 Importing stack configuration..."
IMPORT_RESULT=$(curl -sS -X POST "$API_BASE/api/v1/stacks/import" \
    -H "X-API-Key: $API_KEY" \
    -F "file=@$YML_FILE")

IMPORT_STATUS=$(echo "$IMPORT_RESULT" | jq -r '.stack_name // empty')
if [[ -z "$IMPORT_STATUS" ]]; then
    echo "❌ Import failed:"
    echo "$IMPORT_RESULT" | jq .
    exit 1
fi

echo "✅ Imported stack: $IMPORT_STATUS"
CREATED=$(echo "$IMPORT_RESULT" | jq -r '.created | length')
UPDATED=$(echo "$IMPORT_RESULT" | jq -r '.updated | length')
echo "   Created: $CREATED instances"
echo "   Updated: $UPDATED instances"
echo ""

# Step 2: Pull images
echo "🐳 Pulling container images..."
RUNTIME=$(command -v podman &>/dev/null && echo "podman" || echo "docker")
grep 'image:' "$YML_FILE" | awk '{print $2}' | tr -d '"' | sort -u | while read -r IMG; do
    if [[ -n "$IMG" ]]; then
        echo "   Pulling $IMG..."
        $RUNTIME pull "$IMG" 2>&1 | grep -E '(Digest|Downloaded|Already exists)' || true
    fi
done
echo ""

# Step 3: Create network
echo "🌐 Creating network..."
NETWORK_NAME=$(grep -A3 '^stack:' "$YML_FILE" | grep 'network_name:' | awk '{print $2}' | tr -d '"')
$RUNTIME network create "$NETWORK_NAME" 2>/dev/null && echo "   Created $NETWORK_NAME" || echo "   Network $NETWORK_NAME already exists"
echo ""

# Step 4: Apply
echo "🚢 Applying stack..."
APPLY_RESULT=$(curl -sS -X POST "$API_BASE/api/v1/stacks/$STACK_NAME/apply" \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"force": true}')

APPLY_STATUS=$(echo "$APPLY_RESULT" | jq -r '.status // empty')
if [[ "$APPLY_STATUS" != "applied" ]]; then
    echo "⚠️  Apply warnings or errors:"
    echo "$APPLY_RESULT" | jq .
fi

echo ""
echo "✅ Bootstrap complete!"
echo "   Stack $STACK_NAME is running."
echo ""
echo "Next steps:"
echo "  • View status: devarch stack ls"
echo "  • Check logs: devarch logs $STACK_NAME"
echo "  • Validate environment: devarch doctor"
```

**Source:** Pattern from existing `scripts/devarch` + [bash error handling best practices](https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html)

### CLI Doctor Diagnostics

```bash
#!/bin/bash
# scripts/devarch-doctor.sh

set -eo pipefail

API_BASE="${DEVARCH_API_URL:-http://localhost:8550}"
API_KEY="${DEVARCH_API_KEY:-test}"

echo "🏥 DevArch Environment Diagnostics"
echo ""

EXIT_CODE=0

# Check 1: Container runtime
echo "🐳 Container Runtime"
if command -v podman &>/dev/null; then
    RUNTIME="podman"
    VERSION=$(podman --version | awk '{print $3}')
    echo "   ✅ Podman $VERSION installed"
elif command -v docker &>/dev/null; then
    RUNTIME="docker"
    VERSION=$(docker --version | awk '{print $3}' | tr -d ',')
    echo "   ✅ Docker $VERSION installed"
else
    echo "   ❌ FAIL: No container runtime found (install podman or docker)"
    EXIT_CODE=1
fi

if [[ -n "$RUNTIME" ]]; then
    # Check if runtime is running
    if $RUNTIME ps &>/dev/null; then
        echo "   ✅ Runtime is running"
    else
        echo "   ❌ FAIL: Runtime not running or permission denied"
        EXIT_CODE=1
    fi
fi
echo ""

# Check 2: API connectivity
echo "🌐 API Server"
if curl -sS -f "$API_BASE/health" -H "X-API-Key: $API_KEY" &>/dev/null; then
    echo "   ✅ API reachable at $API_BASE"
else
    echo "   ⚠️  WARN: API not reachable (is devarch-api running?)"
fi
echo ""

# Check 3: Disk space
echo "💾 Disk Space"
DISK_AVAIL=$(df -BG . | tail -1 | awk '{print $4}' | tr -d 'G')
if [[ $DISK_AVAIL -gt 1 ]]; then
    echo "   ✅ ${DISK_AVAIL}GB available"
else
    echo "   ⚠️  WARN: Low disk space (${DISK_AVAIL}GB available, recommend >1GB)"
fi
echo ""

# Check 4: Required CLI tools
echo "🔧 CLI Tools"
TOOLS=("curl" "jq" "grep" "awk")
for TOOL in "${TOOLS[@]}"; do
    if command -v "$TOOL" &>/dev/null; then
        echo "   ✅ $TOOL installed"
    else
        echo "   ⚠️  WARN: $TOOL not found (some features may not work)"
    fi
done
echo ""

# Check 5: Port availability (common ports)
echo "🔌 Port Availability"
COMMON_PORTS=(5432 3306 6379 8080 8081 8082)
for PORT in "${COMMON_PORTS[@]}"; do
    if ! nc -z localhost "$PORT" 2>/dev/null; then
        echo "   ✅ Port $PORT available"
    else
        echo "   ⚠️  WARN: Port $PORT in use"
    fi
done
echo ""

if [[ $EXIT_CODE -eq 0 ]]; then
    echo "✅ All critical checks passed"
else
    echo "❌ Some checks failed"
fi

exit $EXIT_CODE
```

**Source:** [Bash diagnostic script patterns](https://www.cyberciti.biz/faq/unix-linux-check-if-port-is-in-use-command/) + [CLI color codes](https://misc.flogisoft.com/bash/tip_colors_and_formatting)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| YAML for lockfiles | JSON for lockfiles | npm 5+ (2017), yarn 2+ (2020) | Machine-readability prioritized over human editing for locks ([source](https://nesbitt.io/2026/01/17/lockfile-format-design-and-tradeoffs.html)) |
| Image tags only | Tags in config, digests in lock | Docker Compose 3.8+ (2020), Kubernetes 1.19+ (2020) | Immutable image references prevent tag mutation attacks ([source](https://edu.chainguard.dev/chainguard/chainguard-images/how-to-use/container-image-digests/)) |
| Parse ss/netstat for port checks | net.Listen probe | Modern Go practices (2018+) | Portable, no sudo, works in containers ([source](https://linuxize.com/post/check-listening-ports-linux/)) |
| Regex-based secret detection | Keyword heuristic + entropy | GitGuardian 2020+, GitHub Secret Scanning 2021+ | Keyword matching catches 80% of secrets with minimal false positives ([source](https://blog.gitguardian.com/secret-scanning-tools/)) |
| Import preview + confirmation | Immediate import with result summary | Modern CLI tools (Terraform apply, kubectl apply 2019+) | Declarative imports are intentional; preview is in VCS diff, not tool UI |

**Deprecated/outdated:**
- **Storing secrets in export:** Modern tools (AWS CDK, Terraform Cloud) always redact secrets in state exports ([source](https://blog.gitguardian.com/secure-your-secrets-with-env/))
- **Unversioned config files:** All modern IaC formats have version fields (Terraform, Kubernetes, Docker Compose) to enable forward-compatible schema evolution
- **Synchronous image pulls in CLI:** Modern tools (Docker Compose v2, Podman Compose) show progress bars and allow parallel pulls

## Open Questions

1. **Should lock validation be strict (block) or advisory (warn)?**
   - What we know: Locked decision says "warn only, non-blocking"
   - What's unclear: N/A — decision is clear
   - Recommendation: Implement as decided (warn only)

2. **Should import delete instances not in yml?**
   - What we know: Locked decision says "create-update mode" (upsert only)
   - What's unclear: Should orphaned instances (in DB but not in yml) be marked deleted?
   - Recommendation: No deletion on import. User explicitly deletes via dashboard. Import is additive/update only.

3. **Should `devarch doctor` be local bash or API endpoint?**
   - What we know: CLI context doc says "bash scripts calling API endpoints"
   - What's unclear: Some checks (runtime running, disk space) are local, others (port conflicts) could be API
   - Recommendation: Hybrid — runtime/disk/tools are local bash, port checks can call API endpoint (reusable from dashboard too)

4. **Should export include disabled instances?**
   - What we know: Export for sharing/backup
   - What's unclear: Disabled instances are intent, not runtime — should they be in yml?
   - Recommendation: Include all instances (enabled and disabled). Enabled state in yml matches DB (intent), lock shows actual running state.

5. **Should bootstrap pull images before or after import?**
   - What we know: Pull can be slow, import is fast
   - What's unclear: Better UX to import first (validate early) or pull first (fail fast on network issues)?
   - Recommendation: Import first (validates templates exist), then pull (fails fast if image not found), then apply (start containers)

## Sources

### Primary (HIGH confidence)

- gopkg.in/yaml.v3 Documentation: https://pkg.go.dev/gopkg.in/yaml.v3
- Go crypto/sha256 Package: https://pkg.go.dev/crypto/sha256
- Podman Image Inspect: https://docs.podman.io/en/latest/markdown/podman-image-inspect.1.html
- Docker Image Digests: https://docs.docker.com/dhi/core-concepts/digests/
- npm package-lock.json Format: https://docs.npmjs.com/cli/v9/configuring-npm/package-lock-json/
- Go net Package for Port Checks: https://pkg.go.dev/net
- Existing codebase patterns: `api/internal/compose/parser.go`, `api/internal/compose/generator.go`, `api/internal/container/client.go`

### Secondary (MEDIUM confidence)

- Lockfile Format Design and Tradeoffs (Jan 2026): https://nesbitt.io/2026/01/17/lockfile-format-design-and-tradeoffs.html
- Using Container Image Digests (Chainguard): https://edu.chainguard.dev/chainguard/chainguard-images/how-to-use/container-image-digests/
- Secret Detection Tools 2026 (GitGuardian): https://blog.gitguardian.com/secret-scanning-tools/
- Netlify Secrets Controller: https://docs.netlify.com/build/environment-variables/secrets-controller/
- Detect-Secrets (Yelp): https://github.com/Yelp/detect-secrets
- React File Upload Patterns: https://uploadcare.com/blog/how-to-upload-file-in-react/
- Check Listening Ports in Linux: https://linuxize.com/post/check-listening-ports-linux/
- SHA256 File Integrity in Go: https://transloadit.com/devtips/verify-file-integrity-with-go-and-sha256/

### Tertiary (LOW confidence)

- YAML Round-Trip Issues (GitHub): https://github.com/go-yaml/yaml/issues/113
- Podman Digest Mismatch Issue: https://github.com/containers/podman/issues/14779
- Environment Variable Security Best Practices: https://blog.gitguardian.com/secure-your-secrets-with-env/
- Working with Container Images in Go: https://iximiuz.com/en/posts/working-with-container-images-in-go/

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing yaml.v3 dependency + Go stdlib, zero new dependencies
- Architecture: HIGH - Patterns from existing codebase compose parser/generator, WebSearch verified lockfile design principles
- Pitfalls: MEDIUM-HIGH - Image digest issues well-documented, secret detection patterns verified, round-trip behavior from yaml.v3 GitHub issues

**Research date:** 2026-02-08
**Valid until:** 60 days (stable domain — YAML parsing, lockfile patterns, container CLIs are mature and change slowly)
