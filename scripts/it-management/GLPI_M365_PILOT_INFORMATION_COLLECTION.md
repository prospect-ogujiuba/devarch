# GLPI Microsoft 365 pilot information collection guide

**Status:** Pre-change information collection

**Audience:** GLPI, Active Directory, Microsoft Entra and Exchange administrators

**Applies to:** Local GLPI 11 pilot, hybrid AD/Microsoft 365, identity and email intake

**Excludes:** GLPI Agent, Intune application deployment and endpoint inventory

## 1. Purpose

Use this guide to gather the information required before starting the separate `glpi-m365-pilot-execution` initiative. Collection is read-only: do not install plugins, create credentials, enable synchronization, change authentication, or configure a mailbox receiver during this phase.

Record results in the gitignored local file:

```text
scripts/it-management/.local/GLPI_M365_PILOT_INPUTS.md
```

A tracked blank template is available at `GLPI_M365_PILOT_INPUTS.example.md`. The local record may contain internal hostnames and user/group identifiers, but must not contain passwords, bind credentials, OAuth secrets, access/refresh tokens, private keys, certificate files, personal email content, or screenshots. Store sensitive evidence in the approved operational evidence store and record only a non-sensitive reference.

## 2. Proposed pilot architecture

The current proposed path is:

```text
On-premises AD --LDAPS--> GLPI
Microsoft 365 support mailbox --OAuth IMAP--> GLPI receiver
```

AD remains authoritative. Use exactly one provisioning path for pilot users: do not enable SCIM while LDAPS is selected. Use `userPrincipalName` as the GLPI login and validate `objectGUID` as the stable synchronization identifier against the installed GLPI version.

## 3. Confirm hybrid identity and authority

### Entra administrator

1. Open **Microsoft Entra admin center → Identity → Hybrid management → Microsoft Entra Connect**.
2. Confirm synchronization is configured and healthy.
3. Open **Identity → Users → All users**, select one proposed pilot user, and inspect its properties.
4. Record whether **On-premises sync enabled** is `Yes`, the displayed source, and the last synchronization status/time.
5. Confirm that the user's `userPrincipalName` is the Microsoft 365 sign-in name.

Record only the result and a sanitized pilot label. Do not export tenant IDs or screenshots into Git.

### Active Directory administrator

Run these from an authorized management workstation with the ActiveDirectory PowerShell module. Replace placeholders locally:

```powershell
Get-ADUser -Identity '<pilot-user>' -Properties `
  UserPrincipalName,SamAccountName,ObjectGUID,Enabled |
  Select-Object UserPrincipalName,SamAccountName,ObjectGUID,Enabled

Get-ADGroup -Identity '<pilot-group>' |
  Select-Object Name,DistinguishedName,GroupScope,GroupCategory

Get-ADGroupMember -Identity '<pilot-group>' |
  Select-Object Name,ObjectClass
```

Confirm that the pilot group is a security group, membership is intentionally limited, and each pilot user has a unique UPN. Keep raw names, distinguished names and GUIDs only in the gitignored local record or approved evidence store.

## 4. Validate LDAPS without changing AD

Identify at least two domain-controller FQDNs if available. From the GLPI host or the closest equivalent network context, verify DNS and TCP 636:

```powershell
Resolve-DnsName '<dc-fqdn>'
Test-NetConnection '<dc-fqdn>' -Port 636
```

Validate the certificate and bind path using the organisation's approved tool, such as `ldp.exe`:

1. Open **Connection → Connect**.
2. Enter the domain-controller FQDN and port `636`.
3. Select **SSL**, then connect.
4. Confirm the connection succeeds without a certificate warning.
5. Do not perform a write operation.

Record pass/fail for DNS, TCP 636, certificate trust, hostname validation and certificate expiry. Do not copy a private key or certificate bundle into the repository.

Confirm separately whether an AD administrator is authorized to create a dedicated read-only bind account later. Do not create it during information collection and never put its password in the intake record.

## 5. Validate the local GLPI endpoint

The proposed pilot endpoint is `https://glpi.test`. From every browser that will perform the pilot login or Microsoft OAuth authorization:

1. Open the URL in a private/incognito window.
2. Confirm the expected GLPI page loads.
3. Inspect the certificate in the browser.
4. Record whether the chain is trusted, the hostname matches and the certificate is not expired.

If there is a warning, record `FAIL`; do not bypass it for pilot acceptance. Install only the public organisation/local CA certificate through the approved workstation trust process. Never distribute its private key.

