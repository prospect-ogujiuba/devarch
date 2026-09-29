# GLPI Microsoft 365 current-state and access assessment

## Current repository deployment

These are time-bounded local observations from **2026-09-29**, not durable production facts. Re-run the non-secret checks before using them for a change decision:

```shell
podman ps -a --format '{{.Names}}\t{{.Status}}\t{{.Image}}' | grep -i glpi
podman exec glpi php -v
podman exec glpi php bin/console glpi:plugin:list
```

Observed results:

- The local `glpi` container reported `docker.io/glpi/glpi:11.0.9`, and `glpi` plus `glpi-db` reported healthy/running status.
- The documented endpoint is `https://glpi.test`, routed through the repository's local Nginx Proxy Manager environment. This development-only hostname is not suitable for production agents, Entra SCIM, or OAuth redirect URIs.
- `glpi:plugin:list` returned an empty plugin table. SCIM and OAuth SSO were therefore not available in the observed local instance. Production identity work is blocked until the authenticated GLPI Marketplace confirms compatible `scim` and `oauthsso` releases for the revalidated GLPI/PHP versions and the required subscription entitlement.
- Existing runbooks describe both LDAPS and Entra-native designs, but the tenant topology has not been identified.

## Recommended direction

Use GLPI Agent. It provides materially better endpoint inventory than Entra device records alone, including hardware identity, components, OS, storage, software, interfaces, and many attached displays and installed printers. It is worth asking the MSP to deploy after a small pilot.

A workstation inventory commonly detects directly connected monitors through EDID/display data and printers installed or visible to the operating system. Coverage is not guaranteed: docks, KVMs, generic display drivers, sleeping/disconnected monitors, shared print queues, and some USB/network devices may hide serial numbers or relationships. Network-printer consumables, counters, and network equipment require GLPI Inventory discovery/SNMP rather than endpoint inventory alone.

## Identity architecture decision

- **Cloud-only Microsoft 365 / Entra identities:** do not configure LDAP. Use the GLPI SCIM plugin for lifecycle provisioning and the OAuth SSO plugin for authentication. Entra groups scope access; GLPI authorization rules map approved users/groups to entities and profiles.
- **Hybrid identities originating in on-premises AD and synchronized by Entra Connect:** use LDAPS from GLPI to AD when GLPI can securely reach redundant domain controllers, or deliberately choose Entra SCIM/OAuth as the sole GLPI path. Do not provision the same population independently through both paths.

SCIM does not authenticate users, and OAuth SSO does not replace provisioning. Existing GLPI accounts must be reconciled before wider provisioning to prevent duplicate people and split ticket ownership. Matching precedence is: (1) exactly one existing GLPI user carrying the same persisted Entra-object-ID/SCIM-`externalId` correlation, or AD `objectGUID` for LDAPS; then only for a controlled bootstrap, (2) exactly one manually approved GLPI user matched by unique normalized UPN/primary email. The existing GLPI ID must be preserved, and an email candidate never overrides a stable-ID match. Missing/non-unique values, disagreement between identifier and email candidates, out-of-scope identity types, or more than one GLPI candidate fail closed with no automatic create, merge, relink or deactivation. The pilot must prove that the exact installed plugin version can adopt an existing GLPI user ID without creating a second record; otherwise rollout stops pending a vendor-supported migration procedure.

## Access likely required

### Endpoint deployment through Intune

- PCs must be enrolled and managed by Intune.
- The operator needs an Intune role with Mobile apps create/read/update/assign permissions; the built-in **Application Manager** role is the usual least-privilege starting point.
- Scope groups/tags and any Multi Admin Approval policy must permit the pilot assignment.
- If these permissions are unavailable, the MSP can package and deploy the agent through its RMM while following the same pilot and detection requirements.

Merely having access to Entra does not imply permission to deploy software.

### Entra provisioning and SSO

- **Application Administrator** or **Cloud Application Administrator** can normally create/configure the Enterprise Application, SCIM provisioning job and app registration during setup. This duty does not require broad Graph application permissions for Entra-to-GLPI SCIM.
- OAuth SSO should request only the installed plugin's documented delegated `openid`, `profile`, `email`, `offline_access` and `User.Read` scopes; do not grant `User.ReadWrite.All` or `Group.ReadWrite.All`.
- Tenant consent is a separate reviewed duty performed only by an authorized consent administrator under tenant policy. An Enterprise Application/App Registration Owner may then operate the already-owned app with narrower scope.
- Use a pilot Entra security group for assignment. Do not assign the enterprise application tenant-wide initially.

### Hybrid LDAPS

- An AD administrator must create the access groups and a dedicated read-only bind account.
- Network/DNS administrators must provide GLPI-to-domain-controller DNS and TCP 636 connectivity and the issuing CA chain.
- The bind account must not be a Domain Admin and must not have interactive sign-in rights.

## Information required before tenant changes

1. Are users cloud-only, or does Microsoft 365 show synchronization from on-premises AD/Entra Connect?
2. Do target PCs appear in Intune as managed devices?
3. Can the operator create and assign a Windows app in Intune, or will the MSP deploy through RMM?
4. What production HTTPS hostname will replace `glpi.test`, and will it be reachable by managed PCs and Microsoft's Entra provisioning service?
5. Is a GLPI Network/Marketplace entitlement available for the SCIM and OAuth SSO plugins?
6. Which small Entra/AD group and two or three representative PCs will form the pilot?

## Production gates

This assessment and the runbooks prepare a rollout; they do not assert that the current environment is production-ready. Tenant topology, production URL, plugin availability/entitlement, pilot identities and deployment authority remain unresolved operator decisions.

Do not deploy broadly or enable SSO until:

