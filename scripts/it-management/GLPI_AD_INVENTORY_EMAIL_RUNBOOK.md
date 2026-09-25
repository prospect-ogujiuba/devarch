# GLPI AD, inventory and email intake runbook

**Status:** Implementation baseline  
**Audience:** Systems administrators and help-desk administrators  
**Applies to:** GLPI 11, Active Directory, Microsoft Entra ID, Microsoft 365  
**Owner:** IT

## 1. Purpose

This runbook configures GLPI so that:

- Active Directory (AD) or Microsoft Entra ID remains authoritative for users and groups;
- GLPI Agent supplies current computer hardware and software inventory;
- GLPI Inventory optionally discovers network devices and printers; and
- email sent to `support@fluidhose.com` automatically creates and updates GLPI tickets.

Do not manually maintain identity or computer data in GLPI when an authoritative integration supplies it.

## 2. System ownership

| Information | Authoritative system |
|---|---|
| Login, name, email, department and account status | AD or Entra ID |
| Application access entitlement | AD or Entra security groups |
| Computer hardware, serial numbers, installed software and network details | GLPI Agent |
| Printers, switches and other SNMP-capable equipment | GLPI Inventory |
| Asset assignments, contracts, lifecycle and financial information | GLPI |
| Incidents, service requests, communication and SLAs | GLPI |
| Support mailbox | Microsoft 365 or the approved mail provider |

AD computer objects are not an inventory replacement. They normally identify only the domain object and do not provide dependable hardware, software, disk, serial-number or peripheral data.

## 3. Target architecture

```text
Active Directory -- LDAPS --> GLPI users and groups
       |
       +-- Group Policy --> GLPI Agent on Windows computers
                                  |
                                  +-- HTTPS inventory --> GLPI assets

Microsoft Entra ID -- SCIM --> GLPI users and groups       (cloud-only option)
Microsoft Entra ID -- OAuth SSO --> GLPI authentication    (cloud-only option)

support@fluidhose.com -- OAuth IMAP --> GLPI receiver --> ticket
GLPI -- authenticated SMTP --> requester and technician notifications
```

Use either the LDAPS identity path or the Entra SCIM/OAuth path as the primary provisioning design. Do not provision the same users independently through both paths without documented matching and conflict rules.

## 4. Configure identity provisioning

### 4.1 On-premises or hybrid AD

In GLPI, open **Setup → Authentication → LDAP directories**, select the Active Directory preset and configure:

| Setting | Recommended value |
|---|---|
| Server | Domain-controller FQDN; configure redundancy where supported |
| Port | `636` |
| Encryption | LDAPS/TLS with certificate validation |
| Base DN | Organisation's approved AD search root |
| Bind DN | Dedicated read-only GLPI LDAP service account |
| Login field | `sAMAccountName` |
| Synchronization field | `objectGUID` |
| Email | `mail` |
| First name | `givenName` |
| Last name | `sn` |
| User's groups | `memberOf` |
| Group members | `member` |
| Group name | `cn` |

Use this enabled-person filter:

```ldap
(&(objectClass=user)(objectCategory=person)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))
```

For a controlled rollout, add an `APP-GLPI-Users` membership condition using the group's actual distinguished name.

Implementation sequence:

1. Confirm the GLPI container trusts the AD certificate authority.
2. Confirm DNS resolution and TCP 636 connectivity from GLPI to at least two domain controllers.
3. Create a dedicated read-only bind account; deny interactive sign-in and store its password in the approved secret manager.
4. Test the LDAP connection.
5. Import the approved GLPI access groups.
6. Import and test one pilot user before bulk synchronization.
7. Map `APP-GLPI-Users` to Self-Service, `APP-GLPI-Technicians` to Technician and `APP-GLPI-Admins` to the approved administrative profile.
8. Configure disabled or removed directory users to be disabled rather than deleted in GLPI.
9. Schedule GLPI's LDAP user synchronization.
10. Retain and test the local GLPI break-glass administrator.

