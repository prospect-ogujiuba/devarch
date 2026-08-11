#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
APPS_DIR="$PROJECT_ROOT/apps"
PHP_COMPOSE="$PROJECT_ROOT/services-library/backend/php/compose.yml"
MARIADB_COMPOSE="$PROJECT_ROOT/services-library/database/mariadb/compose.yml"
ENV_FILE="$PROJECT_ROOT/.env"
PROFILE_DIR="$SCRIPT_DIR/profiles"

if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

DRY_RUN=false
FORCE=false
BUILD=false
SITE_NAME=""
SITE_TITLE=""
SITE_URL=""
PROFILE=""
PLUGINS_FILE=""
PLUGIN_SOURCES=()
PLUGIN_ACTIVATIONS=()
THEME_SOURCES=()
MU_PLUGIN_SOURCES=()
RUNTIME="${CONTAINER_RUNTIME:-}"
CONTAINER_USER="${WORDPRESS_CONTAINER_USER:-}"
PHP_CONTAINER="php"
MARIADB_CONTAINER="mariadb"
ADMIN_USER_VALUE="${WP_ADMIN_USER:-${ADMIN_USER:-admin}}"
ADMIN_PASSWORD_VALUE="${WP_ADMIN_PASSWORD:-${ADMIN_PASSWORD:-}}"
ADMIN_EMAIL_VALUE="${WP_ADMIN_EMAIL:-${ADMIN_EMAIL:-admin@devarch.test}}"
DB_ROOT_PASSWORD="${MARIADB_ROOT_PASSWORD:-devarch}"
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
Usage: scripts/wordpress/bootstrap.sh <site-name> [options]

Create a local WordPress site in apps/<site-name> using the existing PHP,
MariaDB, and wildcard .test reverse-proxy infrastructure.

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
      --build                Rebuild the PHP image before starting services
  -f, --force                Replace an existing site and reset its database;
                             the old directory is moved to apps/.devarch-backups
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

load_profile() {
  [[ -z "$PROFILE" ]] && return 0
  [[ "$PROFILE" =~ ^[a-z0-9][a-z0-9_-]*$ ]] || die "invalid profile name: $PROFILE"
  local file="$PROFILE_DIR/$PROFILE.profile"
  [[ -f "$file" ]] || die "unknown profile: $PROFILE (use --list-profiles)"

  local kind value option extra
  while read -r kind value option extra || [[ -n "${kind:-}" ]]; do
    [[ -z "${kind:-}" || "$kind" == \#* ]] && continue
    [[ -n "${value:-}" && -z "${extra:-}" ]] || die "invalid profile entry in $file: $kind ${value:-} ${option:-} ${extra:-}"
    case "$kind" in
      github-plugin|github-theme|github-mu-plugin)
        [[ -n "${GITHUB_USER:-}" && "${GITHUB_USER:-}" != github-user ]] || die "profile '$PROFILE' requires GITHUB_USER"
        [[ "$value" =~ ^[A-Za-z0-9._-]+$ ]] || die "invalid repository name in profile '$PROFILE': $value"
        local source="git@github.com:${GITHUB_USER}/$value.git"
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

validate_config() {
  [[ -n "$SITE_NAME" ]] || die "site-name is required"
  [[ "$SITE_NAME" =~ ^[a-z0-9][a-z0-9-]{0,59}$ ]] || die "site-name must match [a-z0-9][a-z0-9-]{0,59} for .test routing"

  SITE_TITLE="${SITE_TITLE:-$(title_case "$SITE_NAME") }"
  SITE_TITLE="${SITE_TITLE% }"
  SITE_URL="${SITE_URL:-https://$SITE_NAME.test}"
  [[ "$SITE_URL" =~ ^https?://[^[:space:]]+$ ]] || die "URL must start with http:// or https:// and contain no spaces"

  load_profile

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
  log "start PHP and MariaDB services"
  local php_args=(-f "$PHP_COMPOSE" up -d)
  [[ "$BUILD" == true ]] && php_args+=(--build)
  run "${COMPOSE[@]}" "${php_args[@]}"
  if [[ "$DRY_RUN" == true ]]; then
    print_command env 'MARIADB_ROOT_PASSWORD=<redacted>' "${COMPOSE[@]}" -f "$MARIADB_COMPOSE" up -d
  else
    env MARIADB_ROOT_PASSWORD="$DB_ROOT_PASSWORD" "${COMPOSE[@]}" -f "$MARIADB_COMPOSE" up -d
  fi
}

wait_for_services() {
  [[ "$DRY_RUN" == true ]] && { log "wait for PHP/WP-CLI and MariaDB readiness"; return; }

  local attempt
  for attempt in {1..60}; do
    if "$RUNTIME" exec "$PHP_CONTAINER" wp --info >/dev/null 2>&1 && \
       "$RUNTIME" exec "$MARIADB_CONTAINER" sh -c 'mariadb-admin ping -uroot -p"$MARIADB_ROOT_PASSWORD" --silent' >/dev/null 2>&1; then
      return
    fi
    sleep 1
  done
  die "PHP or MariaDB did not become ready within 60 seconds"
}

prepare_site_dir() {
  local site_dir="$APPS_DIR/$SITE_NAME"
  if [[ -e "$site_dir" ]]; then
    [[ "$FORCE" == true ]] || die "$site_dir already exists; use --force to replace it"
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
  [[ "$FORCE" == true ]] && drop_sql="DROP DATABASE IF EXISTS \`$db_name\`;"
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
  wp_prompt_secret "$ADMIN_PASSWORD_VALUE" "install WordPress administrator" "$container_site" \
    core install --url="$SITE_URL" --title="$SITE_TITLE" --admin_user="$ADMIN_USER_VALUE" --admin_email="$ADMIN_EMAIL_VALUE" --prompt=admin_password
  wp_exec "$container_site" option update uploads_use_yearmonth_folders 0
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
}

main() {
  parse_args "$@"
  validate_config
  detect_runtime

  log "site: $SITE_NAME ($SITE_TITLE)"
  log "URL: $SITE_URL"
  [[ -z "$PROFILE" ]] || log "profile: $PROFILE"
  log "plugins: ${#PLUGIN_SOURCES[@]}, must-use plugins: ${#MU_PLUGIN_SOURCES[@]}, themes: ${#THEME_SOURCES[@]}"
  [[ "$DRY_RUN" == true ]] && log "dry run; no changes will be made"

  ensure_network
  start_services
  wait_for_services
  prepare_site_dir
  install_wordpress
  install_mu_plugins
  install_plugins
  install_themes

  log "ready through the existing .test reverse proxy: $SITE_URL"
  log "admin: $SITE_URL/wp-admin (user: $ADMIN_USER_VALUE)"
}

main "$@"
