# Milestones

## v1.0 Stacks & Instances (Shipped: 2026-02-09)

**Stats:** 9 phases, 30 plans | 183 commits | 294 files | +47,453 lines | 7 days (Feb 3–9, 2026)

**Key accomplishments:**
- Stack CRUD with soft-delete, clone-as-rename, and full dashboard UI (grid + table views, action dialogs)
- Service instances with full copy-on-write overrides (ports, volumes, env, labels, deps, domains, healthchecks, config files)
- Per-stack network isolation with deterministic container naming (devarch-{stack}-{instance})
- Stack compose generation with config materialization and identity label injection
- Terraform-style plan/apply workflow with advisory locking and staleness detection
- devarch.yml export/import with lockfile (devarch.lock) for deterministic reproduction
- Contract-based auto-wiring with explicit wiring for ambiguous cases, env var injection via DNS
- AES-256-GCM secret encryption at rest with redaction in all outputs
- Resource limits (CPU/memory) mapped to compose deploy.resources

---


## v1.1 Schema Reconciliation (Shipped: 2026-02-10)

**Stats:** 6 phases, 14 plans | 73 commits | 402 files | +21,548 lines | 2 days (Feb 9–10, 2026)

**Key accomplishments:**
- Replaced 23 patch migrations with 9 domain-separated fresh baseline (39 tables, zero seed data)
- Parser/importer preserves env_files, container_name, networks, config mounts with cross-service FK provenance
- Compose generator emits all new fields from DB — 172/173 services match legacy 1:1 (1 whitelisted)
- Streaming multipart import handles 256MB payloads without memory exhaustion, prepared statement batching
- Dashboard UI for env_files, networks, config_mounts at both service and instance level with override support
- Living verification tools (verify-parity, verify-boundary) prove parity and boundary behavior

---

