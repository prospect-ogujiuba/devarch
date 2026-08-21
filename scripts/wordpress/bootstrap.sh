#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="${DEVARCH_APPS_DIR:-$PROJECT_ROOT/apps}"
PHP_COMPOSE="$PROJECT_ROOT/services-library/backend/php/compose.yml"
MARIADB_COMPOSE="$PROJECT_ROOT/services-library/database/mariadb/compose.yml"
PROXY_COMPOSE="$PROJECT_ROOT/services-library/proxy/nginx-proxy-manager/compose.yml"
ENV_FILE="$PROJECT_ROOT/.env"
PROFILE_DIR="$SCRIPT_DIR/profiles"
HOSTS_HELPER="$PROJECT_ROOT/scripts/hosts/register-host.sh"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

DRY_RUN=false
FORCE=false
BUILD=false
REGISTER_HOSTS=true
SITE_NAME=""
SITE_TITLE=""
SITE_URL=""
PROFILE=""
PLUGINS_FILE=""
RESTORE_FILE=""
REPLACE_EXISTING=false
SCAFFOLDS_ONLY=false
PROJECT_SLUG=""
PHP_NAMESPACE=""
JS_NAMESPACE=""
PLUGIN_SOURCES=()
PLUGIN_ACTIVATIONS=()
THEME_SOURCES=()
MU_PLUGIN_SOURCES=()
MAKER_CORE_SLUGS=()
MAKER_CORE_TYPES=()
MAKER_CORE_URLS=()
MAKER_WORKSPACES=()
MAKER_STACK_CHANNEL=""
RUNTIME="${CONTAINER_RUNTIME:-}"
CONTAINER_USER="${WORDPRESS_CONTAINER_USER:-}"
PHP_CONTAINER="php"
MARIADB_CONTAINER="mariadb"
ADMIN_USER_VALUE="${WP_ADMIN_USER:-${ADMIN_USER:-admin}}"
ADMIN_PASSWORD_VALUE="${WP_ADMIN_PASSWORD:-${ADMIN_PASSWORD:-}}"
ADMIN_EMAIL_VALUE="${WP_ADMIN_EMAIL:-${ADMIN_EMAIL:-admin@devarch.test}}"
DB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-devarch}"
AIOWM_GIT_URL="${AIOWM_GIT_URL:-${GITHUB_USER:+git@github.com:${GITHUB_USER}/all-in-one-wp-migration.git}}"
COMPOSE=()

log() {
  printf '[wordpress] %s\n' "$*"
}

