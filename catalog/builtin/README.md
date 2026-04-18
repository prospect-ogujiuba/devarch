# Builtin Catalog

This tree contains first-party DevArch V2 templates owned by `surgeon-catalog`.

## Canonical layout

Each builtin template lives under a category/name directory and uses the canonical filename:

```text
catalog/builtin/<category>/<template-name>/template.yaml
```

Current seeded corpus:

- `catalog/builtin/database/postgres/template.yaml`
- `catalog/builtin/cache/redis/template.yaml`
- `catalog/builtin/backend/node-api/template.yaml`
- `catalog/builtin/backend/laravel-app/template.yaml`
- `catalog/builtin/frontend/vite-web/template.yaml`
- `catalog/builtin/proxy/nginx/template.yaml`

Templates stay plain-file, deterministic, and human-readable so later catalog discovery and indexing can treat this tree as committed source-of-truth data.
