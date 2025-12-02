#!/bin/bash

# =============================================================================
# ENHANCED HOSTS FILE UPDATER
# =============================================================================
# Intelligently updates /etc/hosts with DevArch application entries.
# Auto-detects all apps in apps/ directory and manages .test domains.
#
# Usage: ./update-hosts.sh [action] [app] [options]
# =============================================================================

# Configuration
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
APPS_DIR="${APPS_DIR:-$PROJECT_ROOT/apps}"
DETECT_SCRIPT="${SCRIPT_DIR}/detect-app-runtime.sh"

# Hosts file configuration
HOSTS_FILE="/etc/hosts"
SECTION_START="# DevArch Apps - Start"
SECTION_END="# DevArch Apps - End"

# Default options
ACTION="update"
SPECIFIC_APP=""
DRY_RUN=false
VERBOSE=false
BACKUP=true

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

print_status() {
    local level="$1"
    local message="$2"

    case "$level" in
        info)    echo -e "\e[34m[INFO]\e[0m $message" ;;
        success) echo -e "\e[32m[SUCCESS]\e[0m $message" ;;
        warning) echo -e "\e[33m[WARNING]\e[0m $message" ;;
        error)   echo -e "\e[31m[ERROR]\e[0m $message" ;;
        step)    echo -e "\e[36m[STEP]\e[0m $message" ;;
        *)       echo "$message" ;;
    esac
}

