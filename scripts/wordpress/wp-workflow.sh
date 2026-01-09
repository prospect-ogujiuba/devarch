#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$HOME/projects/devarch}"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"

if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(grep -E '^(GITHUB_TOKEN|GITHUB_USER)=' "$PROJECT_ROOT/.env" | xargs)
fi

print_usage() {
    cat << EOF
WordPress Multi System Workflow Manager

Usage: $(basename "$0") [COMMAND] [OPTIONS]

Commands:
    install         Install WordPress site
    prepare         Prepare migration directories and files
    activate        Activate All In One WP Migration plugin
    deactivate      Deactivate plugin and restore WP_DEBUG state
    remove          Remove existing WordPress site and database
    backup          Create backup of WordPress site
    restore         Restore WordPress site from backup
    full            Run complete workflow (install + prepare + activate)
    help            Show this help message

Install Options:
    -n, --name      Site name (default: myapp)
    -p, --profile   Profile type (default: bare)
    -f, --force     Force installation

Prepare Options:
    -n, --name      Site name (default: myapp)
    -s, --source    Source .wpress file path

Activate Options:
    -n, --name      Site name (default: myapp)

Deactivate Options:
    -n, --name      Site name (default: myapp)

Remove Options:
    -n, --name      Site name (default: myapp)

Backup Options:
    -n, --name      Site name (default: myapp)
    -d, --dest      Destination directory for backup file (optional)
    --exclude-spam-comments     Exclude spam comments
    --exclude-post-revisions    Exclude post revisions
    --exclude-media             Exclude media library
    --exclude-themes            Exclude all themes
    --exclude-inactive-themes   Exclude inactive themes
    --exclude-muplugins         Exclude must-use plugins
    --exclude-plugins           Exclude all plugins
    --exclude-inactive-plugins  Exclude inactive plugins
    --exclude-cache             Exclude cache files
    --exclude-database          Exclude database

Restore Options:
    -n, --name      Site name (default: myapp)
    -s, --source    Source .wpress backup file (required)

Examples:
    $(basename "$0") install -n myapp -p bare -f
    $(basename "$0") prepare -n b2bcnc -s /path/to/backup.wpress
    $(basename "$0") activate -n myapp
    $(basename "$0") deactivate -n myapp
    $(basename "$0") remove -n myapp
    $(basename "$0") backup -n myapp -d /path/to/backups
    $(basename "$0") backup -n myapp --exclude-cache --exclude-post-revisions
    $(basename "$0") restore -n myapp -s /path/to/backup.wpress
    $(basename "$0") full -n myapp -s /path/to/backup.wpress

EOF
}

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

log_progress() {
    echo -n "[PROGRESS] $*"
}

check_dependencies() {
    local deps=("podman")

    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Required dependency not found: $dep"
            exit 1
        fi
    done
}

# Global to store original WP_DEBUG state
ORIGINAL_WP_DEBUG=""

