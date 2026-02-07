# Phase 5: Compose Generation - Research

**Researched:** 2026-02-07
**Domain:** Docker Compose YAML generation in Go
**Confidence:** HIGH

## Summary

Stack-scoped compose generation extends existing single-service generation to produce unified YAML with multiple service instances. The established Go ecosystem provides gopkg.in/yaml.v3 for YAML marshaling and compose-go (v2.10.1) for spec validation. The codebase already has proven patterns: struct-based generation with yaml tags, path resolution for volumes/configs, and ComposeOverrides JSON field for unknown keys.

Key architectural decision: reuse existing serviceConfig struct and generator internals—this is NOT a rewrite but an extension. The materialization directory layout (compose/stacks/{stack}/{instance}/) provides clean isolation from legacy single-service paths. Atomic writes (temp-then-rename) prevent half-written state during failures.

Dashboard integration follows established patterns: CodeMirror 6 with @codemirror/lang-yaml already present in package.json, existing service compose preview uses plain <pre> tags (upgrade to CodeMirror for stack version).

**Primary recommendation:** Extend existing compose.Generator with GenerateStack method that iterates effective configs, reusing serviceConfig marshaling. No new libraries needed—gopkg.in/yaml.v3 handles multi-service YAML correctly.

## Standard Stack

The established libraries/tools for docker-compose generation in Go:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| gopkg.in/yaml.v3 | v3.0.1 | YAML marshaling/unmarshaling | De facto standard for YAML in Go, supports struct tags, custom marshaling, omitempty |
| github.com/compose-spec/compose-go | v2.10.1 (Jan 2026) | Compose spec validation | Official reference library from compose-spec org, validates against Compose spec |
| @codemirror/lang-yaml | ^6.1.2 | YAML syntax highlighting | Official CodeMirror language package for YAML |
| @codemirror/theme-one-dark | ^6.1.3 | Editor theme | Consistent with existing code-editor.tsx usage |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/google/renameio | Latest | Atomic file writes | When Windows compatibility matters (project uses Linux) |
| github.com/natefinch/atomic | Latest | Cross-platform atomic writes | Alternative to renameio |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gopkg.in/yaml.v3 | sigs.k8s.io/yaml | K8s YAML is wrapper around go-yaml with additional JSON compatibility—unnecessary complexity |
| compose-go validation | Custom validation | Hand-rolling compose spec validation misses edge cases (anchors, fragments, merge keys) |
| CodeMirror 6 | Monaco Editor | Monaco is larger bundle, less modular—CodeMirror already in package.json |

**Installation:**
```bash
# Go dependencies already present
go get gopkg.in/yaml.v3  # already in go.mod
go get github.com/compose-spec/compose-go/v2  # if adding validation

# Dashboard dependencies already present
npm install @codemirror/lang-yaml  # already in package.json v6.1.2
```

## Architecture Patterns

### Recommended Project Structure
```
api/internal/compose/
├── generator.go         # Generator.Generate() existing + new GenerateStack()
├── parser.go            # Existing import logic
├── validator.go         # Existing validation
└── importer.go          # Existing import logic

compose/stacks/{stack}/
├── {instance-1}/        # Config files for instance 1
│   ├── nginx.conf
│   └── .env
├── {instance-2}/        # Config files for instance 2
└── docker-compose.yml   # Generated stack compose (optional materialization)
```

### Pattern 1: Struct-Based YAML Generation (Existing Pattern)
**What:** Define Go structs with yaml tags, marshal to YAML via gopkg.in/yaml.v3
**When to use:** All compose generation (already used by existing Generator.Generate)
**Example:**
```go
// Source: api/internal/compose/generator.go lines 38-61
type generatedCompose struct {
	Networks map[string]networkConfig `yaml:"networks"`
	Volumes  map[string]interface{}   `yaml:"volumes,omitempty"`
	Services map[string]serviceConfig `yaml:"services"`
}

type serviceConfig struct {
	Image         string                 `yaml:"image,omitempty"`
	ContainerName string                 `yaml:"container_name"`
	Restart       string                 `yaml:"restart,omitempty"`
	Command       interface{}            `yaml:"command,omitempty"`
	Ports         []string               `yaml:"ports,omitempty"`
	Volumes       []string               `yaml:"volumes,omitempty"`
	Environment   map[string]string      `yaml:"environment,omitempty"`
	DependsOn     []string               `yaml:"depends_on,omitempty"`
	Networks      []string               `yaml:"networks"`
	Healthcheck   *healthcheckConfig     `yaml:"healthcheck,omitempty"`
}
```