### 4.2 Cloud-only Microsoft Entra ID

For Entra-native identities:

1. Use the GLPI **SCIM** plugin to provision users and groups.
2. Use the GLPI **OAuth SSO** plugin for authentication.
3. Create authorization-assignment rules for the correct GLPI entity and profile.
4. When SCIM owns profile data, disable OAuth SSO's option to fetch and overwrite information from the user profile.

OAuth SSO authenticates users but does not replace provisioning. SCIM provisions identities but does not synchronize passwords.

## 5. Configure computer and equipment inventory

### 5.1 Native computer inventory

Deploy GLPI Agent to managed workstations and servers. Configure its server target as:

```text
https://<glpi-host>/front/inventory.php
```

Use HTTPS with a certificate trusted by each endpoint. In this repository's local environment, replace `<glpi-host>` with the reachable GLPI hostname; do not deploy the development-only `.test` hostname to production clients.

Recommended agent deployment:

- install the agent as a Windows service through Group Policy;
- supply `SERVER=https://<glpi-host>/front/inventory.php`;
- use `RUNNOW=1` during initial deployment;
- set an appropriate recurring inventory interval; and
- deploy first to a small computer OU or security group.

In GLPI, configure inventory rules to:

- assign imported computers to the correct entity;
- link inventory to an existing asset rather than creating duplicates;
- use stable identifiers such as UUID and serial number;
- associate the last logged-in directory username with the matching GLPI user where appropriate; and
- reject or quarantine inventory that cannot be assigned safely.

Do not use computer name as the sole matching identifier because names can be reused.

### 5.2 Network devices and printers

Install and enable **GLPI Inventory** only when network discovery, SNMP inventory, ESX inventory, data collection or remote deployment is required. Use a least-privilege SNMP credential, restrict discovery ranges and pilot each range before scheduling broad scans.

### 5.3 Inventory acceptance tests

Verify that:

- one pilot computer appears only once;
- manufacturer, model, UUID, serial number, disks, memory, operating system and installed software are populated;
- a renamed computer updates the existing asset;
- the correct user and entity are assigned;
- an agent rerun updates automatic fields without overwriting intentionally locked manual fields; and
- decommissioning behavior is documented and tested.

## 6. Configure `support@fluidhose.com` email intake

### 6.1 Mailbox preparation

Create `support@fluidhose.com` as a dedicated mailbox rather than relying only on a distribution list. Preserve the original sender address and normal internet message headers. Do not configure an automatic reply that loops back into GLPI.

For Microsoft 365:

1. Install and enable GLPI's **OAuth IMAP** plugin.
2. Create a Microsoft Entra app registration using the callback URL displayed by GLPI.
3. Grant only the documented IMAP permissions and complete administrator consent if required.
4. Create the client credential in the approved secret manager; do not store it in this repository.
5. Authorize the identity that can access `support@fluidhose.com`.

For another provider, use its approved TLS-enabled IMAP service and a dedicated credential.

### 6.2 GLPI receiver

Open **Setup → Receivers** and create a receiver with:

| Setting | Value |
|---|---|
| Name | `Fluid Hose Support` |
| Address | `support@fluidhose.com` |
| Protocol | IMAP with OAuth for Microsoft 365 |
| Incoming folder | Inbox |
| Accepted archive | Provider-approved processed folder, if used |
| Refused archive | Provider-approved rejected folder, if used |

Save the receiver, test its connection and retrieve a single controlled test message.

A successfully imported message maps as follows:

- subject becomes the ticket title;
- body becomes the ticket description;
- attachments become ticket documents; and
- recognized CC addresses can become observers.

### 6.3 Routing and sender policy

Open **Administration → Rules → Rules for assigning a ticket created through a mail receiver** and create an ordered rule for the `Fluid Hose Support` receiver. At minimum, assign:

- the production Fluid Hose entity;
- the Help Desk group; and
- a default category when organisational policy requires one.

Recommended sender policy:

