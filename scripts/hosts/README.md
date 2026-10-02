# Local hosts registration

## Synchronize every DevArch domain

`devarch hosts sync` (see [`cli/README.md`](../../cli/README.md)) discovers every literal `container_name` in the catalog, `.test` hosts listed in `x-devarch.urls`, routable workspaces in `apps/*`, and `devarch.test`. It writes them as sorted `.test` domains inside a clearly delimited managed block:

```bash
devarch --dry-run hosts sync
devarch hosts sync
devarch hosts list          # show the block and whether it is current
```

Only content between `# BEGIN DEVARCH HOSTS` and `# END DEVARCH HOSTS` is replaced. Unrelated hosts-file content is preserved, and rerunning an already-current synchronization performs no write. App directories are included when they contain `index.php`, `public/index.php`, `public/index.html`, or `package.json`. Catalog services are included whether or not their containers are currently running. `devarch up` synchronizes automatically when a started service's hostname is unmapped.

On Linux and macOS `devarch` updates `/etc/hosts` (or `HOSTS_FILE`), requesting `sudo` once when needed. Under WSL and Git Bash/MSYS it delegates to its embedded `sync-hosts.ps1`, which requests Windows UAC elevation and updates `%SystemRoot%\System32\drivers\etc\hosts` while preserving the file's encoding and line endings. Run it as your normal WSL user rather than with `sudo`; it invokes Windows PowerShell through WSL's `/init` interop host so Wine or another `.exe` binfmt handler cannot intercept it.

## Register one domain

`devarch hosts add` idempotently maps one validated local hostname to `127.0.0.1`; `devarch hosts remove` unmaps one on Linux and macOS. The application bootstraps still call `register-host.sh`, which has the same behavior, until they move onto the `devarch` binary.

```bash
devarch hosts add demo.test
scripts/hosts/register-host.sh demo.test --dry-run
```

The helper uses the same Unix/Windows elevation behavior. Existing mappings for the requested hostname are removed without removing other aliases on the same line, then one current mapping is appended. Re-registering an already-canonical mapping performs no write.

Use `--no-hosts` on an application bootstrap when local DNS already resolves `*.test`, when provisioning non-interactively, or when hosts-file access is intentionally managed elsewhere. A denied elevation prompt leaves the created application intact and prints a manual mapping warning.

Tests use temporary hosts-file overrides and never edit the real system file:

```bash
(cd cli && go test ./internal/hosts/)
scripts/hosts/register-host.test.sh
```
