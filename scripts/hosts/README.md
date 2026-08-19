# Local hosts registration

`register-host.sh` idempotently maps a validated local hostname to `127.0.0.1`. The WordPress and Laravel bootstraps call it after successful provisioning.

```bash
scripts/hosts/register-host.sh demo.test
scripts/hosts/register-host.sh demo.test --dry-run
```

On Linux and macOS it updates `/etc/hosts`, requesting `sudo` when needed. Under WSL and Git Bash/MSYS it delegates to `register-host.ps1`, which requests Windows UAC elevation and updates `%SystemRoot%\System32\drivers\etc\hosts`. Unrelated lines, spacing, encoding, and line endings are preserved. Existing mappings for the requested hostname are removed without removing other aliases on the same line, then one current mapping is appended. Re-registering an already-canonical mapping performs no write.

Use `--no-hosts` on either application bootstrap when local DNS already resolves `*.test`, when provisioning non-interactively, or when hosts-file access is intentionally managed elsewhere. A denied elevation prompt leaves the created application intact and prints a manual mapping warning.

Tests use a temporary hosts-file override and never edit the real system file:

```bash
scripts/hosts/register-host.test.sh
```