1. Match the sender to the email address synchronized from AD or Entra ID.
2. Accept known GLPI users.
3. Optionally accept only the known `fluidhose.com` domain.
4. Leave anonymous ticket creation disabled unless external customers must open tickets.
5. If anonymous creation is enabled, add explicit routing, spam controls and rate monitoring.

Review **Setup → Receivers → Non-imported emails** when messages are rejected. Typical causes include an unknown sender, missing destination-entity rule, auto-response headers or a sender address identical to the receiver.

### 6.4 Automatic collection

In **Setup → Automatic actions**, configure `mailgate` as **Scheduled** in **CLI** mode. Run GLPI's scheduler every minute as the web-service account:

```cron
* * * * * php /path/to/glpi/front/cron.php
```

Force a controlled collection test with:

```bash
php /path/to/glpi/front/cron.php --force mailgate
```

For container deployment, run the same command inside the GLPI application container using the image's actual GLPI installation path. Verify the path after image upgrades rather than hard-coding an unverified container path.

## 7. Configure outbound notifications

Email intake does not configure outbound mail. Under GLPI notification settings:

1. Configure authenticated SMTP using the approved mailbox or relay.
2. Set the sender and reply-to behavior so replies return to the ticket workflow without creating loops.
3. Enable the ticket-created, follow-up, assignment and resolution notifications required by policy.
4. Verify requester and technician recipients for each notification template.
5. Test SPF, DKIM and DMARC alignment for the production sender domain.

For OAuth SMTP, the authenticated account normally must be permitted to send as the configured sender or alias.

## 8. End-to-end acceptance tests

Complete these tests before production rollout:

- an approved AD/Entra user signs in and receives the intended GLPI profile;
- a disabled or removed user loses access after synchronization;
- an AD name, email or department change updates without creating a duplicate;
- a GPO-deployed GLPI Agent creates one correctly assigned computer asset;
- repeated inventory updates the same asset;
- email from a known user to `support@fluidhose.com` creates one ticket in the correct entity and group;
- the subject, body and safe test attachment are imported correctly;
- GLPI sends a ticket confirmation to the requester;
- a requester reply is attached to the existing ticket rather than opening a duplicate;
- an unauthorized or malformed message follows the documented refusal policy;
- `mailgate` continues running without an interactive GLPI session; and
- LDAP, inventory and mailbox failures generate actionable monitoring alerts.

## 9. Operations and security

- Monitor LDAP synchronization, agent check-ins, `mailgate`, SMTP delivery and non-imported mail.
- Rotate LDAP, OAuth and SMTP credentials under change control and retest afterward.
- Never commit bind passwords, OAuth client secrets, mailbox credentials or private certificates.
- Review privileged AD group mappings at least quarterly.
- Keep GLPI, GLPI Agent and installed plugins on supported versions.
- Back up GLPI before plugin upgrades or large identity and inventory imports.
- Test local break-glass access during scheduled recovery exercises.

## 10. References

- [GLPI LDAP directory configuration](https://help.glpi-project.org/documentation/modules/configuration/authentication/ldap)
- [GLPI computer inventory](https://help.glpi-project.org/tutorials/inventory/computer_inventory)
- [Deploying GLPI Agent through GPO](https://help.glpi-project.org/tutorials/inventory/deploy_agent_gpo)
- [GLPI Inventory plugin](https://help.glpi-project.org/doc-plugins/plugins-glpi/glpi-inventory)
- [GLPI receivers](https://help.glpi-project.org/documentation/modules/configuration/collectors)
- [GLPI automatic actions](https://help.glpi-project.org/documentation/modules/configuration/crontasks)
- [Microsoft Entra OAuth IMAP receiver](https://help.glpi-project.org/tutorials/receivers/oauth_imap_entra)
- [GLPI SCIM guidance](https://help.glpi-project.org/faq/plugins/scim)
- [GLPI authentication and SSO guidance](https://help.glpi-project.org/faq/plugins/authentication-and-sso)
