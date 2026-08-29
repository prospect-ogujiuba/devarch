# Local hosts registration

## Synchronize every DevArch domain

`sync-hosts.sh` discovers literal `container_name` values in `services-library/**/compose.yml`, routable workspaces in `apps/*`, and `devarch.test`. It writes them as sorted `.test` domains inside a clearly delimited managed block:

```bash
scripts/hosts/sync-hosts.sh --dry-run
scripts/hosts/sync-hosts.sh
```

Only content between `# BEGIN DEVARCH HOSTS` and `# END DEVARCH HOSTS` is replaced. Unrelated hosts-file content is preserved, and rerunning an already-current synchronization performs no write. App directories are included when they contain `index.php`, `public/index.php`, `public/index.html`, or `package.json`. Catalog services are included whether or not their containers are currently running.

On Linux and macOS the script updates `/etc/hosts`, requesting `sudo` once when needed. Under WSL and Git Bash/MSYS it delegates to `sync-hosts.ps1`, which requests Windows UAC elevation and updates `%SystemRoot%\System32\drivers\etc\hosts` while preserving the file's encoding and line endings.

## Register one domain

`register-host.sh` idempotently maps one validated local hostname to `127.0.0.1`. Application bootstraps call it after successful provisioning.

```bash
scripts/hosts/register-host.sh demo.test
scripts/hosts/register-host.sh demo.test --dry-run
```

The helper uses the same Unix/Windows elevation behavior. Existing mappings for the requested hostname are removed without removing other aliases on the same line, then one current mapping is appended. Re-registering an already-canonical mapping performs no write.

Use `--no-hosts` on an application bootstrap when local DNS already resolves `*.test`, when provisioning non-interactively, or when hosts-file access is intentionally managed elsewhere. A denied elevation prompt leaves the created application intact and prints a manual mapping warning.

Tests use temporary hosts-file overrides and never edit the real system file:

```bash
scripts/hosts/sync-hosts.test.sh
scripts/hosts/register-host.test.sh
```
