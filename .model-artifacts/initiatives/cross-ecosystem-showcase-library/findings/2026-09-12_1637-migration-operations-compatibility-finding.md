# Migration, operations, and compatibility finding

- Reviewed spec: r1 (`sha256:8d212d0bd6e0f2770f79080659bfd8b3fc7a81a7bc57b38172487185f3b4ec92`)
- Reviewed plan: r1 (`sha256:a2904fe4dd7cef344fa39f67fc701ba48f2d4d08dac0bc1fd77a3a38bb9c470d`)
- Decision: request revisions before approval

## Findings

1. **High — compatibility:** r1 does not define a version/support policy for current upstream starter kits and third-party capabilities. Add profile metadata for support tier and compatibility probe, record resolved Laravel/starter/package versions in generated local evidence, and fail a profile as unsupported before later capability mutation where resolution proves incompatible.
2. **High — operations:** real smoke expectations for five profiles are costly and mutable. Separate default fixture/characterization verification from opt-in network/runtime golden paths; require unique run IDs, disk/network prerequisite checks, bounded timeouts, cleanup instructions, and preserved failure logs.
3. **High — migration:** legacy `bare`, `standard`, and `loaded` must remain canonical public names rather than silently becoming aliases to new names. Characterize exact resolved plans, retain files/CLI output, and implement new profiles additively.
4. **Medium — operations:** recipe recovery state needs atomic-write and stale-definition behavior plus an explicit operator `status`/`resume`; r1 mentions resume but the top-level action contract omits it.
5. **Medium — compatibility:** adapter protocol must distinguish unsupported lifecycle actions without pretending all ecosystems have identical create/start semantics. Require capability negotiation in list/status and preserve native entry points.
6. **Medium — migration:** existing ignored `apps/showcase-laravel` and `apps/showcase-wp` do not match new deterministic matrix names. Treat them as unmanaged local apps; do not rename, adopt, overwrite, or delete them automatically.
7. **Medium — operations:** service-rich capabilities must declare required catalog service IDs and health probes. Preflight missing services before app/database mutation.
8. **Low — scope:** catalog external-account ideas (`billing`, WorkOS/social, multi-tenancy, Octane) only as documented future candidates, not executable definitions in this initiative.

## Required r2 incorporation

Update spec acceptance, adapter protocol, Laravel model/catalog, recipe engine, verification contract, and migration section with all findings. No blocker remains after those changes and hash/link reconciliation.
