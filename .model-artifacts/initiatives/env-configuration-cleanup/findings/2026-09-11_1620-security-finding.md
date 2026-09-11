# Security Finding — Environment Configuration Cleanup

- Specialist: security
- Status: complete
- Assessed spec: `.model-artifacts/initiatives/env-configuration-cleanup/specs/2026-09-11_1614-initiative-spec-r1.md` (`sha256:16eb04ced8ab44ab1865b2d14f7987fda81644d45d103d0ee047fd3bb006b1e3`)
- Assessed plan: `.model-artifacts/initiatives/env-configuration-cleanup/plans/2026-09-11_1614-plan-index-r1.md` (`sha256:fb6a13277a1f59005a11d28b41f38c84107af91a0dbcea9cf4a8430506b27b48`)

## Decision

Required. The plan addresses the consequential boundary, subject to the controls below.

## Findings to incorporate

1. **High — Treat dotenv as untrusted data.** The loader must not use `source`, `eval`, command substitution, indirect expansion of file content, or executable generated shell. Match requested keys against a caller-supplied fixed allowlist before assignment.
2. **High — Prevent ambient secret propagation.** Do not use `set -a` and do not export file values solely because they came from dotenv. Preserve an existing exported attribute only when inherited Bash semantics naturally do so; explicitly pass the MariaDB password only to the Compose command that requires interpolation.
3. **Medium — Fail without reflecting secrets.** NUL/control-byte, malformed recognized quoting, invalid allowlist key, or explicit env-file errors may name the key/path but must not print the raw line/value.
4. **Medium — Keep tracked examples non-credentialed.** Password/token fields must be blank or unmistakably invalid and documentation must require replacement before mutation. No secret moves into tracked Compose YAML.
5. **Low — Unrecognized malformed lines remain inert.** They should be ignored rather than parsed/evaluated; this preserves Laravel's selective boundary and avoids rejecting legacy debris unnecessarily.

## Residual risk

The shared loader remains custom parsing code. The bounded grammar and table-driven tests are acceptable for this local Bash tooling; expanding toward full dotenv dialect compatibility requires a new plan.