### Pattern 2: Effective Config Iteration for Stack Generation
**What:** Loop through stack instances, build service config from effective config (base + overrides), marshal all at once
**When to use:** Stack compose generation
**Pseudocode:**
```go
func (g *Generator) GenerateStack(stackName string) ([]byte, []string, error) {
	// 1. Query instances for stack with effective configs
	instances := fetchStackInstances(stackName)

	compose := generatedCompose{
		Networks: map[string]networkConfig{networkName: {External: true}},
		Services: make(map[string]serviceConfig),
	}
	warnings := []string{}

	// 2. Iterate instances
	for _, inst := range instances {
		if !inst.Enabled {
			continue  // Skip disabled
		}

		// 3. Build serviceConfig from effective config
		svc := buildServiceConfig(inst.EffectiveConfig)

		// 4. Set instance-specific fields
		svc.ContainerName = fmt.Sprintf("devarch-%s-%s", stackName, inst.ID)

		// 5. Transform depends_on (strip disabled, add conditions)
		svc.DependsOn = resolveDependencies(inst, instances)

		compose.Services[inst.ID] = svc
	}

	// 6. Marshal entire structure
	yamlBytes, err := yaml.Marshal(compose)
	return yamlBytes, warnings, err
}
```

### Pattern 3: Config File Materialization with Atomic Writes
**What:** Write configs to temp directory, rename to final path for atomicity
**When to use:** Config file materialization to compose/stacks/{stack}/{instance}/
**Example:**
```go
// Source: Linux os.Rename is atomic (project targets Linux)
tempDir := filepath.Join(baseDir, "compose", "stacks", stack, ".tmp-"+uuid)
os.MkdirAll(tempDir, 0755)

// Write all configs to tempDir
for _, inst := range instances {
	instPath := filepath.Join(tempDir, inst.ID)
	writeConfigFiles(inst, instPath)
}

// Atomic rename (cleanup old, move new)
finalDir := filepath.Join(baseDir, "compose", "stacks", stack)
os.RemoveAll(finalDir)  // Cleanup stale
os.Rename(tempDir, finalDir)  // Atomic on Linux
```

### Pattern 4: Conditional depends_on (service_healthy vs simple list)
**What:** Use service_healthy when target has healthcheck, simple list otherwise
**When to use:** All depends_on generation
**Example:**
```go
// Simple list (target has no healthcheck)
DependsOn: []string{"db-01", "redis-cache"}

// Would marshal to:
// depends_on:
//   - db-01
//   - redis-cache

// Condition-based (requires long-form YAML, not supported by []string)
// NOTE: gopkg.in/yaml.v3 with []string only produces simple list
// For conditions, need map[string]interface{} or dedicated struct
```

**IMPORTANT:** Current serviceConfig.DependsOn is `[]string` which only produces simple list format. To generate condition-based depends_on (service_healthy), must change to `interface{}` and handle both formats:
```go
// Recommended change
DependsOn interface{} `yaml:"depends_on,omitempty"`

// Then populate:
if targetHasHealthcheck {
	svc.DependsOn = map[string]interface{}{
		"db-01": map[string]string{"condition": "service_healthy"},
	}
} else {
	svc.DependsOn = []string{"db-01"}
}
```

### Pattern 5: Dashboard Compose Preview with CodeMirror
**What:** Read-only CodeMirror editor with YAML syntax highlighting
**When to use:** Stack compose preview tab
**Example:**
```tsx
// Source: dashboard/src/components/services/code-editor.tsx
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'

<CodeEditor
  value={composeYaml}
  language="yaml"
  readOnly={true}
/>

// Hook pattern (following useServiceCompose)
export function useStackCompose(name: string) {
  return useQuery({
    queryKey: ['stacks', name, 'compose'],
    queryFn: async () => {
      const response = await api.get(`/stacks/${name}/compose`)
      return response.data  // { yaml: "...", warnings: [...] }
    },
    enabled: !!name,
  })
}
```