die() {
  printf '[wordpress] error: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage: scripts/wordpress/bootstrap.sh [site-name] [options]

Create a local WordPress site in apps/<site-name> using the existing PHP,
MariaDB, and wildcard .test reverse-proxy infrastructure. When run anywhere
inside an existing apps/<site-name> WordPress tree, site-name may be omitted.

Options:
  -t, --title TITLE          Site title (default: title-cased site name)
  -u, --url URL              WordPress URL (default: https://<site-name>.test)
      --profile NAME         Load an available profile: bare, clean, custom,
                             or loaded (`--preset` is an alias)
      --list-profiles        List available profiles and exit
  -p, --plugin SOURCE        Install an additional plugin; repeatable. SOURCE may be:
                               wp:query-monitor (or a bare WordPress.org slug)
                               git:git@github.com:owner/plugin.git
                               git:https://github.com/owner/plugin.git
      --github-plugin NAME   Clone NAME from GITHUB_USER over SSH; repeatable
      --plugins-file FILE    Read one plugin source per line (# comments allowed)
  -r, --restore FILE         Install normally, then restore a .wpress archive with
                             native AIOWM. Existing sites receive a native AIOWM
                             safety backup before replacement.
      --build                Rebuild the PHP image before starting services
  -f, --force                Replace an existing site and reset its database;
                             the old directory is moved to apps/.devarch-backups
      --no-hosts             Do not register <site-name>.test in the system hosts file
      --scaffolds-only       Provision Maker workspaces in an existing site only
      --project-slug SLUG    Override the workspace slug derived from site-name
      --php-namespace NS     Override the PHP namespace (default: Maker\\<SiteName>)
      --js-namespace NS      Override the block namespace (default: project slug)
      --dry-run              Validate and print the plan without changing anything
  -h, --help                 Show this help

Authentication:
  Private Git plugins use your host Git/SSH credentials because repositories are
  cloned on the host. Credential-bearing HTTPS URLs are rejected. WordPress and
  database credentials come from .env (ADMIN_*, MARIADB_ROOT_PASSWORD) or exported
  WP_ADMIN_* variables.

Examples:
  scripts/wordpress/bootstrap.sh my-site --profile clean
  scripts/wordpress/bootstrap.sh my-site --plugin wp:query-monitor
  scripts/wordpress/bootstrap.sh my-site --github-plugin makerblocks
  scripts/wordpress/bootstrap.sh my-site --plugins-file scripts/wordpress/plugins.example
  scripts/wordpress/bootstrap.sh my-site --restore /backups/site.wpress
EOF
}

title_case() {
  printf '%s' "$1" | tr '_-' '  ' | awk '{ for (i=1; i<=NF; i++) $i=toupper(substr($i,1,1)) substr($i,2); print }'
}

print_command() {
  printf '  +'
  printf ' %q' "$@"
  printf '\n'
}

run() {
  if [[ "$DRY_RUN" == true ]]; then
    print_command "$@"
  else
    "$@"
  fi
}

add_plugin() {
  PLUGIN_SOURCES+=("$1")
  PLUGIN_ACTIVATIONS+=("${2:-true}")
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

load_profile_file() {
  local file="$1" kind value option extra source
  while read -r kind value option extra || [[ -n "${kind:-}" ]]; do
    [[ -z "${kind:-}" || "$kind" == \#* ]] && continue
    [[ -n "${value:-}" && -z "${extra:-}" ]] || die "invalid profile entry in $file: $kind ${value:-} ${option:-} ${extra:-}"
    case "$kind" in
      include)
        [[ -z "${option:-}" && "$value" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || die "invalid profile include in $file: $value ${option:-}"
        [[ -f "$PROFILE_DIR/$value" ]] || die "profile include not found: $value"
        load_profile_file "$PROFILE_DIR/$value"
        ;;
      maker-stack-channel)
        [[ -z "${option:-}" && "$value" =~ ^[a-z][a-z0-9-]*$ ]] || die "invalid Maker stack channel in $file: $value ${option:-}"
        [[ -z "$MAKER_STACK_CHANNEL" || "$MAKER_STACK_CHANNEL" == "$value" ]] || die "conflicting Maker stack channels in profile '$PROFILE'"
        MAKER_STACK_CHANNEL="$value"
        ;;
      maker-core)
        [[ "$value" == plugin || "$value" == theme ]] || die "invalid Maker core package type in $file: $value"
        [[ "${option:-}" =~ ^[A-Za-z0-9._-]+$ ]] || die "invalid Maker core repository in $file: ${option:-}"
        [[ -n "${GITHUB_USER:-}" && "${GITHUB_USER:-}" != github-user ]] || die "profile '$PROFILE' requires GITHUB_USER"
        source="git@github.com:${GITHUB_USER}/${option}.git"
        MAKER_CORE_TYPES+=("$value")
        MAKER_CORE_SLUGS+=("$option")
        MAKER_CORE_URLS+=("$source")
        if [[ "$value" == plugin ]]; then add_plugin "git:$source" true; else THEME_SOURCES+=("$source"); fi
        ;;
      maker-workspace)
        [[ "$value" == child-theme || "$value" == blocks-plugin || "$value" == app-plugin ]] || die "invalid Maker workspace type in $file: $value"
        [[ -n "${option:-}" ]] || die "Maker workspace requires a core package in $file: $value"
        MAKER_WORKSPACES+=("$value:$option")
        ;;
      github-plugin|github-theme|github-mu-plugin)
        [[ -n "${GITHUB_USER:-}" && "${GITHUB_USER:-}" != github-user ]] || die "profile '$PROFILE' requires GITHUB_USER"
        [[ "$value" =~ ^[A-Za-z0-9._-]+$ ]] || die "invalid repository name in profile '$PROFILE': $value"
        source="git@github.com:${GITHUB_USER}/$value.git"
        case "$kind" in
          github-plugin)
            [[ -z "${option:-}" || "$option" == active || "$option" == inactive ]] || die "invalid activation '$option' for $value"
            add_plugin "git:$source" "$([[ "${option:-active}" == inactive ]] && echo false || echo true)"
            ;;
          github-theme)
            [[ -z "${option:-}" ]] || die "github-theme does not accept an option: $value $option"
            THEME_SOURCES+=("$source")
            ;;
          github-mu-plugin)
            [[ -z "${option:-}" ]] || die "github-mu-plugin does not accept an option: $value $option"
            MU_PLUGIN_SOURCES+=("$source")
            ;;
        esac
        ;;
      wp-plugin)
        [[ -z "${option:-}" ]] || die "wp-plugin does not accept an option: $value $option"
        [[ "$value" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || die "invalid WordPress.org plugin in profile '$PROFILE': $value"
        add_plugin "wp:$value" true
        ;;
      *) die "unknown profile directive '$kind' in $file" ;;
    esac
  done < "$file"
}

load_profile() {
  [[ -z "$PROFILE" ]] && return 0
  [[ "$PROFILE" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || die "invalid profile name: $PROFILE"
  local file="$PROFILE_DIR/$PROFILE.profile"
  [[ -f "$file" ]] || die "unknown profile: $PROFILE (use --list-profiles)"
  load_profile_file "$file"
}

parse_args() {
  if [[ $# -eq 0 ]]; then
    usage
    exit 1
  fi

  if [[ "$1" != -* ]]; then
    SITE_NAME="$1"
    shift
  fi

  while [[ $# -gt 0 ]]; do
    case "$1" in
      -t|--title)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        SITE_TITLE="$2"
        shift 2
        ;;
      -u|--url)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        SITE_URL="$2"
        shift 2
        ;;
      --profile|--preset)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        PROFILE="$2"
        shift 2
        ;;
      --list-profiles)
        list_profiles
        exit 0
        ;;
      -p|--plugin)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        add_plugin "$2" true
        shift 2
        ;;
      --github-plugin)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        [[ -n "${GITHUB_USER:-}" && "${GITHUB_USER:-}" != github-user ]] || die "--github-plugin requires GITHUB_USER"
        [[ "$2" =~ ^[A-Za-z0-9._-]+$ ]] || die "invalid GitHub plugin name: $2"
        add_plugin "git:git@github.com:${GITHUB_USER}/$2.git" true
        shift 2
        ;;
      --plugins-file)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        PLUGINS_FILE="$2"
        shift 2
        ;;
      -r|--restore)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        RESTORE_FILE="$2"
        shift 2
        ;;
      --build)
        BUILD=true
        shift
        ;;
      --no-server)
        # Backward-compatible no-op: infrastructure routing is now always used.
        shift
        ;;
      -f|--force)
        FORCE=true
        shift
        ;;
      --no-hosts)
        REGISTER_HOSTS=false
        shift
        ;;
      --scaffolds-only)
        SCAFFOLDS_ONLY=true
        shift
        ;;
      --project-slug)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        PROJECT_SLUG="$2"
        shift 2
        ;;
      --php-namespace)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        PHP_NAMESPACE="$2"
        shift 2
        ;;
      --js-namespace)
        [[ $# -ge 2 ]] || die "$1 requires a value"
        JS_NAMESPACE="$2"
        shift 2
        ;;
      --dry-run)
        DRY_RUN=true
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        die "unknown option: $1"
        ;;
    esac
  done
}

