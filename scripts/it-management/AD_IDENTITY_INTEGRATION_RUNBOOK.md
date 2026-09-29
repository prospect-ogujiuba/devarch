# AD identity integration runbook for GLPI and OpenProject

**Status:** Approved implementation baseline  
**Audience:** Systems administrators  
**Applies to:** GLPI, OpenProject 17.8, Active Directory, Microsoft Entra ID  
**Owner:** IT

## 1. Purpose

This runbook defines how GLPI and OpenProject use the organisation's existing identities without maintaining separate user directories.

- Active Directory (AD) is the organisational authoritative source for identity attributes when accounts originate on premises.
- Microsoft Entra Connect synchronizes those identities to Entra ID for Microsoft 365. Entra may still be selected as GLPI's provisioning/authentication **integration boundary**; that interface choice does not transfer ownership of hybrid identity attributes away from AD.
- GLPI is the system of record for support requests, assets and IT service-management data.
- OpenProject is the system of record for projects, milestones, dependencies and planned work.
- Project-specific membership and roles remain managed by OpenProject.

Passwords must never be copied or synchronized between application databases.

For GLPI computer inventory and ticket creation from `support@fluidhose.com`, follow [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md).

## 2. Target architecture

```text
Hybrid source:  Active Directory -- Entra Connect --> Entra ID / Microsoft 365
                       |                                  |
                       | optional LDAPS                   | SCIM + OAuth SSO
                       +---------------> GLPI <-----------+
                       |
                       +---------------> OpenProject

Cloud-only source: Entra ID / Microsoft 365 -- SCIM + OAuth SSO --> GLPI

GLPI support ticket -------- linked reference -------- OpenProject project/work package
```

Choose one GLPI provisioning path. Use Entra SCIM plus OAuth SSO for cloud-only identities and when Entra is the deliberate GLPI integration boundary. Use LDAPS when identities originate in on-premises AD and GLPI has secure domain-controller connectivity. Never provision the same GLPI population independently through both paths. For every GLPI SCIM/OAuth change, the GLPI-specific runbook's [identity rollback and recovery test](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md#45-identity-rollback-and-recovery-test) is mandatory and authoritative; this broader runbook's generic rollback does not replace it.

Keycloak is not required for the initial implementation. Introduce it only if a shared OIDC/SAML broker is needed and the additional operational responsibility is accepted.

## 3. System ownership

| Information | Authoritative system |
|---|---|
| Name, login, email, department and account status | AD for hybrid identities; Entra ID for cloud-only identities. A hybrid GLPI deployment may consume the synchronized values through Entra without making Entra the upstream authoring source. |
| Microsoft 365 identity | Entra ID |
| Application access entitlement | AD-synchronized or Entra security groups, according to the selected path |
| Assets, contracts and equipment assignments | GLPI |
| Incidents, requests and support SLAs | GLPI |
| GLPI profiles and technician permissions | GLPI, mapped from AD groups where practical |
| Projects, tasks, dates, estimates and dependencies | OpenProject |
| Project-specific membership and roles | OpenProject |
| Cross-system identifiers and links | Integration service |

A field must have only one authoritative owner. Application-local changes must not overwrite AD-managed identity attributes.

## 4. Prerequisites

Record the following non-secret values before configuration:

```text
AD DNS domain:
Base DN:
Users OU DN:
Groups OU DN:
Domain controller FQDN:
LDAPS port: 636
Nested AD groups used: yes/no
Pilot login standard: `userPrincipalName`. If existing GLPI users use `sAMAccountName`, stop and complete the documented reconciliation/migration under a separate approved change before import.
OpenProject Enterprise token available: yes/no
```

Do not store bind passwords, client secrets, private keys or private certificates in this repository.

Operational prerequisites:

1. LDAPS is enabled on at least two domain controllers.
2. GLPI and OpenProject containers trust the issuing AD certificate authority.
3. DNS resolution and TCP 636 connectivity work from both containers.
4. AD users have unique, valid email addresses.
5. A tested backup and rollback procedure exists for both applications.
6. Local break-glass administrator accounts are retained in both applications.

## 5. Active Directory preparation