A local-only URL can support the selected LDAPS pilot. The OAuth mailbox redirect returns through the administrator's browser, so that browser must resolve `glpi.test` and trust its certificate.

## 6. Record GLPI version, plugins and existing users

### Runtime and plugin state

In GLPI, open the system information page and record the exact GLPI and PHP versions. From the GLPI application host, an administrator may also use the installed runtime's console commands:

```sh
php bin/console glpi:system:status
php bin/console glpi:plugin:list
```

When containerized, run the same commands inside the GLPI application container using the site's normal container tooling. Record only versions and plugin names/states; do not record environment variables.

For this architecture:

- `scim` and `oauthsso` are not required for the selected AD-to-GLPI LDAPS identity path;
- `oauthimap` is required for Microsoft 365 OAuth mailbox collection;
- confirm the authenticated Marketplace offers an official TECLIB'/GLPI-compatible `oauthimap` release for the installed GLPI/PHP version, but do not download or enable it yet.

### Existing-user reconciliation

1. Open **Administration → Users**.
2. Export or review: GLPI ID, login, email, authentication source, active state, entity and profiles.
3. Determine whether non-break-glass users already exist.
4. Classify current login format as UPN, `sAMAccountName`, mixed or other.
5. Compare the proposed pilot identities using this precedence:
   1. exactly one stable `objectGUID` correlation, when already present;
   2. for controlled bootstrap only, exactly one manually approved unique normalized UPN/email match.
6. Record counts and sanitized match labels. Do not put the user export in Git.

Fail closed if a UPN/email is missing or non-unique, stable ID and email candidates disagree, or more than one GLPI user could match. Do not merge, relink, deactivate or create users during collection.

## 7. Test break-glass access and document backup readiness

### Break-glass

1. Keep the current administrative session open.
2. Open a new private/incognito browser session.
3. Sign in using the local GLPI break-glass administrator—not AD or Microsoft authentication.
4. Confirm it reaches the expected administrative profile.
5. Sign out of the new session and record pass/fail and UTC time.

Never record the password. The repository deployment stores generated GLPI credentials in a gitignored mode-`0600` `.env`; move operational credentials to the approved password manager and encrypted off-host backup.

### Backup and restore

Before implementation, identify and test a process that captures:

- the complete GLPI database;
- GLPI files/configuration and uploaded documents required by the deployment;
- installed plugin files and versions; and
- the reverse-proxy configuration needed to restore the pilot URL.

Record the backup method, owner, encrypted off-host location reference, completion time and restore-test result. Do not record credentials, backup contents or a sensitive storage URL. A backup without a verified restore procedure does not pass the gate.

## 8. Define the mailbox pilot

The recommended scope is only `support@fluidhose.com`. Do not grant GLPI access to an administrator's entire personal mailbox. Messages received personally can be selectively forwarded to the support address after confirming the sender policy and avoiding forwarding loops.

An Exchange/Microsoft 365 administrator must record:

1. whether `support@fluidhose.com` exists and whether it is a shared or user mailbox;
2. the mailbox owner and change approver by role or sanitized label;
3. whether IMAP/OAuth access is permitted for this mailbox;
4. who can register an Entra application and grant the minimum required consent;
5. whether conditional-access policy permits the authorization flow;
6. the intended treatment of mail sent directly to the operator: no intake, manual forward, or an approved narrowly scoped mail-flow rule; and
7. a test sender domain/address policy, retention expectation and rollback owner.

Do not create the app registration, client credential, OAuth authorization or GLPI receiver during collection. Do not paste mailbox content, client IDs, tenant IDs, secrets or tokens into the repository.

## 9. Readiness decision

Information collection is complete only when the local record shows:

- hybrid synchronization and AD authority confirmed;
- LDAPS DNS, TCP 636 and certificate validation passed;
- one pilot AD security group and one or two pilot users approved;
- UPN uniqueness and `objectGUID` availability confirmed;
- local HTTPS trusted from the authorization browser;
- installed GLPI/PHP versions and plugin state recorded;
- existing GLPI users reconciled with conflicts at zero;
- fresh local break-glass login passed;
- GLPI backup and restore procedure passed;
- support-mailbox scope and responsible administrators confirmed; and
- rollback owners named by role or sanitized label.

When complete, tell the coding session that `scripts/it-management/.local/GLPI_M365_PILOT_INPUTS.md` is ready. The session can then read the local record, redact its reporting, and decide whether to create `glpi-m365-pilot-execution`. Operational evidence and secrets remain outside Git.