discover_site_name() {
  [[ -z "$SITE_NAME" ]] || return 0

  local candidate
  candidate="$(pwd -P)"
  while [[ "$candidate" != / ]]; do
    if [[ -f "$candidate/wp-config.php" ]]; then
      [[ "$(dirname "$candidate")" == "$APPS_DIR" ]] || die "detected WordPress root is not directly under $APPS_DIR: $candidate"
      SITE_NAME="$(basename "$candidate")"
      return
    fi
    candidate="$(dirname "$candidate")"
  done
}

namespace_class_from_slug() {
  local slug="$1" part result=""
  IFS='-' read -ra parts <<< "$slug"
  for part in "${parts[@]}"; do result+="${part^}"; done
  printf '%s' "$result"
}

validate_config() {
  discover_site_name
  [[ -n "$SITE_NAME" ]] || die "site-name is required outside an existing apps/<site-name> WordPress tree"
  [[ "$SITE_NAME" =~ ^[a-z0-9][a-z0-9-]{0,59}$ ]] || die "site-name must match [a-z0-9][a-z0-9-]{0,59} for .test routing"

  PROJECT_SLUG="${PROJECT_SLUG:-$SITE_NAME}"
  PHP_NAMESPACE="${PHP_NAMESPACE:-Maker\\$(namespace_class_from_slug "$PROJECT_SLUG")}"
  JS_NAMESPACE="${JS_NAMESPACE:-$PROJECT_SLUG}"
  [[ "$PROJECT_SLUG" =~ ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ ]] || die "project slug must be lowercase kebab case and start with a letter"
  [[ "$PHP_NAMESPACE" =~ ^[A-Za-z_][A-Za-z0-9_]*(\\[A-Za-z_][A-Za-z0-9_]*)+$ ]] || die "PHP namespace must contain at least two valid namespace segments"
  [[ "$JS_NAMESPACE" =~ ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ ]] || die "JS namespace must be lowercase kebab case and start with a letter"

  SITE_TITLE="${SITE_TITLE:-$(title_case "$SITE_NAME") }"
  SITE_TITLE="${SITE_TITLE% }"
  SITE_URL="${SITE_URL:-https://$SITE_NAME.test}"
  [[ "$SITE_URL" =~ ^https?://[^[:space:]]+$ ]] || die "URL must start with http:// or https:// and contain no spaces"

  if [[ -n "$RESTORE_FILE" ]]; then
    [[ -n "$AIOWM_GIT_URL" ]] || die "restore requires AIOWM_GIT_URL or GITHUB_USER for the established native-CLI AIOWM repository"
    [[ "$AIOWM_GIT_URL" == git@* || "$AIOWM_GIT_URL" == ssh://* || "$AIOWM_GIT_URL" == https://* ]] || die "unsupported AIOWM Git URL: $AIOWM_GIT_URL"
    [[ ! "$AIOWM_GIT_URL" =~ ^https?://[^/@]+@ ]] || die "credential-bearing AIOWM Git URLs are not allowed"
    [[ "$RESTORE_FILE" == *.wpress ]] || die "restore file must use the .wpress extension: $RESTORE_FILE"
    [[ "$(basename "$RESTORE_FILE")" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*\.wpress$ ]] || die "restore filename contains unsupported characters: $RESTORE_FILE"
    [[ -f "$RESTORE_FILE" ]] || die "restore file not found: $RESTORE_FILE"
    RESTORE_FILE="$(cd "$(dirname "$RESTORE_FILE")" && pwd -P)/$(basename "$RESTORE_FILE")"
  fi

  load_profile
  if [[ "$SCAFFOLDS_ONLY" == true ]]; then
    ((${#MAKER_WORKSPACES[@]} > 0)) || die "--scaffolds-only requires a Maker-enabled profile"
    [[ -z "$RESTORE_FILE" && "$FORCE" == false ]] || die "--scaffolds-only cannot be combined with restore or force"
    [[ "$DRY_RUN" == true || -f "$APPS_DIR/$SITE_NAME/wp-config.php" ]] || die "existing WordPress site not found: $APPS_DIR/$SITE_NAME"
  fi

  if [[ -n "$PLUGINS_FILE" ]]; then
    [[ -f "$PLUGINS_FILE" ]] || die "plugins file not found: $PLUGINS_FILE"
    while IFS= read -r source || [[ -n "$source" ]]; do
      source="${source%%#*}"
      source="${source#"${source%%[![:space:]]*}"}"
      source="${source%"${source##*[![:space:]]}"}"
      [[ -z "$source" ]] || add_plugin "$source" true
    done < "$PLUGINS_FILE"
  fi

  local source
  for source in "${PLUGIN_SOURCES[@]}"; do
    if [[ "$source" == git:* ]]; then
      local url="${source#git:}"
      [[ "$url" == git@* || "$url" == ssh://* || "$url" == https://* ]] || die "unsupported Git plugin URL: $url"
      [[ ! "$url" =~ ^https?://[^/@]+@ ]] || die "credential-bearing Git URLs are not allowed; use SSH or a credential helper"
    else
      local slug="${source#wp:}"
      [[ "$slug" =~ ^[a-z0-9][a-z0-9._-]*$ ]] || die "invalid WordPress.org plugin slug: $slug"
    fi
  done

  [[ -f "$PHP_COMPOSE" ]] || die "required Compose file not found: $PHP_COMPOSE"
  [[ -f "$MARIADB_COMPOSE" ]] || die "required Compose file not found: $MARIADB_COMPOSE"
  [[ -f "$PROXY_COMPOSE" ]] || die "required Compose file not found: $PROXY_COMPOSE"

  if [[ "$DRY_RUN" != true ]]; then
    [[ -n "$ADMIN_PASSWORD_VALUE" ]] || die "set ADMIN_PASSWORD or WP_ADMIN_PASSWORD in .env"
    command -v git >/dev/null 2>&1 || die "git is required"
  fi
}

detect_runtime() {
  if [[ "$DRY_RUN" == true ]]; then
    RUNTIME="${RUNTIME:-podman}"
    CONTAINER_USER="${CONTAINER_USER:-$([[ "$RUNTIME" == podman ]] && printf '0:0' || printf '%s:%s' "$(id -u)" "$(id -g)")}"
    COMPOSE=("$RUNTIME" compose)
    return
  fi

  if [[ -z "$RUNTIME" ]]; then
    if command -v podman >/dev/null 2>&1; then
      RUNTIME=podman
    elif command -v docker >/dev/null 2>&1; then
      RUNTIME=docker
    else
      die "Podman or Docker is required"
    fi
  fi

  command -v "$RUNTIME" >/dev/null 2>&1 || die "container runtime not found: $RUNTIME"
  # Root in a rootless Podman container maps to the invoking host user and owns
  # bind-mounted files. Docker needs the explicit host UID/GID instead.
  CONTAINER_USER="${CONTAINER_USER:-$([[ "$RUNTIME" == podman ]] && printf '0:0' || printf '%s:%s' "$(id -u)" "$(id -g)")}"
  if "$RUNTIME" compose version >/dev/null 2>&1; then
    COMPOSE=("$RUNTIME" compose)
  elif [[ "$RUNTIME" == podman ]] && command -v podman-compose >/dev/null 2>&1; then
    COMPOSE=(podman-compose)
  else
    die "no Compose provider found for $RUNTIME"
  fi
}

ensure_network() {
  if [[ "$DRY_RUN" == true ]]; then
    run "$RUNTIME" network create microservices-net
  elif ! "$RUNTIME" network inspect microservices-net >/dev/null 2>&1; then
    run "$RUNTIME" network create microservices-net >/dev/null
  fi
}

start_services() {
  log "start PHP, MariaDB, and Nginx Proxy Manager services"
  local php_args=(-f "$PHP_COMPOSE" up -d)
  [[ "$BUILD" == true ]] && php_args+=(--build)
  run "${COMPOSE[@]}" "${php_args[@]}"
  if [[ "$DRY_RUN" == true ]]; then
    print_command env 'MARIADB_ROOT_PASSWORD=<redacted>' "${COMPOSE[@]}" -f "$MARIADB_COMPOSE" up -d
  else
    env MARIADB_ROOT_PASSWORD="$DB_ROOT_PASSWORD" "${COMPOSE[@]}" -f "$MARIADB_COMPOSE" up -d
  fi
  run "${COMPOSE[@]}" -f "$PROXY_COMPOSE" up -d
}

wait_for_services() {
  [[ "$DRY_RUN" == true ]] && { log "wait for PHP/WP-CLI, MariaDB, and Nginx Proxy Manager readiness"; return; }

  local attempt
  for attempt in {1..90}; do
    if "$RUNTIME" exec "$PHP_CONTAINER" wp --info >/dev/null 2>&1 && \
       "$RUNTIME" exec "$MARIADB_CONTAINER" sh -c 'mariadb-admin ping -uroot -p"$MARIADB_ROOT_PASSWORD" --silent' >/dev/null 2>&1 && \
       "$RUNTIME" exec nginx-proxy-manager curl -fsS http://localhost:81/api/ >/dev/null 2>&1; then
      return
    fi
    sleep 1
  done
  die "PHP, MariaDB, or Nginx Proxy Manager did not become ready within 90 seconds"
}

prepare_site_dir() {
  local site_dir="$APPS_DIR/$SITE_NAME"
  if [[ -e "$site_dir" ]]; then
    [[ "$FORCE" == true || -n "$RESTORE_FILE" ]] || die "$site_dir already exists; use --force to replace it"
    REPLACE_EXISTING=true
    local backup_dir="$APPS_DIR/.devarch-backups/${SITE_NAME}-$(date +%Y%m%d-%H%M%S)"
    log "move existing site to $backup_dir"
    run mkdir -p "$(dirname "$backup_dir")"
    run mv "$site_dir" "$backup_dir"
  fi
  run mkdir -p "$site_dir"
}

generate_db_password() {
  if [[ "$DRY_RUN" == true ]]; then
    printf 'dry-run-password'
  elif command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 18
  else
    od -An -N18 -tx1 /dev/urandom | tr -d ' \n'
  fi
}

reset_database() {
  local db_name="$1" db_user="$2" db_password="$3"
  log "create isolated database and user: $db_name / $db_user"
  [[ "$DRY_RUN" == true ]] && return

  local drop_sql=""
  [[ "$FORCE" == true || "$REPLACE_EXISTING" == true ]] && drop_sql="DROP DATABASE IF EXISTS \`$db_name\`;"
  local sql="$drop_sql
CREATE DATABASE IF NOT EXISTS \`$db_name\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '$db_user'@'%' IDENTIFIED BY '$db_password';
ALTER USER '$db_user'@'%' IDENTIFIED BY '$db_password';
GRANT ALL PRIVILEGES ON \`$db_name\`.* TO '$db_user'@'%';
FLUSH PRIVILEGES;"
  printf '%s\n' "$sql" | "$RUNTIME" exec -i "$MARIADB_CONTAINER" sh -c 'mariadb -uroot -p"$MARIADB_ROOT_PASSWORD"' >/dev/null
}

wp_exec() {
  local container_site="$1"
  shift
  run "$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp "$PHP_CONTAINER" \
    wp --path="$container_site" --allow-root "$@"
}

wp_is_successful() {
  local container_site="$1"
  shift
  "$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp "$PHP_CONTAINER" \
    wp --path="$container_site" --allow-root "$@" >/dev/null 2>&1
}

wp_prompt_secret() {
  local secret="$1" description="$2" container_site="$3"
  shift 3
  if [[ "$DRY_RUN" == true ]]; then
    log "$description (secret supplied on stdin; redacted)"
    print_command "$RUNTIME" exec -i --user "$CONTAINER_USER" "$PHP_CONTAINER" wp --path="$container_site" --allow-root "$@"
  else
    # WP-CLI echoes prompt-derived arguments when stdin is not a TTY, including
    # passwords. Suppress its output for this command so bootstrap logs remain safe.
    if printf '%s\n' "$secret" | "$RUNTIME" exec -i --user "$CONTAINER_USER" -e HOME=/tmp "$PHP_CONTAINER" \
      wp --path="$container_site" --allow-root "$@" >/dev/null 2>&1; then
      log "$description complete"
    else
      die "$description failed (WP-CLI output suppressed because it may contain secrets)"
    fi
  fi
}

configure_wordpress_content_settings() {
  local container_site="/var/www/html/$SITE_NAME"
  local host_uploads="$APPS_DIR/$SITE_NAME/wp-content/uploads"

  wp_exec "$container_site" option update uploads_use_yearmonth_folders 0
  log "remove empty upload subdirectories after disabling year/month folders"
  run mkdir -p "$host_uploads"
  run find "$host_uploads" -mindepth 1 -depth -type d -empty -delete
  wp_exec "$container_site" rewrite structure '/%postname%/' --hard
}

install_wordpress() {
  local container_site="/var/www/html/$SITE_NAME"
  local db_name="wp_${SITE_NAME//-/_}"
  local db_user="${db_name:0:32}"
  local db_password
  db_password="$(generate_db_password)"

  reset_database "$db_name" "$db_user" "$db_password"
  log "download and install WordPress"
  wp_exec "$container_site" core download
  wp_prompt_secret "$db_password" "create wp-config.php" "$container_site" \
    config create --dbname="$db_name" --dbuser="$db_user" --dbhost="$MARIADB_CONTAINER" --skip-check --prompt=dbpass
  # Local sites use bind-mounted, writable content; avoid WordPress requesting FTP credentials.
  wp_exec "$container_site" config set FS_METHOD direct
  wp_prompt_secret "$ADMIN_PASSWORD_VALUE" "install WordPress administrator" "$container_site" \
    core install --url="$SITE_URL" --title="$SITE_TITLE" --admin_user="$ADMIN_USER_VALUE" --admin_email="$ADMIN_EMAIL_VALUE" --prompt=admin_password
  configure_wordpress_content_settings
  wp_exec "$container_site" post delete 1 --force
  wp_exec "$container_site" plugin delete akismet hello || true
}

component_slug_from_url() {
  local url="$1" slug
  slug="${url##*/}"
  slug="${slug%.git}"
  [[ "$slug" =~ ^[A-Za-z0-9._-]+$ ]] || die "cannot derive a safe component slug from $url"
  printf '%s' "$slug"
}

run_component_composer() {
  local container_dir="$1" slug="$2"
  log "run Composer when composer.json exists: $slug"
  if [[ "$DRY_RUN" == true ]]; then
    print_command "$RUNTIME" exec --user "$CONTAINER_USER" "$PHP_CONTAINER" composer install --working-dir="$container_dir"
  else
    "$RUNTIME" exec --user "$CONTAINER_USER" -e HOME=/tmp "$PHP_CONTAINER" sh -c \
      'test ! -f "$1/composer.json" || composer install --no-interaction --working-dir="$1"' _ "$container_dir"
  fi
}

ensure_aiowm() {
  local container_site="/var/www/html/$SITE_NAME"
  local host_plugin="$APPS_DIR/$SITE_NAME/wp-content/plugins/all-in-one-wp-migration"
  local host_backups="$APPS_DIR/$SITE_NAME/wp-content/ai1wm-backups"
  local host_storage="$host_plugin/storage"

  log "ensure established native-CLI All-in-One WP Migration repository is installed and active"
  if [[ "$DRY_RUN" == true ]]; then
    wp_exec "$container_site" plugin deactivate all-in-one-wp-migration
    print_command rm -rf "$host_plugin"
    print_command git clone --depth 1 "$AIOWM_GIT_URL" "$host_plugin"
  elif [[ ! -d "$host_plugin/.git" ]]; then
    wp_exec "$container_site" plugin deactivate all-in-one-wp-migration || true
    rm -rf "$host_plugin"
    run git clone --depth 1 "$AIOWM_GIT_URL" "$host_plugin"
  fi
  if [[ "$DRY_RUN" == true || -f "$host_plugin/composer.json" ]]; then
    run_component_composer "/var/www/html/$SITE_NAME/wp-content/plugins/all-in-one-wp-migration" "all-in-one-wp-migration"
  fi
  wp_exec "$container_site" plugin activate all-in-one-wp-migration

  log "prepare writable AIOWM backup and storage directories"
  run mkdir -p "$host_backups" "$host_storage"
  run chmod -R a+rwX "$host_backups" "$host_storage"
}

stage_restore_archive() {
  [[ -n "$RESTORE_FILE" ]] || return 0
  local site_dir="$APPS_DIR/$SITE_NAME"
  if [[ "$RESTORE_FILE" == "$site_dir"/* ]]; then
    local staging_dir="$APPS_DIR/.devarch-backups/imports"
    local staged="$staging_dir/${SITE_NAME}-$(date +%Y%m%d-%H%M%S)-$(basename "$RESTORE_FILE")"
    log "stage restore archive outside the site before replacement: $staged"
    run mkdir -p "$staging_dir"
    run cp "$RESTORE_FILE" "$staged"
    RESTORE_FILE="$staged"
  fi
}

backup_existing_site() {
  [[ -n "$RESTORE_FILE" && -f "$APPS_DIR/$SITE_NAME/wp-config.php" ]] || return 0
  local container_site="/var/www/html/$SITE_NAME"
  log "create native AIOWM safety backup of existing site"
  ensure_aiowm
  wp_exec "$container_site" ai1wm backup
}

restore_aiowm_archive() {
  [[ -n "$RESTORE_FILE" ]] || return 0
  local container_site="/var/www/html/$SITE_NAME"
  local host_backups="$APPS_DIR/$SITE_NAME/wp-content/ai1wm-backups"
  local archive_name
  archive_name="$(basename "$RESTORE_FILE")"

  ensure_aiowm
  log "copy restore archive into AIOWM backups: $archive_name"
  run cp "$RESTORE_FILE" "$host_backups/$archive_name"
  run chmod a+rw "$host_backups/$archive_name"
  log "restore archive with native AIOWM WP-CLI"
  wp_exec "$container_site" ai1wm restore "$archive_name"
  wp_exec "$container_site" option update home "$SITE_URL"
  wp_exec "$container_site" option update siteurl "$SITE_URL"
  configure_wordpress_content_settings
  wp_exec "$container_site" cache flush || true
}

register_makermaker_galaxy() {
  local host_makermaker="$APPS_DIR/$SITE_NAME/wp-content/plugins/makermaker"
  local host_typerocket="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins/typerocket-pro-v6"
  [[ "$DRY_RUN" == true || -d "$host_makermaker" ]] || return 0

  log "use MakerMaker runtime filter registration and backfill its plugin-specific Galaxy context"
  if [[ "$DRY_RUN" != true && -d "$host_typerocket/.git" ]]; then
    [[ -z "$(git -C "$host_typerocket" status --porcelain)" ]] || die "TypeRocket repository is dirty before MakerMaker Galaxy context registration"
  fi
  wp_exec "/var/www/html/$SITE_NAME" makermaker register-plugin-galaxy \
    --plugin=makermaker --namespace=Maker/MakerMaker
  if [[ "$DRY_RUN" != true && -d "$host_typerocket/.git" ]]; then
    [[ -z "$(git -C "$host_typerocket" status --porcelain)" ]] || die "MakerMaker Galaxy context registration changed tracked TypeRocket files"
  fi
}

install_plugins() {
  local i source activate
  local container_site="/var/www/html/$SITE_NAME"
  local host_plugins="$APPS_DIR/$SITE_NAME/wp-content/plugins"
  for i in "${!PLUGIN_SOURCES[@]}"; do
    source="${PLUGIN_SOURCES[$i]}"
    activate="${PLUGIN_ACTIVATIONS[$i]}"
    if [[ "$source" == git:* ]]; then
      local url="${source#git:}" slug target container_dir
      slug="$(component_slug_from_url "$url")"
      target="$host_plugins/$slug"
      container_dir="/var/www/html/$SITE_NAME/wp-content/plugins/$slug"
      log "clone Git plugin: $slug"
      if [[ -e "$target" ]]; then
        log "plugin directory already exists, skipping clone: $slug"
      else
        run git clone --depth 1 "$url" "$target"
      fi
      if [[ "$DRY_RUN" == true || -f "$target/composer.json" ]]; then
        run_component_composer "$container_dir" "$slug"
      fi
      if [[ "$activate" == true ]]; then
        wp_exec "$container_site" plugin activate "$slug"
      else
        log "plugin installed without activation: $slug"
      fi
    else
      local slug="${source#wp:}"
      log "install WordPress.org plugin: $slug"
      if [[ "$activate" == true ]]; then
        wp_exec "$container_site" plugin install "$slug" --activate
      else
        wp_exec "$container_site" plugin install "$slug"
      fi
    fi
  done
}

configure_site_galaxy_config() {
  local site_dir="$APPS_DIR/$SITE_NAME"
  local context="$site_dir/wp-content/plugins/makermaker/app/Generator/GalaxyContext.php"
  [[ "$DRY_RUN" == true || -f "$context" ]] || return 0
  log "write portable MakerMaker-owned site Galaxy launcher and resolver"
  if [[ "$DRY_RUN" == true ]]; then
    print_command php -r '<MakerMaker GalaxyContext::siteLauncher()/siteConfig()>' "$context" "$site_dir/galaxy" "$site_dir/galaxy-config.php"
  else
    php -r 'require $argv[1]; file_put_contents($argv[2], Maker\MakerMaker\Generator\GalaxyContext::siteLauncher()); file_put_contents($argv[3], Maker\MakerMaker\Generator\GalaxyContext::siteConfig()); chmod($argv[2], 0755); chmod($argv[3], 0644);' "$context" "$site_dir/galaxy" "$site_dir/galaxy-config.php"
  fi
}

install_mu_plugins() {
  local url slug target container_dir host_mu="$APPS_DIR/$SITE_NAME/wp-content/mu-plugins"
  run mkdir -p "$host_mu"
  for url in "${MU_PLUGIN_SOURCES[@]}"; do
    slug="$(component_slug_from_url "$url")"
    target="$host_mu/$slug"
    container_dir="/var/www/html/$SITE_NAME/wp-content/mu-plugins/$slug"
    log "clone Git must-use plugin: $slug"
    [[ -e "$target" ]] || run git clone --depth 1 "$url" "$target"
    if [[ "$DRY_RUN" == true || -f "$target/composer.json" ]]; then
      run_component_composer "$container_dir" "$slug"
    fi
    if [[ "$DRY_RUN" == true || -f "$target/$slug.php" ]]; then
      run cp "$target/$slug.php" "$host_mu/$slug.php"
    else
      die "must-use plugin entry file not found: $target/$slug.php"
    fi
  done
}

make_content_writable() {
  local host_content="$APPS_DIR/$SITE_NAME/wp-content"
  log "make local wp-content writable by PHP"
  run chmod -R a+rwX "$host_content"
}

install_themes() {
  local url slug target container_dir
  local container_site="/var/www/html/$SITE_NAME"
  local host_themes="$APPS_DIR/$SITE_NAME/wp-content/themes"
  for url in "${THEME_SOURCES[@]}"; do
    slug="$(component_slug_from_url "$url")"
    target="$host_themes/$slug"
    container_dir="/var/www/html/$SITE_NAME/wp-content/themes/$slug"
    log "clone Git theme: $slug"
    [[ -e "$target" ]] || run git clone --depth 1 "$url" "$target"
    if [[ "$DRY_RUN" == true || -f "$target/composer.json" ]]; then
      run_component_composer "$container_dir" "$slug"
    fi
    wp_exec "$container_site" theme activate "$slug"
  done

  if ((${#THEME_SOURCES[@]} > 0)); then
    log "delete inactive bundled themes"
    wp_exec "$container_site" theme delete --all
  fi
}

mark_maker_core_packages() {
  local i root marker
  for i in "${!MAKER_CORE_SLUGS[@]}"; do
    if [[ "${MAKER_CORE_TYPES[$i]}" == theme ]]; then
      root="$APPS_DIR/$SITE_NAME/wp-content/themes/${MAKER_CORE_SLUGS[$i]}"
    else
      root="$APPS_DIR/$SITE_NAME/wp-content/plugins/${MAKER_CORE_SLUGS[$i]}"
    fi
    marker="$root/CORE-BOUNDARY.md"
    log "core install marker: ${MAKER_CORE_SLUGS[$i]}"
    if [[ "$DRY_RUN" != true ]]; then
      [[ -f "$marker" ]] || die "Maker core marker missing: $marker"
      grep -Fq 'FRAMEWORK CORE — DO NOT EDIT; update from playground releases.' "$marker" || die "Maker core marker is invalid: $marker"
    fi
  done
}

render_scaffold_tokens() {
  local root="$1" file
  while IFS= read -r -d '' file; do
    php -r '$p=$argv[1]; $s=file_get_contents($p); $s=strtr($s, array("{{SITE_TITLE}}"=>$argv[2], "{{PROJECT_SLUG}}"=>$argv[3], "{{JS_NAMESPACE}}"=>$argv[4])); file_put_contents($p, $s);' \
      "$file" "$SITE_TITLE" "$PROJECT_SLUG" "$JS_NAMESPACE"
  done < <(find "$root" -type f -print0)
}

initialize_workspace_git() {
  local target="$1"
  if [[ "$DRY_RUN" == true ]]; then
    print_command git -C "$target" init -q -b main
  else
    run git -C "$target" init -q -b main
  fi
}

write_maker_lock_manifest() {
  ((${#MAKER_CORE_SLUGS[@]} > 0)) || return 0
  local manifest="$APPS_DIR/$SITE_NAME/.devarch-maker.lock" i root commit resolved installed
  if [[ -e "$manifest" ]]; then
    log "Maker lock manifest already exists, preserving install record"
    return 0
  fi
  log "write Maker core lock manifest: $manifest"
  if [[ "$DRY_RUN" == true ]]; then
    print_command sh -c "write Maker lock manifest to '$manifest'"
    return 0
  fi
  installed="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  {
    printf 'repository_url\tresolved_ref\tcommit\tpackage_type\tinstalled_at\n'
    for i in "${!MAKER_CORE_SLUGS[@]}"; do
      if [[ "${MAKER_CORE_TYPES[$i]}" == theme ]]; then
        root="$APPS_DIR/$SITE_NAME/wp-content/themes/${MAKER_CORE_SLUGS[$i]}"
      else
        root="$APPS_DIR/$SITE_NAME/wp-content/plugins/${MAKER_CORE_SLUGS[$i]}"
      fi
      [[ -d "$root/.git" ]] || die "Maker core checkout missing for lock manifest: $root"
      commit="$(git -C "$root" rev-parse HEAD)"
      resolved="$(git -C "$root" describe --tags --exact-match 2>/dev/null || printf '%s' "$commit")"
      printf '%s\t%s\t%s\t%s\t%s\n' "${MAKER_CORE_URLS[$i]}" "$resolved" "$commit" "${MAKER_CORE_TYPES[$i]}" "$installed"
    done
  } > "$manifest.tmp"
  mv "$manifest.tmp" "$manifest"
}

publish_child_theme_workspace() {
  local target="$APPS_DIR/$SITE_NAME/wp-content/themes/$PROJECT_SLUG-theme"
  if [[ -e "$target" ]]; then
    log "workspace exists; refusing child-theme generation: $target"
  elif [[ "$DRY_RUN" == true ]]; then
    log "workspace create [child-theme]: $target"
    print_command mkdir "$target"
    initialize_workspace_git "$target"
  else
    (
      local stage="${target}.devarch-stage.$$"
      local scaffold="$APPS_DIR/$SITE_NAME/wp-content/themes/makerstarter/scaffolds/child-theme"
      trap 'rm -rf "$stage"' EXIT
      if [[ ! -d "$scaffold" && -d "$PROJECT_ROOT/apps/playground/wp-content/themes/makerstarter/scaffolds/child-theme" ]]; then
        scaffold="$PROJECT_ROOT/apps/playground/wp-content/themes/makerstarter/scaffolds/child-theme"
        log "use released playground MakerStarter scaffold for existing core without a scaffold"
      fi
      [[ -d "$scaffold" ]] || die "MakerStarter child scaffold missing: $scaffold"
      mkdir -p "$stage"
      cp -a "$scaffold/." "$stage/"
      render_scaffold_tokens "$stage"
      initialize_workspace_git "$stage"
      [[ "${DEVARCH_SCAFFOLD_FAIL_AFTER_STAGE:-false}" != true ]] || return 70
      mv "$stage" "$target"
      trap - EXIT
    ) || die "child-theme scaffold failed; staged files rolled back"
    log "workspace created [child-theme]: $target"
  fi
  wp_exec "/var/www/html/$SITE_NAME" theme activate "$PROJECT_SLUG-theme"
}

publish_blocks_workspace() {
  local target="$APPS_DIR/$SITE_NAME/wp-content/plugins/$PROJECT_SLUG-blocks"
  if [[ -e "$target" ]]; then
    log "workspace exists; refusing blocks-plugin generation: $target"
  elif [[ "$DRY_RUN" == true ]]; then
    log "workspace create [blocks-plugin]: $target (namespace: $JS_NAMESPACE)"
    print_command mkdir "$target"
    initialize_workspace_git "$target"
  else
    (
      local stage="${target}.devarch-stage.$$"
      local scaffold="$APPS_DIR/$SITE_NAME/wp-content/plugins/makerblocks/scaffolds/project-plugin"
      trap 'rm -rf "$stage"' EXIT
      if [[ ! -d "$scaffold" && -d "$PROJECT_ROOT/apps/playground/wp-content/plugins/makerblocks/scaffolds/project-plugin" ]]; then
        scaffold="$PROJECT_ROOT/apps/playground/wp-content/plugins/makerblocks/scaffolds/project-plugin"
        log "use released playground MakerBlocks scaffold for existing core without a scaffold"
      fi
      [[ -d "$scaffold" ]] || die "MakerBlocks project scaffold missing: $scaffold"
      mkdir -p "$stage"
      cp -a "$scaffold/." "$stage/"
      render_scaffold_tokens "$stage"
      mv "$stage/plugin.php" "$stage/$PROJECT_SLUG-blocks.php"
      initialize_workspace_git "$stage"
      [[ "${DEVARCH_SCAFFOLD_FAIL_AFTER_STAGE:-false}" != true ]] || return 70
      mv "$stage" "$target"
      trap - EXIT
    ) || die "blocks-plugin scaffold failed; staged files rolled back"
    log "workspace created [blocks-plugin]: $target"
  fi
  wp_exec "/var/www/html/$SITE_NAME" plugin activate "$PROJECT_SLUG-blocks"
}

publish_app_workspace() {
  local target="$APPS_DIR/$SITE_NAME/wp-content/plugins/$PROJECT_SLUG-app"
  if [[ -e "$target" ]]; then
    log "workspace exists; refusing app-plugin generation: $target"
    wp_exec "/var/www/html/$SITE_NAME" plugin activate "$PROJECT_SLUG-app"
    return
  fi
  log "workspace create [app-plugin via MakerMaker]: $target (namespace: $PHP_NAMESPACE)"
  wp_exec "/var/www/html/$SITE_NAME" makermaker create "$PROJECT_SLUG-app" \
    --name="$SITE_TITLE App" --namespace="$PHP_NAMESPACE" --vendor=maker --activate
  if [[ "$DRY_RUN" != true ]]; then
    [[ -d "$target" ]] || die "MakerMaker did not publish expected app workspace: $target"
    if [[ ! -e "$target/README.md" ]]; then
      printf '# PROJECT OWNED — EDIT HERE\n\nApplication workspace generated by MakerMaker.\n' > "$target/README.md"
    elif ! grep -q 'PROJECT OWNED — EDIT HERE' "$target/README.md"; then
      { printf '# PROJECT OWNED — EDIT HERE\n\n'; cat "$target/README.md"; } > "$target/README.md.tmp"
      mv "$target/README.md.tmp" "$target/README.md"
    fi
  fi
  initialize_workspace_git "$target"
}

provision_maker_workspaces() {
  ((${#MAKER_WORKSPACES[@]} > 0)) || return 0
  local declaration kind source
  log "Maker project: slug=$PROJECT_SLUG PHP=$PHP_NAMESPACE JS=$JS_NAMESPACE"
  for declaration in "${MAKER_WORKSPACES[@]}"; do
    kind="${declaration%%:*}"
    source="${declaration#*:}"
    log "workspace declaration: $kind from $source"
    case "$kind" in
      child-theme) publish_child_theme_workspace ;;
      blocks-plugin) publish_blocks_workspace ;;
      app-plugin) publish_app_workspace ;;
    esac
  done
}

register_site_host() {
  local hostname="$SITE_NAME.test"
  [[ "$REGISTER_HOSTS" == true ]] || { log "hosts registration skipped: $hostname"; return 0; }
  if [[ "$DRY_RUN" == true ]]; then
    log "register local host: 127.0.0.1 $hostname"
    return 0
  fi
  if [[ ! -x "$HOSTS_HELPER" ]]; then
    log "warning: hosts helper is unavailable; manually map 127.0.0.1 $hostname"
    return 0
  fi
  if ! "$HOSTS_HELPER" "$hostname"; then
    log "warning: could not register $hostname; manually map it to 127.0.0.1"
  fi
}

main() {
  parse_args "$@"
  validate_config
  detect_runtime

  log "site: $SITE_NAME ($SITE_TITLE)"
  log "URL: $SITE_URL"
  [[ -z "$PROFILE" ]] || log "profile: $PROFILE"
  log "plugins: ${#PLUGIN_SOURCES[@]}, must-use plugins: ${#MU_PLUGIN_SOURCES[@]}, themes: ${#THEME_SOURCES[@]}"
  [[ -z "$RESTORE_FILE" ]] || log "restore: $RESTORE_FILE (native AIOWM)"
  [[ "$DRY_RUN" == true ]] && log "dry run; no changes will be made"

  if [[ "$SCAFFOLDS_ONLY" == true ]]; then
    provision_maker_workspaces
    log "Maker workspaces ready: $APPS_DIR/$SITE_NAME"
    return
  fi

  ensure_network
  start_services
  wait_for_services
  stage_restore_archive
  backup_existing_site
  prepare_site_dir
  install_wordpress
  install_mu_plugins
  install_plugins
  configure_site_galaxy_config
  register_makermaker_galaxy
  install_themes
  mark_maker_core_packages
  write_maker_lock_manifest
  provision_maker_workspaces
  make_content_writable
  restore_aiowm_archive
  make_content_writable
  register_site_host

  log "ready through the Nginx Proxy Manager .test reverse proxy: $SITE_URL"
  log "admin: $SITE_URL/wp-admin (user: $ADMIN_USER_VALUE)"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