print_usage() {
    cat << EOF
Usage: $(basename "$0") [action] [app] [options]

Manages DevArch application entries in /etc/hosts file.

Actions:
    update              Update all app entries (default)
    add APP             Add specific app entry
    remove APP          Remove specific app entry
    scan                Scan and report without modifying
    list                List all DevArch entries in hosts file

Options:
    -n, --dry-run       Show what would be changed without modifying
    -v, --verbose       Verbose output
    --no-backup         Skip backup creation
    -h, --help          Show this help message

Examples:
    sudo $(basename "$0")                   # Update all app entries
    sudo $(basename "$0") add myapp         # Add myapp.test entry
    sudo $(basename "$0") remove myapp      # Remove myapp.test entry
    sudo $(basename "$0") scan              # Scan without modifying
    $(basename "$0") list                   # List current entries (no sudo needed)

Notes:
    - Requires sudo for modifying /etc/hosts
    - Auto-detects all apps in apps/ directory
    - Creates backup before modification
    - Uses markers to manage DevArch section
EOF
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

check_sudo() {
    if [[ $EUID -ne 0 ]]; then
        print_status "error" "This action requires sudo privileges"
        print_status "info" "Run: sudo $(basename "$0") $*"
        exit 1
    fi
}

validate_hosts_file() {
    if [[ ! -f "$HOSTS_FILE" ]]; then
        print_status "error" "Hosts file not found: $HOSTS_FILE"
        exit 1
    fi

    if [[ ! -r "$HOSTS_FILE" ]]; then
        print_status "error" "Cannot read hosts file: $HOSTS_FILE"
        exit 1
    fi
}

validate_app_exists() {
    local app_name="$1"

    if [[ ! -d "${APPS_DIR}/${app_name}" ]]; then
        print_status "error" "App directory not found: ${APPS_DIR}/${app_name}"
        return 1
    fi

    return 0
}

# =============================================================================
# APP DISCOVERY
# =============================================================================

discover_all_apps() {
    local -a apps

    # Check if apps directory exists
    if [[ ! -d "$APPS_DIR" ]]; then
        print_status "warning" "Apps directory not found: $APPS_DIR"
        return 1
    fi

    # Scan all directories in apps/
    for app_dir in "$APPS_DIR"/*; do
        # Skip if not a directory
        [[ ! -d "$app_dir" ]] && continue

        # Skip hidden directories
        local app_name=$(basename "$app_dir")
        [[ "$app_name" =~ ^\. ]] && continue

        # Detect runtime to verify it's a valid app
        if [[ -x "$DETECT_SCRIPT" ]]; then
            local runtime=$("$DETECT_SCRIPT" "$app_name" 2>/dev/null) || runtime="unknown"

            # Only include apps with detected runtime
            if [[ "$runtime" != "unknown" ]]; then
                apps+=("$app_name")
            fi
        else
            # If detection script not available, include all directories
            apps+=("$app_name")
        fi
    done

    # Return apps array
    printf '%s\n' "${apps[@]}" | sort
}

# =============================================================================
# HOSTS FILE MANIPULATION
# =============================================================================

create_backup() {
    if [[ "$BACKUP" != "true" ]]; then
        return 0
    fi

    local backup_file="${HOSTS_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

    print_status "step" "Creating backup: $backup_file"

    if cp "$HOSTS_FILE" "$backup_file"; then
        print_status "success" "Backup created"
        echo "$backup_file"
    else
        print_status "warning" "Failed to create backup, continuing anyway..."
        echo ""
    fi
}

find_devarch_section() {
    local start_line=$(grep -n "^${SECTION_START}" "$HOSTS_FILE" 2>/dev/null | head -1 | cut -d: -f1)
    local end_line=$(grep -n "^${SECTION_END}" "$HOSTS_FILE" 2>/dev/null | head -1 | cut -d: -f1)

    if [[ -n "$start_line" && -n "$end_line" ]]; then
        echo "$start_line $end_line"
        return 0
    else
        echo "0 0"
        return 1
    fi
}

extract_devarch_entries() {
    local section_info=$(find_devarch_section)
    local start_line=$(echo "$section_info" | cut -d' ' -f1)
    local end_line=$(echo "$section_info" | cut -d' ' -f2)

    if [[ "$start_line" -eq 0 ]]; then
        # No existing section
        return 1
    fi

    # Extract entries between markers
    sed -n "${start_line},${end_line}p" "$HOSTS_FILE" | grep -v "^#" | grep "\.test$" | awk '{print $2}' | sed 's/\.test$//'
}

build_devarch_section() {
    local -a apps
    mapfile -t apps

    if [[ ${#apps[@]} -eq 0 ]]; then
        print_status "warning" "No apps to add to hosts file"
        return 1
    fi

    # Build section content
    echo "$SECTION_START"
    echo "# Auto-managed by DevArch - Do not edit manually"
    echo "#"

    for app_name in "${apps[@]}"; do
        echo "127.0.0.1    ${app_name}.test"
    done

    echo "#"
    echo "# Total: ${#apps[@]} application(s)"
    echo "$SECTION_END"
}

update_hosts_file() {
    local -a apps
    mapfile -t apps

    if [[ ${#apps[@]} -eq 0 ]]; then
        print_status "warning" "No apps found, removing DevArch section if it exists"
    fi

    # Create temporary files
    local temp_file=$(mktemp)
    local new_section=$(mktemp)

    # Build new section
    printf '%s\n' "${apps[@]}" | build_devarch_section > "$new_section"

    # Get section info
    local section_info=$(find_devarch_section)
    local start_line=$(echo "$section_info" | cut -d' ' -f1)
    local end_line=$(echo "$section_info" | cut -d' ' -f2)

    if [[ "$start_line" -eq 0 ]]; then
        # No existing section - append to end
        cat "$HOSTS_FILE" > "$temp_file"
        echo "" >> "$temp_file"
        cat "$new_section" >> "$temp_file"
    else
        # Replace existing section
        head -n $((start_line - 1)) "$HOSTS_FILE" > "$temp_file"
        cat "$new_section" >> "$temp_file"
        tail -n +$((end_line + 1)) "$HOSTS_FILE" >> "$temp_file"
    fi

    # Show diff if verbose or dry run
    if [[ "$VERBOSE" == "true" || "$DRY_RUN" == "true" ]]; then
        echo ""
        print_status "info" "Changes to be made:"
        echo ""
        diff -u "$HOSTS_FILE" "$temp_file" || true
        echo ""
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "info" "[DRY RUN] Would update $HOSTS_FILE"
        rm -f "$temp_file" "$new_section"
        return 0
    fi

    # Apply changes
    if cp "$temp_file" "$HOSTS_FILE"; then
        print_status "success" "Hosts file updated successfully"
        rm -f "$temp_file" "$new_section"
        return 0
    else
        print_status "error" "Failed to update hosts file"
        rm -f "$temp_file" "$new_section"
        return 1
    fi
}

add_app_entry() {
    local app_name="$1"

    print_status "step" "Adding entry for: ${app_name}.test"

    # Validate app exists
    if ! validate_app_exists "$app_name"; then
        return 1
    fi

    # Get current apps
    local current_apps=$(extract_devarch_entries 2>/dev/null || echo "")
    local all_apps=$(discover_all_apps)

    # Check if app is already in list
    if echo "$current_apps" | grep -q "^${app_name}$"; then
        print_status "warning" "Entry for ${app_name}.test already exists"
        return 0
    fi

    # Combine and update
    {
        echo "$current_apps"
        echo "$all_apps"
    } | sort -u | grep -v "^$" | update_hosts_file
}

remove_app_entry() {
    local app_name="$1"

    print_status "step" "Removing entry for: ${app_name}.test"

    # Get current apps
    local current_apps=$(extract_devarch_entries 2>/dev/null || echo "")

    # Check if app exists in list
    if ! echo "$current_apps" | grep -q "^${app_name}$"; then
        print_status "warning" "Entry for ${app_name}.test not found in hosts file"
        return 0
    fi

    # Remove app and update
    echo "$current_apps" | grep -v "^${app_name}$" | update_hosts_file
}

scan_and_report() {
    print_status "info" "Scanning apps directory and hosts file..."
    echo ""

    # Discovered apps
    local discovered=$(discover_all_apps)
    local discovered_count=$(echo "$discovered" | grep -v "^$" | wc -l)

    echo "Apps in ${APPS_DIR}:"
    echo "--------------------"
    if [[ -n "$discovered" ]]; then
        echo "$discovered" | while read app; do
            echo "  - $app"
        done
    else
        echo "  (none)"
    fi
    echo ""
    echo "Total discovered: $discovered_count"
    echo ""

    # Current hosts entries
    local current=$(extract_devarch_entries 2>/dev/null || echo "")
    local current_count=$(echo "$current" | grep -v "^$" | wc -l)

    echo "Current DevArch entries in $HOSTS_FILE:"
    echo "----------------------------------------"
    if [[ -n "$current" ]]; then
        echo "$current" | while read app; do
            echo "  - ${app}.test"
        done
    else
        echo "  (none)"
    fi
    echo ""
    echo "Total in hosts: $current_count"
    echo ""

    # Comparison
    echo "Analysis:"
    echo "---------"

    # Apps not in hosts
    local missing=$(comm -23 <(echo "$discovered" | sort) <(echo "$current" | sort))
    if [[ -n "$missing" ]]; then
        echo "Apps not in hosts file:"
        echo "$missing" | while read app; do
            echo "  - $app"
        done
        echo ""
    fi

    # Hosts entries for non-existent apps
    local orphaned=$(comm -13 <(echo "$discovered" | sort) <(echo "$current" | sort))
    if [[ -n "$orphaned" ]]; then
        echo "Orphaned hosts entries (app doesn't exist):"
        echo "$orphaned" | while read app; do
            echo "  - $app"
        done
        echo ""
    fi

    if [[ -z "$missing" && -z "$orphaned" ]]; then
        print_status "success" "Hosts file is in sync with apps directory"
    else
        print_status "warning" "Hosts file needs updating"
        print_status "info" "Run: sudo $(basename "$0") update"
    fi
}

list_devarch_entries() {
    echo ""
    echo "DevArch Hosts Entries"
    echo "====================="
    echo ""

    local section_info=$(find_devarch_section)
    local start_line=$(echo "$section_info" | cut -d' ' -f1)

    if [[ "$start_line" -eq 0 ]]; then
        print_status "warning" "No DevArch section found in $HOSTS_FILE"
        echo ""
        return 1
    fi

    local entries=$(extract_devarch_entries)
    local count=$(echo "$entries" | grep -v "^$" | wc -l)

    if [[ -n "$entries" ]]; then
        echo "$entries" | while read app; do
            echo "  127.0.0.1    ${app}.test"
        done
    else
        echo "  (none)"
    fi

    echo ""
    echo "Total: $count entry(ies)"
    echo ""
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            update)
                ACTION="update"
                shift
                ;;
            add)
                ACTION="add"
                if [[ -n "$2" && "$2" != -* ]]; then
                    SPECIFIC_APP="$2"
                    shift 2
                else
                    print_status "error" "Action 'add' requires an app name"
                    exit 1
                fi
                ;;
            remove)
                ACTION="remove"
                if [[ -n "$2" && "$2" != -* ]]; then
                    SPECIFIC_APP="$2"
                    shift 2
                else
                    print_status "error" "Action 'remove' requires an app name"
                    exit 1
                fi
                ;;
            scan)
                ACTION="scan"
                shift
                ;;
            list)
                ACTION="list"
                shift
                ;;
            -n|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            --no-backup)
                BACKUP=false
                shift
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            *)
                print_status "error" "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse arguments
    parse_arguments "$@"

    # Validate hosts file
    validate_hosts_file

    # Check sudo for modifying actions
    if [[ "$ACTION" != "scan" && "$ACTION" != "list" && "$DRY_RUN" != "true" ]]; then
        check_sudo
    fi

    # Execute action
    case "$ACTION" in
        update)
            print_status "step" "Updating DevArch hosts entries..."
            echo ""

            # Create backup
            if [[ "$DRY_RUN" != "true" ]]; then
                create_backup
                echo ""
            fi

            # Discover and update
            local apps=$(discover_all_apps)
            local count=$(echo "$apps" | grep -v "^$" | wc -l)

            if [[ "$VERBOSE" == "true" ]]; then
                print_status "info" "Discovered $count app(s)"
                echo "$apps" | while read app; do
                    echo "  - $app"
                done
                echo ""
            fi

            echo "$apps" | update_hosts_file

            if [[ $? -eq 0 && "$DRY_RUN" != "true" ]]; then
                print_status "success" "Updated $count app entries"
            fi
            ;;

        add)
            if [[ "$DRY_RUN" != "true" ]]; then
                create_backup
                echo ""
            fi

            add_app_entry "$SPECIFIC_APP"
            ;;

        remove)
            if [[ "$DRY_RUN" != "true" ]]; then
                create_backup
                echo ""
            fi

            remove_app_entry "$SPECIFIC_APP"
            ;;

        scan)
            scan_and_report
            ;;

        list)
            list_devarch_entries
            ;;
    esac
}

# Only run main if script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
