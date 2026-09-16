# openMAINT demo

Local evaluation stack for openMAINT 2.4.2 / CMDBuild 4.2.0. It uses the community-maintained `itmicus/cmdbuild` image and loads its demonstration database.

## Start

```bash
cd services-library/management/openmaint
cp runtime.env.example runtime.env
# Replace both placeholder passwords in runtime.env.
chmod 600 runtime.env
podman compose up -d
```

Initial database import and Tomcat startup can take several minutes. Check readiness with:

```bash
podman ps --filter name=openmaint
podman logs --tail 50 openmaint
```

Open <https://openmaint.test> and sign in with the image's demo account. The direct localhost fallback is <http://127.0.0.1:9219/cmdbuild/ui/>.

- Username: `admin`
- Password: `admin`

Change the application password before entering any non-sample information. The database has no host port, and the web UI is bound to localhost, but this stack still uses demo credentials and is **not production-ready**. Do not enter customer-confidential audit evidence. `runtime.env` is local-only and ignored by Git; preserve it while the named volumes exist because its passwords initialize the database.

## Stop or reset

```bash
podman compose down       # preserves data
podman compose down -v    # deletes the demo database and application volume
```

For a production evaluation, review image provenance, licensing, backups, TLS/reverse proxying, SSO, mail, database credentials, attachment storage, resource limits, monitoring, and upgrade/rollback procedures.

Source deployment reference: <https://github.com/itmicus/cmdbuild_docker/tree/29378aff/openmaint-2.4.2-4.2.0>