### Anti-Patterns to Avoid
- **Don't use map[string]interface{} everywhere:** Typed structs with yaml tags are clearer, safer, easier to test
- **Don't parse generated YAML back to validate:** If generation logic is correct, YAML is correct—parsing adds overhead and circular dependency risk
- **Don't build YAML strings manually:** String concatenation breaks on special chars, indentation—use structs + yaml.Marshal
- **Don't set container_name from service template name:** Stack instances need unique names (devarch-{stack}-{instance})

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML marshaling | String builder concatenation | gopkg.in/yaml.v3 struct tags | Handles escaping, indentation, special chars, flow style, anchors |
| Compose spec validation | Custom field checks | compose-go loader/validator | Compose spec has 100+ edge cases (fragments, extends, profiles, merge) |
| Atomic file writes | Write directly to final path | Write to temp + os.Rename | Half-written state on crash/SIGKILL, race conditions |
| Map key ordering | Custom sorting | Accept Go map randomness | YAML spec doesn't mandate key order, compose tools handle any order |
| Port conflict detection | Parse all port strings | Collect host ports, check duplicates | Port strings have 6 formats (8080, 8080:80, 0.0.0.0:8080:80/tcp, ranges) |
| depends_on circular refs | Manual graph traversal | Skip—compose up handles at runtime | Docker Compose CLI detects cycles, generates better error messages |

**Key insight:** Compose YAML generation is data transformation, not parsing. The hard problems (validation, cycles, network setup) happen at `docker compose up` time, not generation time. Generator produces syntactically valid YAML; compose CLI enforces runtime semantics.

## Common Pitfalls

### Pitfall 1: YAML Indentation Sensitivity
**What goes wrong:** YAML uses spaces (not tabs) and strict indentation—single extra space breaks parsing
**Why it happens:** String concatenation or manual formatting
**How to avoid:** Always use struct marshaling with gopkg.in/yaml.v3—never build YAML strings
**Warning signs:** yaml.Unmarshal errors like "mapping values are not allowed in this context"

### Pitfall 2: depends_on Startup Order vs Readiness
**What goes wrong:** Service starts before dependency is ready (DB accepting connections)
**Why it happens:** depends_on controls start order, not readiness—container running ≠ service ready
**How to avoid:** Use service_healthy condition when target has healthcheck, document that apps should retry connections
**Warning signs:** Services crash on startup with "connection refused" then succeed on restart

### Pitfall 3: Map Key Ordering Assumptions
**What goes wrong:** Generated YAML has different service order on each generation
**Why it happens:** Go maps have random iteration order
**How to avoid:** Accept it—YAML spec and compose spec don't mandate key order, compose tools work correctly regardless
**Warning signs:** Diff noise in version control (services reordered but content identical)
**Note:** If determinism needed, sort keys before iteration—but adds complexity for minimal benefit

### Pitfall 4: Disabled Instance References in depends_on
**What goes wrong:** Service depends on disabled instance, compose up fails with "service X not found"
**Why it happens:** Instance disabled after dependency created, generator doesn't filter
**How to avoid:** Strip depends_on references to disabled instances, add warning
**Warning signs:** Compose validation error "service 'db-02' must be present in depends_on"