### 5.1 Service accounts

Create separate, read-only bind accounts:

- `svc_glpi_ldap`
- `svc_openproject_ldap`

Requirements:

- no Domain Admin or interactive sign-in rights;
- read access only to the required user and group attributes;
- long, managed passwords stored in the approved secret manager;
- monitored password expiry and rotation;
- separate credentials so either integration can be revoked independently.

### 5.2 Access groups

Create application-specific security groups:

- `APP-GLPI-Users`
- `APP-GLPI-Technicians`
- `APP-GLPI-Admins`
- `APP-OpenProject-Users`
- `APP-OpenProject-Admins`

Add only a small pilot cohort initially. Department groups may supply organisational metadata, but must not automatically grant administrative permissions.

## 6. Configure GLPI

For an Entra SCIM/OAuth design, do not follow the LDAPS procedure below. Follow sections 4.1 through 4.5 of the [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md#4-configure-identity-provisioning), including its exact correlation, consent/session revocation and rollback tests.

For the selected LDAPS design, open **Setup → Authentication → LDAP directories** in GLPI and add a directory using the Active Directory preset.

Use values equivalent to:

| Setting | Value |
|---|---|
| Server | Domain-controller FQDN |
| Port | `636` |
| Encryption | LDAPS/TLS enabled |
| Base DN | Organisation's AD base DN |
| Bind DN | DN of `svc_glpi_ldap` |
| Login field | `userPrincipalName` for this Microsoft 365-aligned pilot |
| Synchronization field | `objectGUID` |

`objectGUID` is selected as the stable AD synchronization identifier, subject to the installed GLPI core/version validation in the GLPI-specific runbook. The selected pilot login attribute is `userPrincipalName`. If existing GLPI users use `sAMAccountName`, stop and reconcile/migrate them under a separate approved change before import; do not switch the pilot standard silently.

Use the standard enabled-user filter:

```ldap
(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))
```

During the pilot, additionally restrict access to `APP-GLPI-Users`:

```ldap
(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(memberOf=CN=APP-GLPI-Users,OU=Groups,DC=example,DC=local))
```

Replace the example group DN with the actual DN. Enter the filter on one line.

Configure group lookup using:

| Purpose | AD attribute |
|---|---|
| User's groups | `memberOf` |
| Group members | `member` |
| Group name | `cn` |

Then:

1. Test the LDAP connection.
2. Import the pilot access groups.
3. Import one pilot user.
4. Map `APP-GLPI-Users` to the Self-Service profile.
5. Map `APP-GLPI-Technicians` to the Technician profile.
6. Map `APP-GLPI-Admins` only to the approved administrative profile.
7. Verify authentication, profile assignment and group removal.
8. Schedule `glpi:ldap:synchronize_users` using the GLPI application scheduler or an approved host scheduler.
9. Configure disabled or removed AD users to be disabled in GLPI, not deleted.

Retain and test the local `glpi` break-glass administrator.

## 7. Configure OpenProject

In OpenProject, open **Administration → Authentication → LDAP connections** and create a connection.

| Setting | Value |
|---|---|
| Host | Domain-controller FQDN |
| Port | `636` |
| Encryption | LDAPS/SSL with certificate verification |
| Account | DN of `svc_openproject_ldap` |
| Base DN | Users OU or approved search root |
| Login attribute | `userPrincipalName` for this Microsoft 365-aligned pilot; an existing `sAMAccountName` deployment requires a separately approved reconciliation/migration before import |
| First name | `givenName` |
| Last name | `sn` |
| Email | `mail` |

Restrict eligible users to `APP-OpenProject-Users` during the pilot:

```ldap
(&(objectCategory=person)(objectClass=user)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(memberOf=CN=APP-OpenProject-Users,OU=Groups,DC=example,DC=local))
```

Enable automatic user creation only after existing accounts have been reconciled. Test login with a pilot user and verify that name and email synchronize correctly.

LDAP group synchronization is an OpenProject Enterprise add-on. If it is unavailable:

- use AD groups only to control application access;
- maintain project-specific groups and roles in OpenProject; or
- implement a narrow, one-way group synchronizer through the supported OpenProject API.

Do not write directly to the OpenProject database. Retain and test the local OpenProject break-glass administrator.

## 8. Reconcile existing imported users

The workbook import created application-local accounts whose login names may not match AD. Enabling automatic LDAP creation without reconciliation can create duplicate people and split ticket or work-package ownership.

Before enabling LDAP for the wider user base:

1. Export existing GLPI and OpenProject users.
2. Match them to AD using verified email and employee ID values.
3. Flag ambiguous, duplicate and placeholder email addresses for manual review.
4. Link or convert the existing application account to LDAP authentication where supported.
5. Verify that existing tickets, assets and work packages still reference the correct user.
6. Test one active employee, one department transfer and one disabled employee.
7. Record all unmatched accounts before rollout.
8. Stop using the workbook importer to maintain AD-owned identity attributes.

Display names alone must never be used as matching keys.

## 9. GLPI-to-OpenProject work handoff

A support record remains in GLPI when it is an incident, routine request or short-lived change. Create linked OpenProject work when it has defined deliverables, multiple teams, dependencies, planned dates, or a duration measured in weeks or months.

Handoff process:

1. Create and triage the request in GLPI.
2. Set an approved classification such as **Escalated to project**.
3. Create an OpenProject project or work package through its supported API.
4. Store the OpenProject identifier and URL in GLPI.
5. Store the originating GLPI ticket identifier and URL in OpenProject.
6. Manage tasks, dates and delivery status only in OpenProject.
7. Return only a status summary and completion reference to GLPI.
8. Resolve the GLPI ticket after the agreed project deliverable is accepted.

Do not duplicate GLPI assets or routine tickets as OpenProject work packages. Do not perform bidirectional updates on the same field.

## 10. Pilot and acceptance tests

Complete these tests before production rollout:

- approved pilot user can sign in to both applications;
- user outside the access group cannot sign in;
- disabled AD user loses access after synchronization;
- name, email and department changes propagate without duplication;
- GLPI technician and administrator mappings grant only intended permissions;
- OpenProject project roles remain project-scoped;
- removal from an access group has the documented result;
- LDAPS certificate validation succeeds from both containers;
- service-account password rotation is tested;
- local break-glass login works while LDAP is unavailable;
- existing imported records retain their correct owners;
- backup restoration is tested or covered by a current recovery exercise.

## 11. Rollout and rollback

Roll out in stages:

1. IT administrators.
2. Selected technicians and project managers.
3. One representative department.
4. Remaining authorised users.

During each stage, monitor authentication failures, duplicate accounts, missing memberships and unexpected privilege assignments.

If rollout fails:

1. stop the active directory-provisioning job or LDAP synchronization before changing mappings;
2. prove a new local break-glass login works;
3. disable the new SSO/LDAP source and pilot assignment, then restore the recorded authentication and authorization configuration;
4. revoke or rotate integration tokens, OAuth secrets and bind credentials after the integration is disabled;
5. preserve existing application user IDs and ownership; disable suspect new accounts and do not delete, merge or relink them until ticket, asset and project ownership has been audited;
6. compare users, profiles, memberships and record owners with the pre-change export; and
7. restore application data only inside an approved rollback window when configuration rollback is insufficient and post-backup changes have been accounted for.

For GLPI, the authoritative detailed procedure is **Identity rollback and recovery test** in [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md#45-identity-rollback-and-recovery-test). Rehearse it during the pilot; this broader runbook does not replace its SCIM/OAuth-specific stop, token-revocation and ownership checks.

## 12. References

- [GLPI LDAP directory configuration](https://help.glpi-project.org/documentation/modules/configuration/authentication/ldap)
- [GLPI authentication configuration](https://help.glpi-project.org/documentation/modules/configuration/authentication)
- [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md)
- [OpenProject LDAP connections](https://www.openproject.org/docs/system-admin-guide/authentication/ldap-connections/)
- [OpenProject LDAP group synchronization](https://www.openproject.org/docs/system-admin-guide/authentication/ldap-connections/ldap-group-synchronization/)
- [OpenProject OpenID providers](https://www.openproject.org/docs/system-admin-guide/authentication/openid-providers/)