- cloud-only versus hybrid topology and the single selected GLPI provisioning path are recorded;
- a stable production HTTPS URL and trusted certificate exist;
- the exact official plugin packages, versions, publisher, GLPI/PHP compatibility and entitlement are verified in GLPI Marketplace;
- backup, tested local break-glass login and identity rollback rehearsal are confirmed;
- existing GLPI users are reconciled with an exact tested correlation mapping;
- SCIM/OAuth or LDAPS works for a pilot user;
- one pilot device updates a single computer asset without duplicates;
- monitor and printer observations are compared with the physical workstation; and
- uninstall/rollback and loss-of-directory-access behavior are tested.

## Pilot evidence record required before rollout

No pilot control has yet been asserted as passed. The change owner must complete a dated evidence record before expansion. For every row, record UTC timestamp, operator, tenant/GLPI environment identifier, exact command or UI path, bounded non-secret output/screenshot reference, result, exception owner and approval. Do not store tokens, credentials, certificates or personal mailbox content.

| Gate | Reproducible evidence required | Pass condition |
|---|---|---|
| Identity topology and correlation | Entra/AD source indicator; selected SCIM or LDAPS path; exported pilot mapping of GLPI ID to Entra object ID/SCIM `externalId` or AD `objectGUID`; rename test event IDs | One authoritative path; existing GLPI ID survives; conflicts are quarantined; no duplicate |
| Provisioning queue | Entra provisioning logs or GLPI LDAP automatic-action/import logs spanning the recorded barrier timestamp | Zero queued/running operations; every pilot event has a succeeded/failed/canceled/quarantined terminal result and disposition |
| OAuth and least privilege | Entra enterprise-app/app-registration role assignment and consent UI paths; exported permission list with secrets redacted; successful pilot and break-glass login audit events | Only documented delegated scopes/roles; no broad Graph/application access; local break-glass works |
| Endpoint deployment and inventory | Intune or MSP/RMM device-job page; MSI detection/uninstall result; agent and GLPI inventory logs; GLPI asset ID/UUID plus physical monitor/printer comparison | Management and inventory queues are zero; each submission is terminal; one asset, expected software/hardware, no duplicate |
| Mail collection | GLPI receiver and `mailgate` automatic-action status/log; controlled message ID and accepted/refused disposition with content redacted | Action not running/stuck; no retry backlog; every controlled message accounted for |
| Fail-closed rollback | Backup identifier; SCIM/LDAPS stop logs; endpoint uninstall/queue drain; OAuth revocation and fresh-session denial; post-barrier comparison | No post-barrier mutation or unauthorized login; ownership preserved; exception list empty or formally accepted |

The evidence record is a production change artifact and must remain outside Git when it contains tenant identifiers or operational screenshots.

## Claim-level sources

External sources below were retrieved or reviewed on **2026-09-29** for planning against GLPI **11.0.9**. Live vendor documentation and the authenticated in-application Marketplace must be revalidated at deployment time.

- Repository state and local deployment: `scripts/project-management/README.md`, `scripts/it-management/GLPI_AD_INVENTORY_EMAIL_RUNBOOK.md`, `scripts/it-management/AD_IDENTITY_INTEGRATION_RUNBOOK.md`, the running `docker.io/glpi/glpi:11.0.9` container, and `php bin/console glpi:plugin:list`.
- GLPI SCIM provisioning behavior and the requirement to pair Entra provisioning with OAuth SSO authentication: <https://help.glpi-project.org/faq/plugins/scim>
- GLPI OAuth SSO requirements, Entra configuration and GLPI Network entitlement: <https://help.glpi-project.org/doc-plugins/fr/plugin-glpi-network/oauthsso> and <https://help.glpi-project.org/faq/plugins/authentication-and-sso>
- GLPI Marketplace compatibility behavior and installed-version gating: <https://help.glpi-project.org/documentation/modules/configuration/plugins>
- GLPI LDAP configuration and installed-core validation basis for `objectGUID`: <https://help.glpi-project.org/documentation/modules/configuration/authentication/ldap>
- GLPI Agent endpoint inventory categories: <https://glpi-agent.readthedocs.io/en/latest/man/glpi-agent.html>
- GLPI computer, monitor, printer and network-equipment asset behavior: <https://help.glpi-project.org/documentation/modules/assets/computers>, <https://help.glpi-project.org/documentation/modules/assets/monitors>, <https://help.glpi-project.org/documentation/modules/assets/printers>, and <https://help.glpi-project.org/documentation/modules/assets/network-equipments>
- Intune Application Manager/RBAC and Win32 app packaging/assignment: <https://learn.microsoft.com/en-us/intune/intune-service/fundamentals/role-based-access-control> and <https://learn.microsoft.com/en-us/intune/app-management/deployment/add-win32>
- Entra SCIM provisioning and application administration: <https://learn.microsoft.com/en-us/entra/identity/app-provisioning/use-scim-to-provision-users-and-groups>, <https://learn.microsoft.com/en-us/entra/identity/app-provisioning/how-provisioning-works>, and <https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference>
- Official OAuth IMAP package and GLPI 11 manifest: <https://github.com/pluginsGLPI/oauthimap> and <https://github.com/pluginsGLPI/oauthimap/blob/main/oauthimap.xml>
- Exchange Online delegated IMAP/SMTP OAuth and app-only mailbox scoping: <https://learn.microsoft.com/en-us/exchange/client-developer/legacy-protocols/how-to-authenticate-an-imap-pop-smtp-application-by-using-oauth>, <https://learn.microsoft.com/en-us/exchange/clients-and-mobile-in-exchange-online/authenticated-client-smtp-submission>, and <https://learn.microsoft.com/en-us/exchange/permissions-exo/application-rbac>