get_wp_debug_state() {
    local site_name="$1"
    ORIGINAL_WP_DEBUG=$(podman exec -it php zsh -c "
        cd $site_name && \
        wp config get WP_DEBUG --allow-root 2>/dev/null || echo '0'
    " | tr -d '\r\n')
    log_info "Captured WP_DEBUG state: $ORIGINAL_WP_DEBUG"
}

restore_wp_debug_state() {
    local site_name="$1"
    if [[ -n "$ORIGINAL_WP_DEBUG" ]]; then
        log_info "Restoring WP_DEBUG to: $ORIGINAL_WP_DEBUG"
        podman exec -it php zsh -c "
            cd $site_name && \
            wp config set WP_DEBUG $ORIGINAL_WP_DEBUG --raw --allow-root
        "
    fi
}

ensure_plugin_available() {
    local site_name="$1"

    if ! podman exec -it php zsh -c "cd $site_name && wp plugin is-installed all-in-one-wp-migration --allow-root" 2>/dev/null; then
        log_info "Installing All-in-One WP Migration plugin"

        local plugins_dir="wp-content/plugins"
        local plugin_repo="https://${GITHUB_TOKEN}@github.com/${GITHUB_USER}/all-in-one-wp-migration.git"

        podman exec -it php zsh -c "
            cd $site_name && \
            git clone '$plugin_repo' '$plugins_dir/all-in-one-wp-migration'
        "

        if [ $? -ne 0 ]; then
            log_error "Failed to install All-in-One WP Migration plugin"
            exit 1
        fi

        log_success "Plugin installed successfully"
    fi
}

ensure_plugin_activated() {
    local site_name="$1"

    if ! podman exec -it php zsh -c "cd $site_name && wp plugin is-active all-in-one-wp-migration --allow-root" 2>/dev/null; then
        log_info "Activating All-in-One WP Migration plugin"
        podman exec -it php zsh -c "
            cd $site_name && \
            wp plugin activate all-in-one-wp-migration --allow-root
        " 2>&1 | grep -v "Warning: Undefined" || true

        if [ $? -ne 0 ]; then
            log_error "Failed to activate plugin"
            exit 1
        fi
        log_success "Plugin activated successfully"
    fi
}

deactivate_aiowm_plugin() {
    local site_name="$1"
    log_info "Deactivating All-in-One WP Migration plugin"
    podman exec -it php zsh -c "
        cd $site_name && \
        wp plugin deactivate all-in-one-wp-migration --allow-root
    " 2>&1 | grep -v "Warning: Undefined" || true
}

cleanup_aiowm() {
    local site_name="$1"
    log_info "Running AIOWM cleanup for: $site_name"
    deactivate_aiowm_plugin "$site_name"
    restore_wp_debug_state "$site_name"
    log_success "AIOWM cleanup completed"
}

ensure_debug_off() {
    local site_name="$1"

    log_info "Setting WP_DEBUG to false"
    podman exec -it php zsh -c "
        cd $site_name && \
        wp config set WP_DEBUG false --raw --allow-root
    "
}

ensure_backup_permissions() {
    local site_name="$1"
    local backup_dir="wp-content/ai1wm-backups"

    log_info "Ensuring backup directory permissions"
    podman exec -it php zsh -c "
        cd $site_name && \
        mkdir -p '$backup_dir' && \
        chmod -R 777 '$backup_dir'
    "
}

ensure_backup_ready() {
    local site_name="$1"

    get_wp_debug_state "$site_name"
    ensure_plugin_available "$site_name"
    ensure_plugin_activated "$site_name"
    ensure_debug_off "$site_name"
    ensure_backup_permissions "$site_name"
}

remove_previous_site() {
    local site_name="myapp"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    log_info "Removing existing site if present: $site_name"
    
    podman exec -it php zsh -c "
        if [ -d $site_name ]; then
            cd $site_name && \
            wp db drop --yes --allow-root 2>/dev/null || true && \
            cd .. && rm -rf $site_name
        fi
    "
    
    log_info "Site cleanup completed."
}

install_wordpress() {
    local site_name="myapp"
    local profile="bare"
    local force_flag=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            -p|--profile)
                profile="$2"
                shift 2
                ;;
            -f|--force)
                force_flag="--force"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    log_info "Installing WordPress site: $site_name"
    "$SCRIPTS_DIR/wordpress/install-wordpress.sh" "$site_name" -p "$profile" $force_flag
}

prepare_migration() {
    local site_name="myapp"
    local source_file=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            -s|--source)
                source_file="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    local backup_dir="$PROJECT_ROOT/apps/$site_name/wp-content/ai1wm-backups"
    
    log_info "Creating backup directory: $backup_dir"
    mkdir -p "$backup_dir"
    
    log_info "Setting permissions on backup directory"

    podman exec -it php zsh -c "
        cd $site_name && \
        chmod -R 777 "$backup_dir"
    "

    
    if [[ -n "$source_file" ]]; then
        if [[ ! -f "$source_file" ]]; then
            log_error "Source file not found: $source_file"
            exit 1
        fi
        
        log_info "Copying backup file to: $backup_dir"
        cp -r "$source_file" "$backup_dir/"
    else
        log_info "No source file specified. Directory prepared for manual file placement."
    fi
}