### Pitfall 5: Port Conflicts Between Stack Instances
**What goes wrong:** Two instances bind same host port, second fails with "port already allocated"
**Why it happens:** Instance configs copied from same template without port adjustments
**How to avoid:** Detect duplicate host ports during generation, add warning (don't block—user may intend to run separately)
**Warning signs:** docker compose up succeeds for first service, fails for second with bind error

### Pitfall 6: Config File Race Conditions During Materialization
**What goes wrong:** compose up reads half-written config file, container fails to start
**Why it happens:** Writing directly to final path while compose reads
**How to avoid:** Write to temp directory, atomic rename to final path (os.Rename is atomic on Linux)
**Warning signs:** Intermittent config parse errors, works on retry

### Pitfall 7: Stale Config Files from Deleted Instances
**What goes wrong:** Old instance configs remain on disk after instance deleted
**Why it happens:** Materialization adds new files but doesn't clean old ones
**How to avoid:** Delete compose/stacks/{stack}/ before materializing (files always generated from DB)
**Warning signs:** Extra directories in compose/stacks/{stack}/ that don't match current instances

### Pitfall 8: omitempty with Zero-Valued Structs
**What goes wrong:** Empty struct fields appear in YAML despite omitempty tag
**Why it happens:** Go considers struct{} non-zero unless IsZero() method defined
**How to avoid:** Use pointer types for optional struct fields (e.g., *healthcheckConfig)
**Warning signs:** Empty healthcheck blocks in YAML when no healthcheck configured

### Pitfall 9: Command Field Type (string vs []string)
**What goes wrong:** Commands with spaces break when unmarshaled back
**Why it happens:** YAML supports both "command: npm start" and "command: [npm, start]"—interface{} needed
**How to avoid:** Use interface{} for command field, handle both string and array when reading
**Warning signs:** Commands with args parsed as single string with spaces

## Code Examples

Verified patterns from existing codebase:

### Generating Multi-Service Compose
```go
// Reuse existing pattern, extend for multiple services
func (g *Generator) GenerateStack(stackName string) ([]byte, []string, error) {
	// Query instances with effective configs
	rows, err := g.db.Query(`
		SELECT i.id, i.enabled, i.effective_config, s.name as service_name
		FROM stack_instances i
		JOIN stacks st ON i.stack_id = st.id
		JOIN services s ON i.service_id = s.id
		WHERE st.name = $1
		ORDER BY i.id
	`, stackName)

	compose := generatedCompose{
		Networks: map[string]networkConfig{g.networkName: {External: true}},
		Services: make(map[string]serviceConfig),
	}
	warnings := []string{}

	for rows.Next() {
		var inst instanceData
		rows.Scan(&inst.ID, &inst.Enabled, &inst.EffectiveConfig, &inst.ServiceName)

		if !inst.Enabled {
			warnings = append(warnings, fmt.Sprintf("Skipped disabled instance: %s", inst.ID))
			continue
		}

		svc := g.buildServiceConfig(inst)
		svc.ContainerName = fmt.Sprintf("devarch-%s-%s", stackName, inst.ID)
		compose.Services[inst.ID] = svc
	}

	yamlBytes, err := yaml.Marshal(compose)
	return yamlBytes, warnings, err
}
```

### Materializing Config Files Atomically
```go
// Source: Atomic write pattern (os.Rename is atomic on Linux)
func (g *Generator) MaterializeStackConfigs(stackName, baseDir string) error {
	tempDir := filepath.Join(baseDir, "compose", "stacks", ".tmp-"+stackName)
	finalDir := filepath.Join(baseDir, "compose", "stacks", stackName)

	// Clean slate
	os.RemoveAll(tempDir)
	os.MkdirAll(tempDir, 0755)

	// Write all instance configs to temp
	instances := fetchStackInstances(stackName)
	for _, inst := range instances {
		instDir := filepath.Join(tempDir, inst.ID)
		os.MkdirAll(instDir, 0755)

		for _, configFile := range inst.ConfigFiles {
			path := filepath.Join(instDir, configFile.FilePath)
			os.MkdirAll(filepath.Dir(path), 0755)
			mode := parseFileMode(configFile.FileMode)
			os.WriteFile(path, []byte(configFile.Content), mode)
		}
	}

	// Atomic swap (cleanup old, move new)
	os.RemoveAll(finalDir)
	return os.Rename(tempDir, finalDir)
}
```

### Filtering Disabled Dependencies
```go
// Build map of enabled instances for lookup
enabledMap := make(map[string]bool)
for _, inst := range instances {
	if inst.Enabled {
		enabledMap[inst.ID] = true
	}
}

// Filter depends_on to enabled instances only
filteredDeps := []string{}
for _, depID := range rawDependencies {
	if enabledMap[depID] {
		filteredDeps = append(filteredDeps, depID)
	} else {
		warnings = append(warnings, fmt.Sprintf(
			"Stripped dependency on disabled instance: %s -> %s",
			inst.ID, depID,
		))
	}
}
svc.DependsOn = filteredDeps
```

### Dashboard Compose Preview
```tsx
// Source: Existing useServiceCompose pattern + CodeEditor component
import { CodeEditor } from '@/components/services/code-editor'
import { useStackCompose } from '@/features/stacks/queries'

function StackCompose({ stackName }: { stackName: string }) {
  const { data, isLoading } = useStackCompose(stackName)

  if (isLoading) return <Loader />

  const { yaml, warnings } = data

  return (
    <div>
      <CodeEditor value={yaml} language="yaml" readOnly />

      {warnings?.length > 0 && (
        <div className="mt-4 space-y-2">
          {warnings.map((w, i) => (
            <Alert key={i} variant="warning">{w}</Alert>
          ))}
        </div>
      )}

      <Button onClick={() => downloadFile('docker-compose.yml', yaml)}>
        Download
      </Button>
    </div>
  )
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Per-service compose files | Stack compose with N services | Phase 5 (2026) | Eliminates manual YAML merging |
| Compose files as source of truth | DB as source of truth | Project inception | YAML always generated from DB state |
| Simple depends_on lists | Condition-based (service_healthy) | Compose v2 (2020+) | Services wait for readiness, not just start |
| container_name = service name | container_name explicitly set | Phase 5 | Prevents naming collisions in multi-stack setups |
| Manual network creation | external: true (pre-created) | Phase 4 | Generator doesn't manage network lifecycle |

**Deprecated/outdated:**
- **libcompose:** Abandoned in favor of docker/compose v2 and compose-go
- **docker-compose v1 (Python):** Replaced by v2 (Go) in 2020—v2 is now standard
- **YAML 1.1 octal literals (0777):** YAML 1.2 prefers 0o777 but 0777 still works for backward compat
- **depends_on without conditions:** Still valid but service_healthy is best practice when healthchecks exist

## Open Questions

Things that couldn't be fully resolved:

1. **Map key ordering in gopkg.in/yaml.v3**
   - What we know: Go maps have random iteration order, YAML spec doesn't mandate key order
   - What's unclear: Whether gopkg.in/yaml.v3 preserves insertion order or randomizes
   - Recommendation: Accept random order (compose tools work correctly), or sort keys before iteration if determinism critical

2. **compose-go validation integration value**
   - What we know: compose-go can validate generated YAML against Compose spec
   - What's unclear: Whether validation overhead is worth it (generation logic should produce valid YAML)
   - Recommendation: Skip validation in generation path—if YAML is malformed, docker compose up will error clearly

3. **Windows compatibility for atomic writes**
   - What we know: os.Rename is atomic on Linux but not Windows
   - What's unclear: Whether project targets Windows (Dockerfile suggests Linux containers)
   - Recommendation: Stick with os.Rename (project uses Linux), document Windows limitation if needed

4. **Handling circular dependencies**
   - What we know: Docker Compose CLI detects circular depends_on at runtime
   - What's unclear: Whether generator should detect/warn or let compose CLI handle
   - Recommendation: Let compose CLI handle—better error messages, avoids graph traversal complexity

## Sources

### Primary (HIGH confidence)
- [compose-go GitHub](https://github.com/compose-spec/compose-go) - Official compose spec library, v2.10.1 (Jan 2026)
- [gopkg.in/yaml.v3 docs](https://pkg.go.dev/gopkg.in/yaml.v3) - YAML v3 API, struct tags, custom marshaling
- [Docker Compose services reference](https://docs.docker.com/reference/compose-file/services/) - Service definition fields
- [Docker Compose networking](https://docs.docker.com/compose/how-tos/networking/) - external network pattern
- [CodeMirror lang-yaml](https://github.com/codemirror/lang-yaml) - Official YAML language package

### Secondary (MEDIUM confidence)
- [Docker Compose startup order](https://docs.docker.com/compose/how-tos/startup-order/) - depends_on with service_healthy
- [Docker Compose health checks](https://last9.io/blog/docker-compose-health-checks/) - Healthcheck configuration
- [Atomic file writes in Go](https://michael.stapelberg.ch/posts/2017-01-28-golang_atomically_writing/) - Temp-then-rename pattern
- [CodeMirror 6 React setup](https://github.com/uiwjs/react-codemirror) - React wrapper for CodeMirror

### Tertiary (LOW confidence)
- [Common Docker Compose mistakes](https://moldstud.com/articles/p-avoid-these-common-docker-compose-pitfalls-tips-and-best-practices) - YAML pitfalls, port conflicts
- [Docker Compose port conflicts](https://www.netdata.cloud/academy/docker-compose-networking-mysteries/) - Multi-service port detection
- [Go os.RemoveAll race conditions](https://github.com/golang/go/issues/51442) - Windows cleanup failures

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - gopkg.in/yaml.v3 is de facto standard, compose-go is official reference, CodeMirror already in use
- Architecture: HIGH - Existing generator.go provides proven patterns, struct-based approach is idiomatic Go
- Pitfalls: MEDIUM - YAML pitfalls are well-documented, depends_on behavior is official docs, some edge cases are experiential

**Research date:** 2026-02-07
**Valid until:** 2026-04-07 (60 days—Go YAML libraries stable, compose spec changes infrequently)
