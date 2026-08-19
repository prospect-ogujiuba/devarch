#!/usr/bin/env bash
set -Eeuo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="$PROJECT_ROOT/apps"
ENV_FILE="$PROJECT_ROOT/.env"
PHP_COMPOSE="$PROJECT_ROOT/services-library/backend/php/compose.yml"
MARIADB_COMPOSE="$PROJECT_ROOT/services-library/database/mariadb/compose.yml"
REDIS_COMPOSE="$PROJECT_ROOT/services-library/database/redis/compose.yml"
MAILPIT_COMPOSE="$PROJECT_ROOT/services-library/mail/mailpit/compose.yml"
PROFILE_DIR="$SCRIPT_DIR/profiles"

readonly REDIS_HOST=redis
readonly REDIS_PASSWORD=devarch
readonly REDIS_PORT=6379
readonly REDIS_DB=0
readonly REDIS_CACHE_DB=1

APP_NAME=""
APP_TITLE=""
APP_URL=""
VERSION=""
PROFILE=bare
PACKAGES_FILE=""
DATABASE=mariadb
WITH_MAILPIT=false
WITH_REDIS=false
MIGRATE=true
MIGRATE_EXPLICIT=false
NO_MIGRATE_EXPLICIT=false
SEED=false
FORCE=false
DRY_RUN=false
RUNTIME=""
CONTAINER_USER=""
DB_ROOT_PASSWORD=""
COMPOSE=()
TARGET=""
CONTAINER_APP=""
DB_NAME=""
DB_USER=""
DB_PASSWORD=""
BACKUP_PATH=""
FAILED_PATH=""
RECOVERY_MARKER=""
BACKUP_MOVED=false
DB_CREATED=false
DB_USER_CREATED=false
PROVISIONING_STARTED=false
SUCCESS=false
CLI_PACKAGE_KINDS=()
CLI_PACKAGE_SPECS=()
PACKAGE_KINDS=()
PACKAGE_SPECS=()

log() { printf '[laravel] %s\n' "$*"; }
die() { printf '[laravel] error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage: scripts/laravel/bootstrap.sh <app-name> [options]

Create a fresh Laravel application at apps/<app-name> using DevArch's shared
PHP container and wildcard https://<app-name>.test proxy.

Options:
  --version CONSTRAINT       Composer create-project version constraint
  --profile NAME             Load a profile (default: bare)
  --list-profiles            List available profiles and exit
  --package PACKAGE          Add a production Composer package; repeatable
  --dev-package PACKAGE      Add a development Composer package; repeatable
  --packages-file FILE       Read composer-* directives from FILE
  --database mariadb|sqlite Database backend (default: mariadb)
  --with-mailpit             Configure the shared Mailpit SMTP service
  --with-redis               Configure shared Redis for cache and queue
  --migrate                  Run migrations (default)
  --no-migrate               Do not run migrations
  --seed                     Run db:seed once
  --force                    Preserve and replace an existing target
  --dry-run                  Validate and print a redacted, mutation-free plan
  --help                     Show this help (must be the sole argument)

PACKAGE is vendor/package or vendor/package:constraint. Profiles and package
files are declarative; use --list-profiles to inspect available profiles.
EOF
}

usage_error() {
  printf '[laravel] error: %s\n' "$1" >&2
  usage >&2
  exit 2
}

parse_informational_mode() {
  local arg saw_info=false
  for arg in "$@"; do
    [[ "$arg" == --help || "$arg" == --list-profiles ]] && saw_info=true
  done
  [[ "$saw_info" == true ]] || return 0
  if [[ $# -ne 1 ]]; then
    usage_error '--help and --list-profiles are standalone modes'
  fi
  case "$1" in
    --help) usage; exit 0 ;;
    --list-profiles) list_profiles; exit 0 ;;
  esac
}

require_value() {
  [[ $# -ge 2 && -n "$2" ]] || usage_error "$1 requires one nonempty value"
}

parse_args() {
  parse_informational_mode "$@"
  [[ $# -gt 0 ]] || usage_error 'app-name is required'

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --version)
        require_value "$@"
        VERSION="$2"
        shift 2
        ;;
      --profile)
        require_value "$@"
        PROFILE="$2"
        shift 2
        ;;
      --package)
        require_value "$@"
        parse_cli_package prod "$2"
        shift 2
        ;;
      --dev-package)
        require_value "$@"
        parse_cli_package dev "$2"
        shift 2
        ;;
      --packages-file)
        require_value "$@"
        PACKAGES_FILE="$2"
        shift 2
        ;;
      --database)
        require_value "$@"
        DATABASE="$2"
        shift 2
        ;;
      --with-mailpit) WITH_MAILPIT=true; shift ;;
      --with-redis) WITH_REDIS=true; shift ;;
      --migrate) MIGRATE=true; MIGRATE_EXPLICIT=true; shift ;;
      --no-migrate) MIGRATE=false; NO_MIGRATE_EXPLICIT=true; shift ;;
      --seed) SEED=true; shift ;;
      --force) FORCE=true; shift ;;
      --dry-run) DRY_RUN=true; shift ;;
      --repository|--branch|--database-import|--npm-install|--build|--npm-build)
        usage_error "unsupported v1 option: $1"
        ;;
      --*) usage_error "unknown option: $1" ;;
      *)
        [[ -z "$APP_NAME" ]] || usage_error "unexpected extra positional argument: $1"
        APP_NAME="$1"
        shift
        ;;
    esac
  done
}

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