activate_plugin() {
    local site_name="myapp"

    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    ensure_plugin_available "$site_name"

    log_info "Configuring WordPress and activating plugin in container"

    podman exec -it php zsh -c "
        cd $site_name && \
        wp config set WP_DEBUG false --raw --allow-root && \
        wp plugin activate all-in-one-wp-migration --allow-root
    "

    log_info "Plugin activated. Use WordPress Admin UI to restore the migration."
    log_info "After restore, run: $(basename "$0") deactivate -n $site_name"
}

deactivate_and_cleanup() {
    local site_name="myapp"

    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    get_wp_debug_state "$site_name"
    cleanup_aiowm "$site_name"
}

backup_site() {
    local site_name="myapp"
    local dest_dir=""
    local backup_flags=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            -d|--dest)
                dest_dir="$2"
                shift 2
                ;;
            --exclude-spam-comments)
                backup_flags="$backup_flags --exclude-spam-comments"
                shift
                ;;
            --exclude-post-revisions)
                backup_flags="$backup_flags --exclude-post-revisions"
                shift
                ;;
            --exclude-media)
                backup_flags="$backup_flags --exclude-media"
                shift
                ;;
            --exclude-themes)
                backup_flags="$backup_flags --exclude-themes"
                shift
                ;;
            --exclude-inactive-themes)
                backup_flags="$backup_flags --exclude-inactive-themes"
                shift
                ;;
            --exclude-muplugins)
                backup_flags="$backup_flags --exclude-muplugins"
                shift
                ;;
            --exclude-plugins)
                backup_flags="$backup_flags --exclude-plugins"
                shift
                ;;
            --exclude-inactive-plugins)
                backup_flags="$backup_flags --exclude-inactive-plugins"
                shift
                ;;
            --exclude-cache)
                backup_flags="$backup_flags --exclude-cache"
                shift
                ;;
            --exclude-database)
                backup_flags="$backup_flags --exclude-database"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    ensure_backup_ready "$site_name"

    log_info "Creating backup for site: $site_name"
    echo "Backup in progress..."

    podman exec -it php zsh -c "
        cd $site_name && \
        wp ai1wm backup $backup_flags --allow-root
    " 2>&1 | while IFS= read -r line; do
        if echo "$line" | grep -qE "(Backup (in progress|complete)|Archiving|Preparing|Finalizing)"; then
            echo "$line"
        fi
    done

    local source_backup_dir="$PROJECT_ROOT/apps/$site_name/wp-content/ai1wm-backups"
    local latest_backup=$(ls -t "$source_backup_dir"/*.wpress 2>/dev/null | head -n 1)

    if [[ -z "$latest_backup" ]]; then
        log_error "Backup command completed but no .wpress file found"
        exit 1
    fi

    local backup_size=$(du -h "$latest_backup" | cut -f1)
    log_success "Backup complete: $(basename "$latest_backup") ($backup_size)"

    # Cleanup: deactivate plugin and restore WP_DEBUG
    cleanup_aiowm "$site_name"

    if [[ -n "$dest_dir" ]]; then
        log_info "Moving backup to destination: $dest_dir"
        mkdir -p "$dest_dir"
        cp "$latest_backup" "$dest_dir/"
        log_info "Backup copied to: $dest_dir/$(basename "$latest_backup")"
    else
        log_info "Backup saved to: $PROJECT_ROOT/apps/$site_name/wp-content/ai1wm-backups"
    fi
}

restore_site() {
    local site_name="myapp"
    local source_file=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            -s|--source)
                source_file="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$source_file" ]]; then
        log_error "Source backup file required. Use -s or --source"
        exit 1
    fi
    
    if [[ ! -f "$source_file" ]]; then
        log_error "Source file not found: $source_file"
        exit 1
    fi

    ensure_backup_ready "$site_name"

    local backup_filename=$(basename "$source_file")

    log_info "Validating backup file: $backup_filename"
    local file_size=$(stat -c%s "$source_file" 2>/dev/null || stat -f%z "$source_file" 2>/dev/null)
    if [ "$file_size" -lt 1024 ]; then
        log_error "Backup file appears corrupted (size: $file_size bytes)"
        exit 1
    fi

    local backup_dir="$PROJECT_ROOT/apps/$site_name/wp-content/ai1wm-backups"
    
    log_info "Preparing restore for site: $site_name"
    
    mkdir -p "$backup_dir"
    cp "$source_file" "$backup_dir/"
    
    log_info "Restoring backup: $backup_filename"
    
    podman exec -it php zsh -c "
        cd $site_name && \
        wp ai1wm restore $backup_filename --allow-root
    "

    # Cleanup: deactivate plugin and restore WP_DEBUG
    cleanup_aiowm "$site_name"

    log_info "Restore completed successfully"
}

run_full_workflow() {
    local site_name="myapp"
    local profile="bare"
    local source_file=""
    local force_flag=""
    local backup_dest=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -n|--name)
                site_name="$2"
                shift 2
                ;;
            -p|--profile)
                profile="$2"
                shift 2
                ;;
            -s|--source)
                source_file="$2"
                shift 2
                ;;
            -f|--force)
                force_flag="-f"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    log_info "Running full workflow for site: $site_name"
    
    local site_exists=$(podman exec php zsh -c "[ -d $site_name ] && echo 'yes' || echo 'no'")
    
    if [[ "$site_exists" == "yes" ]]; then
        if [[ -n "$source_file" ]]; then
            backup_dest="$(dirname "$source_file")"
        else
            backup_dest="$PROJECT_ROOT/backups/safety"
        fi
        
        log_info "Creating safety backup before removal to: $backup_dest"
        backup_site -n "$site_name" -d "$backup_dest" || log_error "Backup failed, but continuing..."
        
        if [[ -n "$source_file" ]]; then
            local source_backup_dir="$PROJECT_ROOT/apps/$site_name/wp-content/ai1wm-backups"
            local latest_backup=$(ls -t "$source_backup_dir"/*.wpress 2>/dev/null | head -n 1)
            
            if [[ -n "$latest_backup" ]]; then
                local timestamp=$(date +%Y%m%d_%H%M%S)
                local new_name="${site_name}_pre-restore_${timestamp}.wpress"
                mv "$backup_dest/$(basename "$latest_backup")" "$backup_dest/$new_name"
                log_info "Safety backup saved as: $backup_dest/$new_name"
            fi
        fi
    fi
    
    remove_previous_site -n "$site_name"
    install_wordpress -n "$site_name" -p "$profile" $force_flag
    activate_plugin -n "$site_name"
    
    if [[ -n "$source_file" ]]; then
        log_info "Restoring from source file"
        restore_site -n "$site_name" -s "$source_file"
    else
        log_info "No source file provided. Site ready for manual restore."
    fi
}

main() {
    if [[ $# -eq 0 ]]; then
        print_usage
        exit 0
    fi
    
    check_dependencies
    
    local command="$1"
    shift
    
    case "$command" in
        install)
            install_wordpress "$@"
            ;;
        prepare)
            prepare_migration "$@"
            ;;
        activate)
            activate_plugin "$@"
            ;;
        deactivate)
            deactivate_and_cleanup "$@"
            ;;
        remove)
            remove_previous_site "$@"
            ;;
        backup)
            backup_site "$@"
            ;;
        restore)
            restore_site "$@"
            ;;
        full)
            run_full_workflow "$@"
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            log_error "Unknown command: $command"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"