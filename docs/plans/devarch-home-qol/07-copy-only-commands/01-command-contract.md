# Phase 1 — Command catalog and shell-safety contract

Created: 2026-08-29
Purpose: Make command generation explicit, portable, and resistant to shell injection.

## Goal

Represent copyable commands as structured descriptors and render them with one tested POSIX-shell quoting implementation.

## Scope

Structured command descriptors, POSIX rendering, classification, initial allowlisted catalog, adversarial fixtures, and action-registry payload contract. No visible command UI.

## Descriptor

```text
id, entityId, label, description,
classification: navigation | inspect | lifecycle,
segments: argv groups joined by explicit operators,
platform: posix,
confirmationNote
```

The backend renderer receives validated argv arrays and explicit operators (`&&` only); it never accepts a pre-interpolated fragment. Use Python `shlex.join` where inventory paths/names are introduced. The serialized descriptor contains the rendered copy string and bounded metadata, never argv/operator arrays. Never use `shell=True`.

## Initial catalog

- App: `cd -- <absolute-workspace-path>` (`navigation`).
- Service: `cd -- <catalog-path> && podman compose ps` (`inspect`).
- Service: `cd -- <catalog-path> && podman compose logs --tail 200` (`inspect`).
- Service: existing `podman compose up -d` (`lifecycle`) retained but visually distinguished; do not add more lifecycle commands.
- Container: `podman logs --tail 200 -- <name-or-id>` (`inspect`).

Prefer immutable container ID when available. Commands must remain bounded where the native tool supports it (`--tail 200`). Treat commands as a snapshot: if inventory changes after refresh, the native command may report that the target no longer exists.

## Outputs

- Structured backend command builders and POSIX renderer with an explicit descriptor schema version.
- Additive serialized `commands` descriptors containing the rendered copy string and bounded metadata; internal argv/operator segments do not cross the API boundary.
- Action-registry copy entries generated from those descriptors without rebuilding or interpolating commands in the client.

## Acceptance criteria

- Spaces, quotes, dollar signs, semicolons, newlines, leading dashes, Unicode, and empty values are handled or rejected safely.
- Operators cannot originate from inventory data.
- No command runs during generation or copying.
- Classification is mandatory and tested.
- Exact argv/operator order is verified against the installed Podman and configured Compose provider help/behavior; unsupported provider commands are omitted with an explanation rather than copied optimistically.

## Verification

Pure descriptor/rendering tests plus adversarial argv fixtures covering whitespace, shell metacharacters, Unicode, leading dashes, control characters, and empty values.