validate_no_controls() {
  local label="$1" value="$2" clean
  clean="$(printf '%s' "$value" | LC_ALL=C tr -d '\001-\037\177')"
  [[ "$value" == "$clean" ]] || die "$label contains an ASCII control byte"
}

parse_repository_env_value() {
  local key="$1" raw quote char remainder value="" escaped=false closed=false i
  raw="$(trim "$2")"
  if [[ "$raw" == \"* || "$raw" == \'* ]]; then
    quote="${raw:0:1}"
    for ((i = 1; i < ${#raw}; i++)); do
      char="${raw:i:1}"
      if [[ "$escaped" == true ]]; then
        if [[ "$char" == "\\" || "$char" == "$quote" || ( "$quote" == '"' && "$char" == '$' ) ]]; then
          value+="$char"
        else
          die "$key contains an unsupported quoted escape: \\$char"
        fi
        escaped=false
      elif [[ "$char" == "\\" ]]; then
        escaped=true
      elif [[ "$char" == "$quote" ]]; then
        closed=true
        ((i += 1))
        break
      else
        value+="$char"
      fi
    done
    [[ "$closed" == true ]] || die "$key has an unterminated quoted value"
    remainder="$(trim "${raw:i}")"
    [[ -z "$remainder" || "$remainder" == \#* ]] || die "$key has unexpected text after its quoted value"
  else
    for ((i = 1; i < ${#raw}; i++)); do
      if [[ "${raw:i:1}" == '#' && "${raw:i-1:1}" == [[:space:]] ]]; then
        raw="${raw:0:i}"
        break
      fi
    done
    value="$(trim "$raw")"
  fi
  validate_no_controls "$key" "$value"
  PARSED_ENV_VALUE="$value"
}

load_repository_env() {
  [[ -f "$ENV_FILE" ]] || return 0
  if od -An -tx1 "$ENV_FILE" | grep -Eq '(^|[[:space:]])00([[:space:]]|$)'; then
    die "$ENV_FILE contains NUL"
  fi

  local line original_line key
  while IFS= read -r line || [[ -n "$line" ]]; do
    original_line="$line"
    line="$(trim "$line")"
    [[ -z "$line" || "$line" == \#* ]] && continue
    [[ "$line" == export\ * ]] && line="$(trim "${line#export }")"
    [[ "$line" == *=* ]] || continue
    key="$(trim "${line%%=*}")"
    case "$key" in
      LARAVEL_APP_NAME|LARAVEL_APP_URL|LARAVEL_DB_ROOT_PASSWORD|MARIADB_ROOT_PASSWORD|LARAVEL_CONTAINER_USER|CONTAINER_RUNTIME)
        validate_no_controls "$key" "$original_line"
        parse_repository_env_value "$key" "${line#*=}"
        printf -v "$key" '%s' "$PARSED_ENV_VALUE"
        ;;
    esac
  done < "$ENV_FILE"
}

title_case() {
  printf '%s' "$1" | tr '-' ' ' | awk '{ for (i=1; i<=NF; i++) $i=toupper(substr($i,1,1)) substr($i,2); print }'
}

validate_url() {
  local value="$1" rest authority host port=""
  validate_no_controls 'LARAVEL_APP_URL' "$value"
  [[ "$value" != *[[:space:]]* ]] || die 'LARAVEL_APP_URL must not contain whitespace'
  [[ "$value" != *'#'* ]] || die 'LARAVEL_APP_URL must not contain a fragment'
  [[ "$value" =~ ^https?:// ]] || die 'LARAVEL_APP_URL scheme must be http or https'
  rest="${value#*://}"
  authority="${rest%%[/?]*}"
  [[ -n "$authority" ]] || die 'LARAVEL_APP_URL requires a hostname'
  [[ "$authority" != *'@'* ]] || die 'LARAVEL_APP_URL must not contain userinfo'

  if [[ "$authority" == \[* ]]; then
    [[ "$authority" =~ ^\[([^]]+)\](:([0-9]+))?$ ]] || die 'LARAVEL_APP_URL has an invalid host or port'
    host="${BASH_REMATCH[1]}"
    port="${BASH_REMATCH[3]:-}"
  else
    [[ "$authority" =~ ^([^:]+)(:([0-9]+))?$ ]] || die 'LARAVEL_APP_URL has an invalid host or port'
    host="${BASH_REMATCH[1]}"
    port="${BASH_REMATCH[3]:-}"
  fi
  [[ -n "$host" ]] || die 'LARAVEL_APP_URL requires a hostname'
  if [[ -n "$port" ]]; then
    ((${#port} <= 5)) || die 'LARAVEL_APP_URL port must be between 1 and 65535'
    ((10#$port >= 1 && 10#$port <= 65535)) || die 'LARAVEL_APP_URL port must be between 1 and 65535'
  fi
}

sha256_12() {
  local value="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s' "$value" | sha256sum | awk '{print substr($1,1,12)}'
  elif command -v shasum >/dev/null 2>&1; then
    printf '%s' "$value" | shasum -a 256 | awk '{print substr($1,1,12)}'
  else
    die 'sha256sum or shasum is required'
  fi
}

derive_identifiers() {
  local normalized hash full_db full_user
  normalized="${APP_NAME//-/_}"
  hash="$(sha256_12 "$APP_NAME")"
  full_db="laravel_$normalized"
  full_user="lv_$normalized"
  if ((${#full_db} <= 64)); then DB_NAME="$full_db"; else DB_NAME="laravel_${normalized:0:43}_$hash"; fi
  if ((${#full_user} <= 32)); then DB_USER="$full_user"; else DB_USER="lv_${normalized:0:16}_$hash"; fi
  [[ "$DB_NAME" =~ ^[a-z0-9_]+$ && ${#DB_NAME} -le 64 ]] || die 'internal database identifier derivation failed'
  [[ "$DB_USER" =~ ^[a-z0-9_]+$ && ${#DB_USER} -le 32 ]] || die 'internal database user derivation failed'
}

choose_unique_path() {
  local base="$1" candidate="$1" suffix=0
  while [[ -e "$candidate" || -L "$candidate" ]]; do
    ((suffix += 1))
    candidate="$base-$suffix"
  done
  printf '%s' "$candidate"
}

is_composer_version_atom() {
  local atom="$1"
  [[ "$atom" =~ ^(dev-[A-Za-z0-9][A-Za-z0-9._/-]*|\*|v?[0-9]+([.][0-9]+){0,3}(-[A-Za-z0-9][A-Za-z0-9.-]*)?(\+[A-Za-z0-9][A-Za-z0-9.-]*)?|v?[0-9]+([.][0-9]+){0,2}[.](x|X|\*)(-dev)?)(@[A-Za-z]+)?$ ]]
}

validate_composer_constraint() {
  local constraint normalized clause token atom after_comma=false seen_atom=false
  local -a tokens
  constraint="$(trim "$1")"
  [[ -n "$constraint" && "$constraint" != \|* && "$constraint" != *\| && "$constraint" != *'|||'* ]] || die 'invalid Composer version constraint'
  normalized="${constraint//||/$'\n'}"
  normalized="${normalized//|/$'\n'}"
  while IFS= read -r clause || [[ -n "$clause" ]]; do
    clause="$(trim "$clause")"
    [[ -n "$clause" ]] || die 'invalid Composer version constraint'
    clause="${clause//,/ , }"
    read -r -a tokens <<< "$clause"
    if [[ ${#tokens[@]} -eq 3 && "${tokens[1]}" == '-' ]]; then
      is_composer_version_atom "${tokens[0]}" && is_composer_version_atom "${tokens[2]}" || die 'invalid Composer version constraint'
      continue
    fi
    after_comma=false
    seen_atom=false
    for token in "${tokens[@]}"; do
      if [[ "$token" == ',' ]]; then
        [[ "$seen_atom" == true && "$after_comma" == false ]] || die 'invalid Composer version constraint'
        after_comma=true
        continue
      fi
      case "$token" in
        '=='*|'!='*|'>='*|'<='*) atom="${token:2}" ;;
        '='*|'>'*|'<'*|'^'*|'~'*) atom="${token:1}" ;;
        *) atom="$token" ;;
      esac
      [[ -n "$atom" ]] && is_composer_version_atom "$atom" || die 'invalid Composer version constraint'
      seen_atom=true
      after_comma=false
    done
    [[ "$seen_atom" == true && "$after_comma" == false ]] || die 'invalid Composer version constraint'
  done <<< "$normalized"
}

list_profiles() {
  local file name description
  for file in "$PROFILE_DIR"/*.profile; do
    [[ -f "$file" ]] || continue
    name="$(basename "$file" .profile)"
    description="$(grep -m1 '^# ' "$file" 2>/dev/null || true)"
    printf '%-10s %s\n' "$name" "${description#\# }"
  done
}

validate_composer_package_name() {
  local package="$1"
  validate_no_controls 'Composer package name' "$package"
  [[ "$package" =~ ^[a-z0-9]+([_.-][a-z0-9]+)*/[a-z0-9]+([_.-][a-z0-9]+)*$ ]] || \
    die "invalid Composer package name: $package"
}

add_package() {
  local kind="$1" package="$2" constraint="${3:-}" spec
  validate_composer_package_name "$package"
  if [[ -n "$constraint" ]]; then
    validate_no_controls 'Composer package constraint' "$constraint"
    validate_composer_constraint "$constraint"
    spec="$package:$constraint"
  else
    spec="$package"
  fi
  PACKAGE_KINDS+=("$kind")
  PACKAGE_SPECS+=("$spec")
}

parse_cli_package() {
  local kind="$1" value="$2" package constraint=""
  validate_no_controls 'Composer package request' "$value"
  package="${value%%:*}"
  [[ "$value" != *:* ]] || constraint="${value#*:}"
  [[ -n "$package" && ( "$value" != *:* || -n "$constraint" ) ]] || die "invalid Composer package request: $value"
  validate_composer_package_name "$package"
  [[ -z "$constraint" ]] || validate_composer_constraint "$constraint"
  CLI_PACKAGE_KINDS+=("$kind")
  CLI_PACKAGE_SPECS+=("$value")
}

parse_directives_file() {
  local file="$1" allow_features="$2" line directive fields package constraint line_number=0 clean
  [[ -f "$file" ]] || die "declarative file not found: $file"
  command -v od >/dev/null 2>&1 || die 'od is required for declarative file validation'
  if od -An -tx1 "$file" | grep -Eq '(^|[[:space:]])00([[:space:]]|$)'; then
    die "$file contains NUL"
  fi
  while IFS= read -r line || [[ -n "$line" ]]; do
    ((line_number += 1))
    clean="$(printf '%s' "$line" | LC_ALL=C tr -d '\001-\010\013-\037\177')"
    [[ "$line" == "$clean" ]] || die "invalid control byte in $file:$line_number"
    line="${line%%#*}"
    line="$(trim "$line")"
    [[ -n "$line" ]] || continue
    directive=""; fields=""; package=""; constraint=""
    read -r directive fields <<< "$line"
    read -r package constraint <<< "$fields"
    case "$directive" in
      composer-package|composer-dev-package)
        [[ -n "$package" ]] || die "invalid directive fields in $file:$line_number"
        add_package "$([[ "$directive" == composer-package ]] && printf prod || printf dev)" "$package" "$constraint"
        ;;
      feature)
        [[ "$allow_features" == true ]] || die "unknown package-file directive 'feature' in $file:$line_number"
        [[ -n "$package" && -z "$constraint" ]] || die "invalid feature fields in $file:$line_number"
        case "$package" in
          mailpit) WITH_MAILPIT=true ;;
          redis) WITH_REDIS=true ;;
          *) die "unknown profile feature '$package' in $file:$line_number" ;;
        esac
        ;;
      *) die "unknown directive '$directive' in $file:$line_number" ;;
    esac
  done < "$file"
}

load_profile() {
  [[ "$PROFILE" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || die "invalid profile name: $PROFILE"
  local file="$PROFILE_DIR/$PROFILE.profile"
  [[ -f "$file" ]] || die "unknown profile: $PROFILE (use --list-profiles)"
  parse_directives_file "$file" true
}

load_packages_file() {
  [[ -n "$PACKAGES_FILE" ]] || return 0
  validate_no_controls '--packages-file' "$PACKAGES_FILE"
  parse_directives_file "$PACKAGES_FILE" false
}

append_cli_packages() {
  local i
  for ((i = 0; i < ${#CLI_PACKAGE_SPECS[@]}; i++)); do
    PACKAGE_KINDS+=("${CLI_PACKAGE_KINDS[i]}")
    PACKAGE_SPECS+=("${CLI_PACKAGE_SPECS[i]}")
  done
}

validate_config() {
  [[ -n "$APP_NAME" ]] || usage_error 'app-name is required'
  validate_no_controls 'app-name' "$APP_NAME"
  [[ ${#APP_NAME} -le 63 && "$APP_NAME" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]] || \
    die 'app-name must be 1-63 ASCII characters matching ^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$'
  [[ "$DATABASE" == mariadb || "$DATABASE" == sqlite ]] || die '--database must be mariadb or sqlite'
  [[ "$MIGRATE_EXPLICIT" != true || "$NO_MIGRATE_EXPLICIT" != true ]] || die '--migrate and --no-migrate are mutually exclusive'

  if [[ -n "$VERSION" ]]; then
    validate_no_controls '--version' "$VERSION"
    validate_composer_constraint "$VERSION"
  fi

  load_profile
  load_packages_file
  append_cli_packages

  APP_TITLE="${LARAVEL_APP_NAME:-$(title_case "$APP_NAME")}"
  [[ -n "$APP_TITLE" ]] || die 'LARAVEL_APP_NAME must be nonempty'
  validate_no_controls 'LARAVEL_APP_NAME' "$APP_TITLE"
  APP_URL="${LARAVEL_APP_URL:-https://$APP_NAME.test}"
  validate_url "$APP_URL"

  RUNTIME="${CONTAINER_RUNTIME:-}"
  validate_no_controls 'CONTAINER_RUNTIME' "$RUNTIME"
  [[ -z "$RUNTIME" || "$RUNTIME" == podman || "$RUNTIME" == docker ]] || die 'CONTAINER_RUNTIME must be podman or docker'
  CONTAINER_USER="${LARAVEL_CONTAINER_USER:-}"
  validate_no_controls 'LARAVEL_CONTAINER_USER' "$CONTAINER_USER"
  [[ -z "$CONTAINER_USER" || "$CONTAINER_USER" =~ ^[0-9]+:[0-9]+$ ]] || die 'LARAVEL_CONTAINER_USER must match numeric uid:gid'
  DB_ROOT_PASSWORD="${LARAVEL_DB_ROOT_PASSWORD:-${MARIADB_ROOT_PASSWORD:-devarch}}"
  validate_no_controls 'database root password' "$DB_ROOT_PASSWORD"

  TARGET="$APPS_DIR/$APP_NAME"
  CONTAINER_APP="/var/www/html/$APP_NAME"
  RECOVERY_MARKER="$APPS_DIR/.devarch-recovery/$APP_NAME"
  [[ ! -e "$RECOVERY_MARKER" && ! -L "$RECOVERY_MARKER" ]] || die "unresolved recovery marker exists: $RECOVERY_MARKER"
  if [[ ( -e "$TARGET" || -L "$TARGET" ) && "$FORCE" != true ]]; then
    die "$TARGET already exists; use --force to preserve and replace it"
  fi

  local stamp
  stamp="$(date +%Y%m%d-%H%M%S)"
  BACKUP_PATH="$(choose_unique_path "$APPS_DIR/.devarch-backups/$APP_NAME-$stamp")"
  FAILED_PATH="$(choose_unique_path "$APPS_DIR/.devarch-failed/$APP_NAME-$stamp")"
  [[ "$DATABASE" == mariadb ]] && derive_identifiers

  local file
  for file in "$PHP_COMPOSE"; do
    [[ -f "$file" ]] || die "required Compose file not found: $file"
  done
  [[ "$DATABASE" != mariadb || -f "$MARIADB_COMPOSE" ]] || die "required Compose file not found: $MARIADB_COMPOSE"
  [[ "$WITH_REDIS" != true || -f "$REDIS_COMPOSE" ]] || die "required Compose file not found: $REDIS_COMPOSE"
  [[ "$WITH_MAILPIT" != true || -f "$MAILPIT_COMPOSE" ]] || die "required Compose file not found: $MAILPIT_COMPOSE"
  command -v awk >/dev/null 2>&1 || die 'awk is required'
  command -v tr >/dev/null 2>&1 || die 'tr is required'
  if [[ "$DRY_RUN" != true ]]; then
    command -v openssl >/dev/null 2>&1 || command -v od >/dev/null 2>&1 || die 'openssl or od is required for secure password generation'
  fi
}

detect_runtime() {
  if [[ -z "$RUNTIME" ]]; then
    if command -v podman >/dev/null 2>&1; then RUNTIME=podman
    elif command -v docker >/dev/null 2>&1; then RUNTIME=docker
    else die 'Podman or Docker is required'
    fi
  fi
  command -v "$RUNTIME" >/dev/null 2>&1 || die "container runtime not found: $RUNTIME"
  if "$RUNTIME" compose version >/dev/null 2>&1; then
    COMPOSE=("$RUNTIME" compose)
  elif [[ "$RUNTIME" == podman ]] && command -v podman-compose >/dev/null 2>&1; then
    COMPOSE=(podman-compose)
  else
    die "no Compose provider found for $RUNTIME"
  fi
  if [[ -z "$CONTAINER_USER" ]]; then
    if [[ "$RUNTIME" == podman ]]; then CONTAINER_USER='0:0'
    else CONTAINER_USER="$(id -u):$(id -g)"
    fi
  fi
}

print_plan() {
  log 'dry run; no changes will be made'
  log "target: $TARGET"
  if [[ -e "$TARGET" || -L "$TARGET" ]]; then log "backup existing target: $BACKUP_PATH"; else log 'target is new'; fi
  log "runtime: $RUNTIME; container user: $CONTAINER_USER"
  log "profile: $PROFILE"
  log 'ensure external network: microservices-net'
  log 'start and wait: php'
  [[ "$DATABASE" == mariadb ]] && log "start and wait: mariadb; create isolated database/user: $DB_NAME / $DB_USER"
  [[ "$WITH_MAILPIT" == true ]] && log 'start and wait: mailpit'
  [[ "$WITH_REDIS" == true ]] && log 'start and wait: redis (password redacted)'
  if [[ -n "$VERSION" ]]; then log "scaffold Laravel through Composer (constraint: $VERSION)"; else log 'scaffold Laravel through Composer (current stable)'; fi
  local i label
  for ((i = 0; i < ${#PACKAGE_SPECS[@]}; i++)); do
    [[ "${PACKAGE_KINDS[i]}" == prod ]] && label=production || label=development
    log "Composer $label package: ${PACKAGE_SPECS[i]}"
  done
  log "configure APP_URL=$APP_URL and database=$DATABASE; generate APP_KEY"
  [[ "$WITH_MAILPIT" == true ]] && log 'configure Mailpit SMTP at mailpit:1025'
  [[ "$WITH_REDIS" == true ]] && log 'configure phpredis at redis:6379 (credentials redacted), cache DB 1, queue DB 0'
  [[ "$MIGRATE" != true ]] || log 'run migrations'
  [[ "$SEED" != true ]] || log 'run database seeder once'
  return 0
}

ensure_network() {
  if ! "$RUNTIME" network inspect microservices-net >/dev/null 2>&1; then
    "$RUNTIME" network create microservices-net >/dev/null
  fi
}

compose_up() {
  local file="$1"
  shift
  "${COMPOSE[@]}" -f "$file" up -d "$@"
}

start_services() {
  log 'start shared PHP service'
  compose_up "$PHP_COMPOSE"
  if [[ "$DATABASE" == mariadb ]]; then
    log 'start shared MariaDB service'
    env MARIADB_ROOT_PASSWORD="$DB_ROOT_PASSWORD" "${COMPOSE[@]}" -f "$MARIADB_COMPOSE" up -d
  fi
  if [[ "$WITH_MAILPIT" == true ]]; then log 'start shared Mailpit service'; compose_up "$MAILPIT_COMPOSE"; fi
  if [[ "$WITH_REDIS" == true ]]; then log 'start shared Redis service'; compose_up "$REDIS_COMPOSE"; fi
}

wait_for_services() {
  local attempt ready
  log 'wait for required services'
  for attempt in {1..90}; do
    ready=true
    "$RUNTIME" exec php php -v >/dev/null 2>&1 || ready=false
    "$RUNTIME" exec php composer --version >/dev/null 2>&1 || ready=false
    if [[ "$DATABASE" == mariadb ]]; then
      "$RUNTIME" exec mariadb sh -c 'mariadb-admin ping -uroot -p"$MARIADB_ROOT_PASSWORD" --silent' >/dev/null 2>&1 || ready=false
    fi
    if [[ "$WITH_REDIS" == true ]]; then
      "$RUNTIME" exec redis redis-cli -a "$REDIS_PASSWORD" ping >/dev/null 2>&1 || ready=false
    fi
    if [[ "$WITH_MAILPIT" == true ]]; then
      "$RUNTIME" exec mailpit wget --spider -q http://localhost:8025/ >/dev/null 2>&1 || ready=false
    fi
    [[ "$ready" == true ]] && return 0
    sleep 1
  done
  die 'required services did not become ready within 90 seconds'
}

db_exec() {
  local sql="$1"
  if ! printf '%s\n' "$sql" | "$RUNTIME" exec -i mariadb sh -c 'mariadb -uroot -p"$MARIADB_ROOT_PASSWORD"' >/dev/null; then
    die 'MariaDB operation failed (SQL and credentials suppressed)'
  fi
}

db_query() {
  local sql="$1"
  printf '%s\n' "$sql" | "$RUNTIME" exec -i mariadb sh -c 'mariadb -N -B -uroot -p"$MARIADB_ROOT_PASSWORD"' 2>/dev/null
}

assert_database_available() {
  [[ "$DATABASE" == mariadb ]] || return 0
  local found
  found="$(db_query "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME';")" || die 'could not check derived database identifier'
  [[ -z "$found" ]] || die "derived database identifier already exists; refusing: $DB_NAME"
  found="$(db_query "SELECT User FROM mysql.user WHERE User='$DB_USER' LIMIT 1;")" || die 'could not check derived database user identifier'
  [[ -z "$found" ]] || die "derived database user identifier already exists; refusing: $DB_USER"
}

generate_db_password() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 24
  else
    od -An -N24 -tx1 /dev/urandom | tr -d ' \n'
  fi
}

create_database() {
  [[ "$DATABASE" == mariadb ]] || return 0
  DB_PASSWORD="$(generate_db_password)"
  db_exec "CREATE DATABASE \`$DB_NAME\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
  DB_CREATED=true
  db_exec "CREATE USER '$DB_USER'@'%' IDENTIFIED BY '$DB_PASSWORD';"
  DB_USER_CREATED=true
  db_exec "GRANT ALL PRIVILEGES ON \`$DB_NAME\`.* TO '$DB_USER'@'%'; FLUSH PRIVILEGES;"
  log "created isolated MariaDB database/user: $DB_NAME / $DB_USER"
  if [[ "${LARAVEL_TEST_FAIL_AFTER_DATABASE:-0}" == 1 ]]; then
    die 'injected failure after database/user creation'
  fi
}

encode_dotenv() {
  local value="$1"
  validate_no_controls 'dotenv value' "$value"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//\$/\\\$}"
  DOTENV_ENCODED="\"$value\""
}

dotenv_set() {
  local key="$1" value="$2" file="$3" line found=false tmp
  tmp="$file.devarch-tmp"
  [[ "$key" =~ ^[A-Z][A-Z0-9_]*$ ]] || die "invalid internal dotenv key: $key"
  encode_dotenv "$value"
  : > "$tmp"
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" == "$key="* ]]; then
      if [[ "$found" == false ]]; then printf '%s=%s\n' "$key" "$DOTENV_ENCODED" >> "$tmp"; found=true; fi
    else
      printf '%s\n' "$line" >> "$tmp"
    fi
  done < "$file"
  [[ "$found" == true ]] || printf '%s=%s\n' "$key" "$DOTENV_ENCODED" >> "$tmp"
  mv "$tmp" "$file"
}

artisan() {
  "$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp php php "$CONTAINER_APP/artisan" "$@"
}

scaffold_application() {
  local command=("$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp php composer create-project --no-interaction laravel/laravel "$CONTAINER_APP")
  [[ -n "$VERSION" ]] && command+=("$VERSION")
  log 'scaffold Laravel through the shared PHP container'
  "${command[@]}"
  [[ -f "$TARGET/.env" ]] || die 'Composer scaffold did not create .env'
}

install_packages() {
  local i kind
  local -a command
  for ((i = 0; i < ${#PACKAGE_SPECS[@]}; i++)); do
    kind="${PACKAGE_KINDS[i]}"
    command=("$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp --workdir "$CONTAINER_APP" php composer require --no-interaction --no-progress)
    [[ "$kind" == prod ]] || command+=(--dev)
    command+=("${PACKAGE_SPECS[i]}")
    log "install Composer $([[ "$kind" == prod ]] && printf production || printf development) package: ${PACKAGE_SPECS[i]}"
    "${command[@]}"
  done
}

make_runtime_writable() {
  [[ -d "$TARGET/storage" && -d "$TARGET/bootstrap/cache" ]] || die 'Laravel runtime directories are missing'
  chmod -R a+rwX "$TARGET/storage" "$TARGET/bootstrap/cache"
  if [[ "$DATABASE" == sqlite ]]; then
    [[ -d "$TARGET/database" ]] || die 'Laravel SQLite directory is missing'
    chmod -R a+rwX "$TARGET/database"
  fi
}

configure_application() {
  local env_file="$TARGET/.env"
  dotenv_set APP_NAME "$APP_TITLE" "$env_file"
  dotenv_set APP_ENV local "$env_file"
  dotenv_set APP_DEBUG true "$env_file"
  dotenv_set APP_URL "$APP_URL" "$env_file"

  if [[ "$DATABASE" == mariadb ]]; then
    dotenv_set DB_CONNECTION mysql "$env_file"
    dotenv_set DB_HOST mariadb "$env_file"
    dotenv_set DB_PORT 3306 "$env_file"
    dotenv_set DB_DATABASE "$DB_NAME" "$env_file"
    dotenv_set DB_USERNAME "$DB_USER" "$env_file"
    dotenv_set DB_PASSWORD "$DB_PASSWORD" "$env_file"
  else
    mkdir -p "$TARGET/database"
    : > "$TARGET/database/database.sqlite"
    dotenv_set DB_CONNECTION sqlite "$env_file"
    dotenv_set DB_DATABASE "$CONTAINER_APP/database/database.sqlite" "$env_file"
  fi

  dotenv_set MAIL_MAILER "$([[ "$WITH_MAILPIT" == true ]] && printf smtp || printf log)" "$env_file"
  if [[ "$WITH_MAILPIT" == true ]]; then
    dotenv_set MAIL_HOST mailpit "$env_file"
    dotenv_set MAIL_PORT 1025 "$env_file"
    dotenv_set MAIL_USERNAME '' "$env_file"
    dotenv_set MAIL_PASSWORD '' "$env_file"
    dotenv_set MAIL_ENCRYPTION '' "$env_file"
  fi

  dotenv_set CACHE_STORE "$([[ "$WITH_REDIS" == true ]] && printf redis || printf file)" "$env_file"
  dotenv_set QUEUE_CONNECTION "$([[ "$WITH_REDIS" == true ]] && printf redis || printf sync)" "$env_file"
  if [[ "$WITH_REDIS" == true ]]; then
    dotenv_set REDIS_CLIENT phpredis "$env_file"
    dotenv_set REDIS_HOST "$REDIS_HOST" "$env_file"
    dotenv_set REDIS_PASSWORD "$REDIS_PASSWORD" "$env_file"
    dotenv_set REDIS_PORT "$REDIS_PORT" "$env_file"
    dotenv_set REDIS_DB "$REDIS_DB" "$env_file"
    dotenv_set REDIS_CACHE_DB "$REDIS_CACHE_DB" "$env_file"
  fi

  log 'generate Laravel application key'
  artisan key:generate --force --no-interaction >/dev/null
  if [[ "$MIGRATE" == true ]]; then log 'run Laravel migrations'; artisan migrate --force --no-interaction; fi
  if [[ "$SEED" == true ]]; then log 'run Laravel database seeder'; artisan db:seed --force --no-interaction; fi
}

create_recovery_guard() {
  [[ "${LARAVEL_TEST_FAIL_RECOVERY_MARKER_CREATE:-0}" != 1 ]] || die "could not create durable recovery marker: $RECOVERY_MARKER"
  mkdir -p "$(dirname "$RECOVERY_MARKER")" || die "could not create recovery marker directory: $(dirname "$RECOVERY_MARKER")"
  (set -o noclobber; printf 'provisioning in progress\n' > "$RECOVERY_MARKER") 2>/dev/null || \
    die "could not create durable recovery marker: $RECOVERY_MARKER"
  [[ "$(<"$RECOVERY_MARKER")" == 'provisioning in progress' ]] || die "could not verify durable recovery marker: $RECOVERY_MARKER"
}

persist_recovery_state() {
  local tmp="$RECOVERY_MARKER.tmp.$$" message="$1"
  if [[ "${LARAVEL_TEST_FAIL_RECOVERY_MARKER_WRITE:-0}" == 1 ]] || \
     ! printf '%s\n' "$message" > "$tmp" || ! mv "$tmp" "$RECOVERY_MARKER" || \
     [[ "$(<"$RECOVERY_MARKER")" != "$message" ]]; then
    rm -f "$tmp"
    printf '[laravel] recovery marker update failed; original durable guard remains: %s\n' "$RECOVERY_MARKER" >&2
    return 1
  fi
}

clear_recovery_guard() {
  rm -f "$RECOVERY_MARKER" && [[ ! -e "$RECOVERY_MARKER" && ! -L "$RECOVERY_MARKER" ]]
}

move_existing_target() {
  [[ -e "$TARGET" || -L "$TARGET" ]] || return 0
  [[ ! -e "$BACKUP_PATH" && ! -L "$BACKUP_PATH" ]] || die "backup path was claimed concurrently: $BACKUP_PATH"
  mkdir -p "$(dirname "$BACKUP_PATH")"
  log "preserve existing target: $BACKUP_PATH"
  mv "$TARGET" "$BACKUP_PATH"
  BACKUP_MOVED=true
  PROVISIONING_STARTED=true
  if [[ "${LARAVEL_TEST_FAIL_AFTER_BACKUP:-0}" == 1 ]]; then
    die 'injected failure after target backup'
  fi
}

rollback() {
  local original_status="$1" quarantine_status=not-present db_status=not-created user_status=not-created restore_status=not-needed incomplete=false
  trap - EXIT ERR
  set +e
  printf '[laravel] provisioning failed; starting recovery\n' >&2

  if [[ -e "$TARGET" || -L "$TARGET" ]]; then
    FAILED_PATH="$(choose_unique_path "$FAILED_PATH")"
    mkdir -p "$(dirname "$FAILED_PATH")"
    if [[ "${LARAVEL_TEST_FAIL_QUARANTINE:-0}" == 1 ]]; then
      quarantine_status=failed
      incomplete=true
    elif mv "$TARGET" "$FAILED_PATH"; then
      quarantine_status="quarantined:$FAILED_PATH"
    else
      quarantine_status=failed
      incomplete=true
    fi
  fi

  if [[ "$DB_CREATED" == true ]]; then
    if [[ "${LARAVEL_TEST_FAIL_CLEANUP_DATABASE:-0}" == 1 ]]; then
      db_status=failed; incomplete=true
    elif printf 'DROP DATABASE `%s`;\n' "$DB_NAME" | "$RUNTIME" exec -i mariadb sh -c 'mariadb -uroot -p"$MARIADB_ROOT_PASSWORD"' >/dev/null 2>&1; then
      db_status=dropped
    else
      db_status=failed; incomplete=true
    fi
  fi
  if [[ "$DB_USER_CREATED" == true ]]; then
    if [[ "${LARAVEL_TEST_FAIL_CLEANUP_USER:-0}" == 1 ]]; then
      user_status=failed; incomplete=true
    elif printf "DROP USER '%s'@'%%';\n" "$DB_USER" | "$RUNTIME" exec -i mariadb sh -c 'mariadb -uroot -p"$MARIADB_ROOT_PASSWORD"' >/dev/null 2>&1; then
      user_status=dropped
    else
      user_status=failed; incomplete=true
    fi
  fi

  if [[ "$BACKUP_MOVED" == true ]]; then
    if [[ "${LARAVEL_TEST_FAIL_RESTORE:-0}" == 1 ]]; then
      restore_status=failed; incomplete=true
    elif [[ -e "$TARGET" ]]; then
      restore_status='failed:target-occupied'; incomplete=true
    elif mv "$BACKUP_PATH" "$TARGET"; then
      restore_status=restored
    else
      restore_status=failed; incomplete=true
    fi
  fi

  if [[ "$incomplete" == true ]]; then
    persist_recovery_state 'recovery incomplete' || true
    printf '[laravel] recovery incomplete; unsafe retry refused until resolved\n' >&2
    printf '[laravel] recovery target=%s\n' "$TARGET" >&2
    printf '[laravel] recovery backup=%s\n' "$BACKUP_PATH" >&2
    printf '[laravel] recovery failed-target=%s status=%s\n' "$FAILED_PATH" "$quarantine_status" >&2
    printf '[laravel] recovery database=%s status=%s\n' "${DB_NAME:-not-applicable}" "$db_status" >&2
    printf '[laravel] recovery user=%s status=%s\n' "${DB_USER:-not-applicable}" "$user_status" >&2
    printf '[laravel] recovery restore=%s\n' "$restore_status" >&2
  else
    if clear_recovery_guard; then
      [[ "$BACKUP_MOVED" == true ]] && printf '[laravel] recovery complete; old target restored: %s\n' "$TARGET" >&2
      [[ "$quarantine_status" == quarantined:* ]] && printf '[laravel] partial target quarantined: %s\n' "$FAILED_PATH" >&2
    else
      printf '[laravel] recovery complete, but durable marker removal failed; unsafe retry remains refused: %s\n' "$RECOVERY_MARKER" >&2
    fi
  fi
  exit "$original_status"
}

on_exit() {
  local status="$1"
  if ((status != 0)) && [[ "$PROVISIONING_STARTED" == true && "$SUCCESS" != true ]]; then
    rollback "$status"
  fi
}
main() {
  parse_args "$@"
  load_repository_env
  validate_config
  detect_runtime

  if [[ "$DRY_RUN" == true ]]; then
    print_plan
    SUCCESS=true
    return 0
  fi

  create_recovery_guard
  PROVISIONING_STARTED=true
  ensure_network
  start_services
  wait_for_services
  assert_database_available
  mkdir -p "$APPS_DIR"
  move_existing_target
  PROVISIONING_STARTED=true
  create_database
  scaffold_application
  install_packages
  make_runtime_writable
  configure_application
  clear_recovery_guard || die "provisioning succeeded but durable recovery marker could not be removed: $RECOVERY_MARKER"
  SUCCESS=true
  log "ready through the wildcard proxy: $APP_URL"
  [[ "$BACKUP_MOVED" != true ]] || log "previous target preserved at: $BACKUP_PATH"
  return 0
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  trap 'on_exit $?' EXIT
  main "$@"
fi
