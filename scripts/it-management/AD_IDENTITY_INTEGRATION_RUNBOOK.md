# AD identity integration runbook for GLPI and OpenProject

**Status:** Approved implementation baseline  
**Audience:** Systems administrators  
**Applies to:** GLPI, OpenProject 17.8, Active Directory, Microsoft Entra ID  
**Owner:** IT

## 1. Purpose

This runbook defines how GLPI and OpenProject use the organisation's existing identities without maintaining separate user directories.

- Active Directory (AD) is the authoritative source for identities when accounts originate on premises.
- Microsoft Entra Connect synchronizes those identities to Entra ID for Microsoft 365.
- GLPI is the system of record for support requests, assets and IT service-management data.
- OpenProject is the system of record for projects, milestones, dependencies and planned work.
- Project-specific membership and roles remain managed by OpenProject.

Passwords must never be copied or synchronized between application databases.

For GLPI computer inventory and ticket creation from `support@fluidhose.com`, follow [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md).

## 2. Target architecture

```text
                    Microsoft Entra Connect
Active Directory ------------------------------> Entra ID / Microsoft 365
      |                                                |
      | LDAPS                                          | optional future SSO
      +------------------> GLPI                        |
      |                                                |
      +------------------> OpenProject <---------------+

GLPI support ticket -------- linked reference -------- OpenProject project/work package
```

Use LDAPS as the initial common authentication mechanism. Entra-based OpenID Connect may replace it where the installed application edition supports that integration.

Keycloak is not required for the initial implementation. Introduce it only if a shared OIDC/SAML broker is needed and the additional operational responsibility is accepted.

## 3. System ownership

| Information | Authoritative system |
|---|---|
| Name, login, email, department and account status | AD |
| Microsoft 365 identity | Entra ID |
| Application access entitlement | AD security groups |
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
Preferred login: sAMAccountName or userPrincipalName
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

In GLPI, open **Setup → Authentication → LDAP directories** and add a directory using the Active Directory preset.

Use values equivalent to:

| Setting | Value |
|---|---|
| Server | Domain-controller FQDN |
| Port | `636` |
| Encryption | LDAPS/TLS enabled |
| Base DN | Organisation's AD base DN |
| Bind DN | DN of `svc_glpi_ldap` |
| Login field | `sAMAccountName` |
| Synchronization field | `objectGUID` |

`objectGUID` is the stable synchronization identifier and must be selected before importing users.

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
| Login attribute | `sAMAccountName` |
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

1. disable the new LDAP source or automatic account creation;
2. use the local break-glass administrator;
3. restore the previous authentication configuration;
4. do not delete newly created accounts until record ownership has been checked;
5. restore application data only when configuration rollback is insufficient.

## 12. References

- [GLPI LDAP directory configuration](https://help.glpi-project.org/documentation/modules/configuration/authentication/ldap)
- [GLPI authentication configuration](https://help.glpi-project.org/documentation/modules/configuration/authentication)
- [GLPI AD, inventory and email intake runbook](GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md)
- [OpenProject LDAP connections](https://www.openproject.org/docs/system-admin-guide/authentication/ldap-connections/)
- [OpenProject LDAP group synchronization](https://www.openproject.org/docs/system-admin-guide/authentication/ldap-connections/ldap-group-synchronization/)
- [OpenProject OpenID providers](https://www.openproject.org/docs/system-admin-guide/authentication/openid-providers/)
