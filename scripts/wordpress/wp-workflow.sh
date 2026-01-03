#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${PROJECT_ROOT:-$HOME/projects/devarch}"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"

print_usage() {
    cat << EOF
WordPress Multi System Workflow Manager

Usage: $(basename "$0") [COMMAND] [OPTIONS]

Commands:
    install         Install WordPress site
    prepare         Prepare migration directories and files
    activate        Activate All In One WP Migration plugin
    remove          Remove existing WordPress site and database
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

Remove Options:
    -n, --name      Site name (default: myapp)

Examples:
    $(basename "$0") install -n myapp -p bare -f
    $(basename "$0") prepare -n b2bcnc -s /path/to/backup.wpress
    $(basename "$0") activate -n myapp
    $(basename "$0") remove -n myapp
    $(basename "$0") full -n myapp -s /path/to/backup.wpress

EOF
}

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
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

start_services() {
    log_info "Starting services: mariadb, nginx-proxy-manager, php"
    "$SCRIPTS_DIR/service-manager.sh" start mariadb 
    "$SCRIPTS_DIR/service-manager.sh" start nginx-proxy-manager 
    "$SCRIPTS_DIR/service-manager.sh" start php
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
    start_services
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
    sudo chmod -R 777 "$backup_dir"
    
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
    
    log_info "Configuring WordPress and activating plugin in container"
    
    podman exec -it php zsh -c "
        cd $site_name && \
        wp config set WP_DEBUG false --raw --allow-root && \
        wp plugin activate all-in-one-wp-migration --allow-root
    "
    
    log_info "Plugin activated. Use WordPress Admin UI to restore the migration."
}

run_full_workflow() {
    local site_name="myapp"
    local profile="bare"
    local source_file=""
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
    
    remove_previous_site -n "$site_name"
    install_wordpress -n "$site_name" -p "$profile" $force_flag
    
    if [[ -n "$source_file" ]]; then
        prepare_migration -n "$site_name" -s "$source_file"
    else
        prepare_migration -n "$site_name"
    fi
    
    activate_plugin -n "$site_name"
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
        remove)
            remove_previous_site "$@"
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