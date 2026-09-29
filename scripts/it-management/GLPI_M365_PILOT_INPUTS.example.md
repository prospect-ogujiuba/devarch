# GLPI Microsoft 365 pilot input record

> Copy to `.local/GLPI_M365_PILOT_INPUTS.md` and complete it there. The `.local/` copy is gitignored. Do not enter passwords, bind credentials, OAuth secrets, tokens, private keys, certificate files, tenant IDs, personal email content or screenshots.

**Collection owner:** `<role or sanitized label>`

**Started (UTC):** `<YYYY-MM-DD HH:MM>`

**Evidence store reference:** `<non-sensitive reference only>`

## 1. Scope and decisions

- [ ] Endpoint agent/inventory deployment is excluded.
- [ ] On-premises AD is authoritative.
- [ ] The only pilot provisioning/authentication path is AD → GLPI over LDAPS.
- [ ] SCIM provisioning will remain disabled.
- [ ] GLPI login attribute will be `userPrincipalName`.
- [ ] Stable synchronization attribute to validate is `objectGUID`.
- [ ] Mail intake is included.
- Mailbox scope: `<support mailbox only / other approved scope>`
- Personal-mail handling: `<none / manual selective forward / approved narrow rule>`

## 2. Hybrid identity confirmation

- Entra Connect configured and healthy: `<PASS/FAIL/UNKNOWN>`
- Pilot user on-premises sync enabled: `<YES/NO>`
- Pilot user source indicates Windows Server AD: `<YES/NO>`
- UPN equals Microsoft 365 sign-in name: `<YES/NO>`
- Evidence date (UTC): `<date/time>`
- Evidence reference: `<non-sensitive reference>`
- Notes: `<no secrets or tenant ID>`

## 3. LDAPS readiness

| Check | Result | Evidence date/reference |
|---|---|---|
| Primary DC DNS resolution | `<PASS/FAIL>` | `<reference>` |
| Primary DC TCP 636 | `<PASS/FAIL>` | `<reference>` |
| Secondary DC DNS resolution | `<PASS/FAIL/N/A>` | `<reference>` |
| Secondary DC TCP 636 | `<PASS/FAIL/N/A>` | `<reference>` |
| LDAPS certificate chain trusted | `<PASS/FAIL>` | `<reference>` |
| Certificate hostname valid | `<PASS/FAIL>` | `<reference>` |
| Certificate not expired | `<PASS/FAIL>` | `<reference>` |
| GLPI runtime/container trusts issuing CA | `<PASS/FAIL>` | `<reference>` |

- Dedicated read-only bind account creation authorized: `<YES/NO>`
- Bind-account owner: `<role or sanitized label>`
- Bind password location: `<approved password-manager reference; never the password>`

## 4. Pilot group and users

- Pilot AD security-group label: `<sanitized label>`
- Group scope/category confirmed: `<value>`
- Membership approved by: `<role or sanitized label>`

| Pilot label | Enabled | Unique UPN | M365 sign-in matches UPN | `objectGUID` available |
|---|---|---|---|---|
| `pilot-user-1` | `<YES/NO>` | `<YES/NO>` | `<YES/NO>` | `<YES/NO>` |
| `pilot-user-2` | `<YES/NO/N/A>` | `<YES/NO/N/A>` | `<YES/NO/N/A>` | `<YES/NO/N/A>` |

Keep actual UPNs, DNs and GUIDs here only when operationally necessary; never copy this completed file into Git.

## 5. Local HTTPS

- Pilot URL: `<internal URL>`
- Resolves from pilot browser: `<PASS/FAIL>`
- GLPI page loads: `<PASS/FAIL>`
- Certificate chain trusted: `<PASS/FAIL>`
- Hostname matches: `<PASS/FAIL>`
- Certificate not expired: `<PASS/FAIL>`
- Tested from Microsoft authorization browser: `<PASS/FAIL>`
- Evidence date/reference: `<date and non-sensitive reference>`

## 6. GLPI current state

- GLPI version: `<version>`
- PHP version: `<version>`
- Current plugin states recorded: `<YES/NO>`
- Official compatible `oauthimap` available: `<YES/NO/UNKNOWN>`
- Marketplace/source and displayed publisher: `<non-sensitive description>`
- `oauthimap` version offered: `<version>`
- Plugin installed during collection: `NO`

### Existing users

- Total active users: `<count>`
- Local break-glass administrators: `<count>`
- Other local users: `<count>`
- LDAP users: `<count>`
- Other external users: `<count>`
- Current login convention: `<UPN/sAMAccountName/mixed/other>`
- Unique pilot matches: `<count>`
- Missing matches: `<count>`
- Conflicting/ambiguous matches: `<count>`
- Reconciliation disposition: `<PASS/BLOCKED>`
- Evidence reference: `<non-sensitive reference; do not attach export>`

## 7. Break-glass and backup

- Fresh local break-glass login: `<PASS/FAIL>`
- Administrative profile confirmed: `<YES/NO>`
- Test time (UTC): `<date/time>`
- Break-glass owner: `<role or sanitized label>`

- Database backup method identified: `<YES/NO>`
- GLPI files/configuration backup method identified: `<YES/NO>`
- Reverse-proxy configuration included: `<YES/NO>`
- Encrypted off-host destination confirmed: `<YES/NO>`
- Restore procedure tested: `<PASS/FAIL>`
- Backup/restore owner: `<role or sanitized label>`
- Evidence reference: `<non-sensitive reference>`

## 8. Microsoft 365 mailbox readiness

- `support@fluidhose.com` exists: `<YES/NO>`
- Mailbox type: `<shared/user>`
- Mailbox/change owner: `<role or sanitized label>`
- IMAP/OAuth permitted: `<YES/NO/UNKNOWN>`
- Entra app-registration operator available: `<YES/NO>`
- Required consent approver available: `<YES/NO>`
- Conditional-access review completed: `<YES/NO>`
- Personal mailbox excluded from direct GLPI access: `<YES/NO>`
- Forwarding-loop prevention reviewed: `<YES/NO/N/A>`
- Sender acceptance/refusal policy approved: `<YES/NO>`
- Mail rollback owner: `<role or sanitized label>`
- Evidence reference: `<non-sensitive reference>`

## 9. Rollback ownership

- Identity rollback owner: `<role or sanitized label>`
- GLPI rollback owner: `<role or sanitized label>`
- Mailbox/OAuth rollback owner: `<role or sanitized label>`
- Queue-drain evidence location: `<non-sensitive reference>`
- Exceptions requiring approval: `<none or summary>`

## 10. Completion

- [ ] Every required result is complete.
- [ ] All failures and unknowns have an owner and disposition.
- [ ] No secrets or sensitive evidence are present in Git.
- [ ] The completed local record is ready for read-only review.

**Collection result:** `<READY/BLOCKED>`

**Completed (UTC):** `<YYYY-MM-DD HH:MM>`

**Approved by:** `<role or sanitized label>`
