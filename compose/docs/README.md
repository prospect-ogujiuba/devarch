# Documentation Platforms

Knowledge management and documentation services for DevArch.

## Services

### Wiki.js (Modern wiki platform)
- **File**: `wikijs.yml`
- **Port**: 10000
- **Dependencies**: PostgreSQL (database: wikijs)
- **Usage**:
  ```bash
  docker compose -f compose/docs/wikijs.yml up -d
  ```
- **Access**: http://localhost:10000

### BookStack (Documentation platform)
- **File**: `bookstack.yml`
- **Port**: 10010
- **Dependencies**: MySQL/MariaDB (database: bookstack)
- **Usage**:
  ```bash
  docker compose -f compose/docs/bookstack.yml up -d
  ```
- **Access**: http://localhost:10010

### Outline (Team knowledge base)
- **File**: `outline.yml`
- **Port**: 10020
- **Dependencies**: PostgreSQL (database: outline), Redis
- **Usage**:
  ```bash
  docker compose -f compose/docs/outline.yml up -d
  ```
- **Access**: http://localhost:10020
- **Note**: Set OUTLINE_SECRET_KEY and OUTLINE_UTILS_SECRET in .env (32+ chars)

### Docusaurus (Static doc generator)
- **File**: `docusaurus.yml`
- **Port**: 10030
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  docker compose -f compose/docs/docusaurus.yml up -d
  ```
- **Access**: http://localhost:10030

## Port Range
10000-10099
