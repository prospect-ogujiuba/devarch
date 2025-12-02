#!/bin/zsh

# =============================================================================
# HOSTS FILE UPDATER - CROSS-PLATFORM
# =============================================================================
# Intelligently updates system hosts file with microservices entries
# Integrates with generate-context.sh and preserves existing hosts entries

# Source the central configuration if available
SCRIPT_DIR=$(dirname "$0")
if [[ -f "$SCRIPT_DIR/config.sh" ]]; then
    . "$SCRIPT_DIR/config.sh"
else
    # Source environment variables
    set -a
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        source "$PROJECT_ROOT/.env"
    fi
    set +a
    
    # Set context directory if not already set
    CONTEXT_DIR="${CONTEXT_DIR:-$PROJECT_ROOT/context}"
fi

# =============================================================================
# SCRIPT OPTIONS & DEFAULTS
# =============================================================================

# Default values
opt_os="auto"
opt_backup=true
opt_dry_run=false
opt_verbose=false
opt_force=false
opt_restore=false
opt_check_only=false

# Cross-platform hosts file paths
typeset -A HOSTS_PATHS
HOSTS_PATHS=(
    [linux]="/etc/hosts"
    [macos]="/etc/hosts"
    [windows]=""  # Will be detected dynamically
    [wsl]="/etc/hosts"
)

# Microservices section markers
SECTION_START="# Microservices Host File Entries"
SECTION_END_PATTERN="# Summary: [0-9]* domains found"

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

print_status() {
    local level="$1"
    local message="$2"
    
    case "$level" in
        "info")
            echo "ℹ️  $message"
            ;;
        "success")
            echo "✅ $message"
            ;;
        "warning")
            echo "⚠️  $message"
            ;;
        "error")
            echo "❌ $message"
            ;;
        "step")
            echo "🔄 $message"
            ;;
        *)
            echo "$message"
            ;;
    esac
}

handle_error() {
    print_status "error" "$1"
    exit 1
}

# =============================================================================
# OS DETECTION AND VALIDATION
# =============================================================================

detect_os() {
    local detected_os="unknown"
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Check if we're in WSL
        if grep -qEi "(Microsoft|WSL)" /proc/version 2>/dev/null; then
            detected_os="wsl"
        else
            detected_os="linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        detected_os="macos"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        detected_os="windows"
    elif [[ -n "$WSL_DISTRO_NAME" ]]; then
        detected_os="wsl"
    fi
    
    echo "$detected_os"
}

get_hosts_file_path() {
    local target_os="$1"
    
    if [[ "$target_os" == "auto" ]]; then
        target_os=$(detect_os)
    fi
    
    # Handle Windows paths dynamically (especially for WSL)
    if [[ "$target_os" == "windows" ]]; then
        local windows_hosts_path
        windows_hosts_path=$(find_windows_hosts_path)
        if [[ -n "$windows_hosts_path" ]]; then
            echo "$windows_hosts_path"
            return 0
        else
            handle_error "Could not locate Windows hosts file from WSL/current environment"
        fi
    fi
    
    local hosts_path="${HOSTS_PATHS[$target_os]}"
    
    if [[ -z "$hosts_path" ]]; then
        handle_error "Unsupported operating system: $target_os"
    fi
    
    echo "$hosts_path"
}

find_windows_hosts_path() {
    # Try common WSL mount points for Windows hosts file
    local -a possible_paths=(
        "/mnt/c/Windows/System32/drivers/etc/hosts"
        "/c/Windows/System32/drivers/etc/hosts"
        "/drive_c/Windows/System32/drivers/etc/hosts"
    )
    
    # Check if we're in WSL and can access Windows filesystem
    if [[ -n "$WSL_DISTRO_NAME" ]] || grep -qEi "(Microsoft|WSL)" /proc/version 2>/dev/null; then
        # Suppress output during path detection to avoid contaminating return value
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "info" "Detected WSL environment, searching for Windows hosts file..." >&2
        fi
        
        # Try to find the Windows mount point dynamically
        local windows_drive
        windows_drive=$(mount | grep -i "drvfs.*windows" | head -1 | awk '{print $3}')
        
        if [[ -n "$windows_drive" ]]; then
            local dynamic_path="$windows_drive/System32/drivers/etc/hosts"
            if [[ -f "$dynamic_path" ]]; then
                if [[ "$opt_verbose" == "true" ]]; then
                    print_status "success" "Found Windows hosts file via mount: $dynamic_path" >&2
                fi
                echo "$dynamic_path"
                return 0
            fi
        fi
        
        # Try standard mount points
        for path in "${possible_paths[@]}"; do
            if [[ -f "$path" ]]; then
                if [[ "$opt_verbose" == "true" ]]; then
                    print_status "success" "Found Windows hosts file: $path" >&2
                fi
                echo "$path"
                return 0
            fi
        done
        
        # Try to find any C: drive mount
        local c_mount
        c_mount=$(mount | grep -E "(type drvfs|type 9p)" | grep -i "/c\b" | head -1 | awk '{print $3}')
        if [[ -z "$c_mount" ]]; then
            c_mount=$(mount | grep -E "(type drvfs|type 9p)" | head -1 | awk '{print $3}')
        fi
        
        if [[ -n "$c_mount" ]]; then
            local fallback_path="$c_mount/Windows/System32/drivers/etc/hosts"
            if [[ -f "$fallback_path" ]]; then
                if [[ "$opt_verbose" == "true" ]]; then
                    print_status "success" "Found Windows hosts file via C: mount: $fallback_path" >&2
                fi
                echo "$fallback_path"
                return 0
            fi
        fi
        
        if [[ "$opt_verbose" == "true" ]]; then
            print_status "warning" "Could not locate Windows hosts file in WSL" >&2
            print_status "info" "Tried paths: ${possible_paths[*]}" >&2
            print_status "info" "Current mounts:" >&2
            mount | grep -E "(drvfs|9p)" | head -5 >&2
        fi
    else
        # Not in WSL, try Git Bash or other Windows environments
        for path in "${possible_paths[@]}"; do
            if [[ -f "$path" ]]; then
                echo "$path"
                return 0
            fi
        done
    fi
    
    return 1
}

validate_hosts_file() {
    local hosts_file="$1"
    
    if [[ ! -f "$hosts_file" ]]; then
        handle_error "Hosts file not found: $hosts_file"
    fi
    
    if [[ ! -r "$hosts_file" ]]; then
        handle_error "Cannot read hosts file: $hosts_file (permission denied)"
    fi
    
    # Check if we need sudo for writing
    if [[ ! -w "$hosts_file" ]]; then
        if [[ "$opt_dry_run" == "false" && "$opt_check_only" == "false" ]]; then
            print_status "info" "Hosts file requires elevated privileges for writing"
            return 1
        fi
    fi
    
    return 0
}

# =============================================================================
# MICROSERVICES ENTRIES MANAGEMENT
# =============================================================================

find_microservices_section() {
    local hosts_file="$1"
    local -a section_info
    
    # Find start line
    local start_line=$(grep -n "^$SECTION_START" "$hosts_file" | head -1 | cut -d: -f1)
    
    if [[ -z "$start_line" ]]; then
        # No existing section found
        echo "0 0"
        return 1
    fi
    
    # Find end line (looking for summary line)
    local end_line=$(tail -n +$start_line "$hosts_file" | grep -n "$SECTION_END_PATTERN" | head -1 | cut -d: -f1)
    
    if [[ -n "$end_line" ]]; then
        # Calculate actual line number in file
        end_line=$((start_line + end_line - 1))
    else
        # Fallback: look for next non-microservices comment or end of file
        local next_section=$(tail -n +$((start_line + 1)) "$hosts_file" | grep -n "^# [^M]" | head -1 | cut -d: -f1)
        if [[ -n "$next_section" ]]; then
            end_line=$((start_line + next_section - 2))
        else
            # If no clear end, include all remaining microservices lines
            end_line=$(wc -l < "$hosts_file")
        fi
    fi
    
    echo "$start_line $end_line"
    return 0
}

extract_existing_entries() {
    local hosts_file="$1"
    local temp_file="$2"
    local section_info=$(find_microservices_section "$hosts_file")
    local start_line=$(echo "$section_info" | cut -d' ' -f1)
    local end_line=$(echo "$section_info" | cut -d' ' -f2)
    
    if [[ "$start_line" -eq 0 ]]; then
        # No existing section, copy entire file
        cp "$hosts_file" "$temp_file"
        echo "$temp_file $(wc -l < "$hosts_file")"
        return 0
    fi
    
    # Extract everything before microservices section
    if [[ "$start_line" -gt 1 ]]; then
        head -n $((start_line - 1)) "$hosts_file" > "$temp_file"
    else
        > "$temp_file"  # Empty file if section starts at line 1
    fi
    
    # Extract everything after microservices section
    local total_lines=$(wc -l < "$hosts_file")
    if [[ "$end_line" -lt "$total_lines" ]]; then
        tail -n +$((end_line + 1)) "$hosts_file" >> "$temp_file"
    fi
    
    echo "$temp_file $start_line"
    return 0
}

get_new_microservices_entries() {
    local context_hosts_file="$CONTEXT_DIR/hosts.txt"
    
    if [[ ! -f "$context_hosts_file" ]]; then
        handle_error "Microservices hosts file not found: $context_hosts_file"
    fi
    
    # Return the entire content of the generated hosts file
    cat "$context_hosts_file"
}

# =============================================================================
# BACKUP AND RESTORE FUNCTIONS
# =============================================================================

create_backup() {
    local hosts_file="$1"
    local backup_file="${hosts_file}.backup.$(date +%Y%m%d_%H%M%S)"
    
    if [[ "$opt_backup" == "true" ]]; then
        print_status "step" "Creating backup: $backup_file"
        if cp "$hosts_file" "$backup_file"; then
            print_status "success" "Backup created successfully"
            echo "$backup_file"
        else
            print_status "warning" "Failed to create backup, continuing anyway..."
            echo ""
        fi
    else
        echo ""
    fi
}

list_backups() {
    local hosts_file="$1"
    local backup_pattern="${hosts_file}.backup.*"
    
    print_status "info" "Available backups:"
    
    local -a backups
    backups=(${hosts_file}.backup.*(N))
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        print_status "info" "No backups found"
        return 1
    fi
    
    for backup in "${backups[@]}"; do
        local backup_date=$(basename "$backup" | sed 's/.*backup\.//')
        local formatted_date=$(date -d "${backup_date:0:8} ${backup_date:9:2}:${backup_date:11:2}:${backup_date:13:2}" 2>/dev/null || echo "$backup_date")
        print_status "info" "  $backup ($formatted_date)"
    done
    
    return 0
}

restore_backup() {
    local hosts_file="$1"
    local backup_file="$2"
    
    if [[ -z "$backup_file" ]]; then
        # Find most recent backup
        local -a backups
        backups=(${hosts_file}.backup.*(N))
        
        if [[ ${#backups[@]} -eq 0 ]]; then
            handle_error "No backups found to restore"
        fi
        
        # Sort and get most recent
        backup_file="${backups[-1]}"
        print_status "info" "Using most recent backup: $backup_file"
    fi
    
    if [[ ! -f "$backup_file" ]]; then
        handle_error "Backup file not found: $backup_file"
    fi
    
    print_status "step" "Restoring from backup: $backup_file"
    
    if [[ "$opt_dry_run" == "true" ]]; then
        print_status "info" "[DRY RUN] Would restore: $backup_file -> $hosts_file"
        return 0
    fi
    
    if require_sudo_for_write "$hosts_file"; then
        if sudo cp "$backup_file" "$hosts_file"; then
            print_status "success" "Hosts file restored from backup"
        else
            handle_error "Failed to restore backup"
        fi
    else
        if cp "$backup_file" "$hosts_file"; then
            print_status "success" "Hosts file restored from backup"
        else
            handle_error "Failed to restore backup"
        fi
    fi
}

# =============================================================================
# HOSTS FILE UPDATE FUNCTIONS
# =============================================================================

require_sudo_for_write() {
    local hosts_file="$1"
    [[ ! -w "$hosts_file" ]]
}

update_hosts_file() {
    local hosts_file="$1"
    local use_sudo=false
    local is_windows_from_wsl=false
    
    # Check if we're updating Windows hosts from WSL
    if [[ "$hosts_file" == /mnt/* ]] && [[ -n "$WSL_DISTRO_NAME" || $(grep -qEi "(Microsoft|WSL)" /proc/version 2>/dev/null) ]]; then
        is_windows_from_wsl=true
        print_status "info" "Detected Windows hosts file access from WSL - using special handling"
    fi
    
    # Check if we need sudo (except for Windows from WSL)
    if [[ "$is_windows_from_wsl" == "false" ]] && require_sudo_for_write "$hosts_file"; then
        use_sudo=true
        print_status "info" "Elevated privileges required for hosts file modification"
    fi
    
    # Create backup
    local backup_file
    if [[ "$is_windows_from_wsl" == "true" ]]; then
        backup_file=$(create_windows_backup_from_wsl "$hosts_file")
    elif [[ "$use_sudo" == "true" ]]; then
        backup_file=$(sudo bash -c "$(declare -f create_backup); create_backup '$hosts_file'")
    else
        backup_file=$(create_backup "$hosts_file")
    fi
    
    # Create temporary files
    local temp_hosts=$(mktemp)
    local temp_new=$(mktemp)
    
    # Clean up temp files on exit
    trap "rm -f '$temp_hosts' '$temp_new'" EXIT
    
    print_status "step" "Extracting existing hosts entries..."
    local extract_info
    if [[ "$is_windows_from_wsl" == "true" ]]; then
        extract_info=$(extract_existing_entries_windows_wsl "$hosts_file" "$temp_hosts")
    elif [[ "$use_sudo" == "true" ]]; then
        extract_info=$(sudo bash -c "$(declare -f extract_existing_entries find_microservices_section); extract_existing_entries '$hosts_file' '$temp_hosts'")
    else
        extract_info=$(extract_existing_entries "$hosts_file" "$temp_hosts")
    fi
    
    local insertion_line=$(echo "$extract_info" | cut -d' ' -f2)
    
    print_status "step" "Getting new microservices entries..."
    get_new_microservices_entries > "$temp_new"
    
    print_status "step" "Merging hosts file entries..."
    # Create the new hosts file content
    local final_hosts=$(mktemp)
    
    if [[ "$insertion_line" == "$(wc -l < "$temp_hosts")" ]] || [[ -z "$(tail -c1 "$temp_hosts")" ]]; then
        # Append to end of file or file ends with newline
        cat "$temp_hosts" > "$final_hosts"
        [[ -s "$final_hosts" ]] && echo "" >> "$final_hosts"  # Add blank line if file has content
        cat "$temp_new" >> "$final_hosts"
    else
        # Insert in middle of file
        cat "$temp_hosts" > "$final_hosts"
        echo "" >> "$final_hosts"  # Add blank line separator
        cat "$temp_new" >> "$final_hosts"
    fi
    
    # Show changes if verbose or dry run
    if [[ "$opt_verbose" == "true" || "$opt_dry_run" == "true" ]]; then
        show_changes "$hosts_file" "$final_hosts"
    fi
    
    if [[ "$opt_dry_run" == "true" ]]; then
        print_status "info" "[DRY RUN] Would update hosts file with new microservices entries"
        return 0
    fi
    
    # Apply changes
    print_status "step" "Writing updated hosts file..."
    if [[ "$is_windows_from_wsl" == "true" ]]; then
        if write_windows_hosts_from_wsl "$final_hosts" "$hosts_file"; then
            print_status "success" "Windows hosts file updated successfully from WSL"
        else
            handle_error "Failed to update Windows hosts file from WSL"
        fi
    elif [[ "$use_sudo" == "true" ]]; then
        if sudo cp "$final_hosts" "$hosts_file"; then
            print_status "success" "Hosts file updated successfully"
        else
            handle_error "Failed to update hosts file"
        fi
    else
        if cp "$final_hosts" "$hosts_file"; then
            print_status "success" "Hosts file updated successfully"
        else
            handle_error "Failed to update hosts file"
        fi
    fi
    
    # Clean up
    rm -f "$final_hosts"
}

create_windows_backup_from_wsl() {
    local hosts_file="$1"
    local backup_file="${hosts_file}.backup.$(date +%Y%m%d_%H%M%S)"
    
    if [[ "$opt_backup" == "true" ]]; then
        print_status "step" "Creating Windows hosts backup: $(basename "$backup_file")"
        
        # Try to create backup in the same directory (Windows)
        if cp "$hosts_file" "$backup_file" 2>/dev/null; then
            print_status "success" "Windows backup created successfully"
            echo "$backup_file"
        else
            # Fallback: create backup in WSL temp directory
            local wsl_backup="/tmp/windows_hosts_backup_$(date +%Y%m%d_%H%M%S)"
            if cp "$hosts_file" "$wsl_backup" 2>/dev/null; then
                print_status "success" "Backup created in WSL temp: $wsl_backup"
                echo "$wsl_backup"
            else
                print_status "warning" "Failed to create backup, continuing anyway..."
                echo ""
            fi
        fi
    else
        echo ""
    fi
}

extract_existing_entries_windows_wsl() {
    local hosts_file="$1"
    local temp_file="$2"
    
    # For Windows hosts from WSL, we need to be more careful about file access
    # First, try to read the file
    if ! cat "$hosts_file" > "$temp_file" 2>/dev/null; then
        handle_error "Cannot read Windows hosts file: $hosts_file"
    fi
    
    # Now process the copied content
    local section_info=$(find_microservices_section "$temp_file")
    local start_line=$(echo "$section_info" | cut -d' ' -f1)
    local end_line=$(echo "$section_info" | cut -d' ' -f2)
    
    if [[ "$start_line" -eq 0 ]]; then
        # No existing section, file is already copied to temp_file
        echo "$temp_file $(wc -l < "$temp_file")"
        return 0
    fi
    
    # Create a new temp file without the microservices section
    local temp_without_section=$(mktemp)
    
    # Extract everything before microservices section
    if [[ "$start_line" -gt 1 ]]; then
        head -n $((start_line - 1)) "$temp_file" > "$temp_without_section"
    else
        > "$temp_without_section"  # Empty file if section starts at line 1
    fi
    
    # Extract everything after microservices section
    local total_lines=$(wc -l < "$temp_file")
    if [[ "$end_line" -lt "$total_lines" ]]; then
        tail -n +$((end_line + 1)) "$temp_file" >> "$temp_without_section"
    fi
    
    # Replace the temp file with the cleaned version
    mv "$temp_without_section" "$temp_file"
    
    echo "$temp_file $start_line"
    return 0
}

write_windows_hosts_from_wsl() {
    local source_file="$1"
    local hosts_file="$2"
    
    # Method 1: Try PowerShell through WSL
    if command -v powershell.exe >/dev/null 2>&1; then
        print_status "info" "Attempting to update Windows hosts file via PowerShell..."
        
        # Convert WSL path to Windows path for PowerShell
        local windows_source_path
        windows_source_path=$(wslpath -w "$source_file" 2>/dev/null)
        local windows_hosts_path
        windows_hosts_path=$(wslpath -w "$hosts_file" 2>/dev/null)
        
        if [[ -n "$windows_source_path" && -n "$windows_hosts_path" ]]; then
            # Use PowerShell to copy with admin rights
            local ps_command="Start-Process powershell -ArgumentList 'Copy-Item \"$windows_source_path\" \"$windows_hosts_path\" -Force' -Verb RunAs -Wait"
            
            if powershell.exe -Command "$ps_command" 2>/dev/null; then
                return 0
            fi
        fi
    fi
    
    # Method 2: Try direct copy (might work in some WSL configurations)
    print_status "info" "Attempting direct copy to Windows hosts file..."
    if cp "$source_file" "$hosts_file" 2>/dev/null; then
        return 0
    fi
    
    # Method 3: Guide user to manual update
    print_status "warning" "Automatic update failed. Manual steps required:"
    print_status "info" "1. Open PowerShell as Administrator in Windows"
    print_status "info" "2. Run this command:"
    
    local windows_source_path
    windows_source_path=$(wslpath -w "$source_file" 2>/dev/null || echo "$source_file")
    local windows_hosts_path
    windows_hosts_path=$(wslpath -w "$hosts_file" 2>/dev/null || echo "$hosts_file")
    
    echo "   Copy-Item \"$windows_source_path\" \"$windows_hosts_path\" -Force"
    echo ""
    print_status "info" "Or copy the content manually:"
    echo "Source file (in WSL): $source_file"
    echo "Target file (Windows): $hosts_file"
    
    return 1
}

show_changes() {
    local original_file="$1"
    local new_file="$2"
    
    print_status "info" "Changes to be made:"
    echo ""
    
    if command -v diff >/dev/null 2>&1; then
        diff -u "$original_file" "$new_file" | head -20
        local diff_lines=$(diff -u "$original_file" "$new_file" | wc -l)
        if [[ "$diff_lines" -gt 20 ]]; then
            echo "... (showing first 20 lines of $diff_lines total diff lines)"
        fi
    else
        print_status "info" "New microservices section:"
        get_new_microservices_entries | head -10
        local entry_lines=$(get_new_microservices_entries | wc -l)
        if [[ "$entry_lines" -gt 10 ]]; then
            echo "... (showing first 10 lines of $entry_lines total lines)"
        fi
    fi
    
    echo ""
}

check_hosts_status() {
    local hosts_file="$1"
    
    print_status "info" "Hosts File Status Check"
    echo ""
    
    echo "Hosts file: $hosts_file"
    if [[ -f "$hosts_file" ]]; then
        print_status "success" "Hosts file exists"
        echo "  Size: $(wc -c < "$hosts_file") bytes"
        echo "  Lines: $(wc -l < "$hosts_file") lines"
        
        if [[ -r "$hosts_file" ]]; then
            print_status "success" "Hosts file is readable"
        else
            print_status "error" "Hosts file is not readable"
        fi
        
        if [[ -w "$hosts_file" ]]; then
            print_status "success" "Hosts file is writable"
        else
            print_status "warning" "Hosts file requires elevated privileges"
        fi
    else
        print_status "error" "Hosts file does not exist"
        return 1
    fi
    
    echo ""
    print_status "info" "Microservices section status:"
    
    local section_info=$(find_microservices_section "$hosts_file")
    local start_line=$(echo "$section_info" | cut -d' ' -f1)
    local end_line=$(echo "$section_info" | cut -d' ' -f2)
    
    if [[ "$start_line" -eq 0 ]]; then
        print_status "warning" "No microservices section found"
    else
        print_status "success" "Microservices section found (lines $start_line-$end_line)"
        
        # Show current entries count
        local current_entries=$(sed -n "${start_line},${end_line}p" "$hosts_file" | grep "^127.0.0.1" | wc -l)
        echo "  Current entries: $current_entries"
        
        # Compare with available entries
        local available_entries=$(get_new_microservices_entries | grep "^127.0.0.1" | wc -l)
        echo "  Available entries: $available_entries"
        
        if [[ "$current_entries" -eq "$available_entries" ]]; then
            print_status "success" "Hosts file appears to be up to date"
        else
            print_status "warning" "Hosts file may need updating"
        fi
    fi
    
    echo ""
    list_backups "$hosts_file"
}

# =============================================================================
# USAGE AND HELP
# =============================================================================

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

DESCRIPTION:
    Intelligently updates system hosts file with microservices entries from 
    the generated context. Preserves existing hosts entries and only replaces
    the microservices section.

OPTIONS:
    -o, --os TYPE           Target operating system (auto, linux, macos, windows, wsl)
    -n, --no-backup         Skip creating backup before modification
    -d, --dry-run           Show what would be changed without making changes
    -v, --verbose           Show detailed output and changes
    -f, --force             Force update even if entries appear current
    -r, --restore [FILE]    Restore from backup (uses most recent if no file specified)
    -c, --check             Check current hosts file status only
    -h, --help              Show this help message

EXAMPLES:
    $0                                      # Auto-detect OS and update hosts
    $0 -o linux -v                        # Update Linux hosts with verbose output
    $0 -o windows --dry-run               # Preview changes on Windows
    $0 --check                            # Check current status
    $0 --restore                          # Restore most recent backup
    $0 --restore /etc/hosts.backup.file   # Restore specific backup

OPERATING SYSTEMS:
    auto        Auto-detect current OS (default)
    linux       Linux (/etc/hosts)
    macos       macOS (/etc/hosts)  
    windows     Windows hosts file (auto-detected mount in WSL)
    wsl         Windows Subsystem for Linux (/etc/hosts)

INTEGRATION:
    This script integrates with your existing microservices architecture:
    - Uses entries from: $CONTEXT_DIR/hosts.txt
    - Preserves all existing hosts entries
    - Only replaces the microservices section
    - Creates automatic backups for safety

HOSTS FILE SECTIONS:
    The script looks for and replaces content between:
    Start: "$SECTION_START"
    End:   Lines matching "$SECTION_END_PATTERN"

REQUIREMENTS:
    - Must be run after generate-context.sh to have current entries
    - May require sudo/elevated privileges for hosts file modification
    - Context directory: $CONTEXT_DIR

NOTES:
    - Always creates backups unless --no-backup is specified
    - Preserves file permissions and ownership
    - Cross-platform compatible
    - Safe for automation and CI/CD pipelines
EOF
}

# =============================================================================
# ARGUMENT PARSING
# =============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -o|--os)
                if [[ -n "$2" && "$2" != -* ]]; then
                    case "$2" in
                        auto|linux|macos|windows|wsl)
                            opt_os="$2"
                            ;;
                        *)
                            handle_error "Invalid OS type: $2"
                            ;;
                    esac
                    shift 2
                else
                    handle_error "Option $1 requires an OS type"
                fi
                ;;
            -n|--no-backup)
                opt_backup=false
                shift
                ;;
            -d|--dry-run)
                opt_dry_run=true
                opt_verbose=true
                shift
                ;;
            -v|--verbose)
                opt_verbose=true
                shift
                ;;
            -f|--force)
                opt_force=true
                shift
                ;;
            -r|--restore)
                opt_restore=true
                if [[ -n "$2" && "$2" != -* ]]; then
                    opt_restore_file="$2"
                    shift 2
                else
                    shift
                fi
                ;;
            -c|--check)
                opt_check_only=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                handle_error "Unknown option: $1. Use -h for help."
                ;;
        esac
    done
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Ensure context directory is set before displaying info
    if [[ -z "$CONTEXT_DIR" ]] || [[ ! -d "$CONTEXT_DIR" ]]; then
        # Try to find context directory
        if [[ -d "$PROJECT_ROOT/context" ]]; then
            CONTEXT_DIR="$PROJECT_ROOT/context"
        else
            CONTEXT_DIR="$PROJECT_ROOT/context (not found)"
        fi
    fi
    
    # Get hosts file path for target OS
    local hosts_file
    hosts_file=$(get_hosts_file_path "$opt_os")
    
    # Show initial detection info
    if [[ "$opt_os" == "windows" ]] && [[ "$opt_verbose" == "true" ]]; then
        print_status "info" "Detected WSL environment, found Windows hosts file: $hosts_file"
    fi
    
    print_status "info" "Hosts File Updater"
    print_status "info" "Target OS: $(detect_os) $(if [[ "$opt_os" != "auto" ]]; then echo "(forced: $opt_os)"; fi)"
    print_status "info" "Hosts file: $hosts_file"
    print_status "info" "Context dir: $CONTEXT_DIR"
    echo ""
    
    # Validate hosts file
    validate_hosts_file "$hosts_file"
    
    # Handle different modes
    if [[ "$opt_restore" == "true" ]]; then
        restore_backup "$hosts_file" "$opt_restore_file"
        return 0
    fi
    
    if [[ "$opt_check_only" == "true" ]]; then
        check_hosts_status "$hosts_file"
        return 0
    fi
    
    # Validate context directory and hosts file exist
    if [[ ! -d "$CONTEXT_DIR" ]]; then
        print_status "error" "Context directory not found: $CONTEXT_DIR"
        print_status "info" "Run generate-context.sh first to create context directory and hosts entries"
        exit 1
    fi
    
    local context_hosts_file="$CONTEXT_DIR/hosts.txt"
    if [[ ! -f "$context_hosts_file" ]]; then
        print_status "error" "Context hosts file not found: $context_hosts_file"
        print_status "info" "Run generate-context.sh first to create microservices hosts entries"
        exit 1
    fi
    
    # Check if update is needed (unless forced)
    if [[ "$opt_force" == "false" ]]; then
        local section_info=$(find_microservices_section "$hosts_file")
        local start_line=$(echo "$section_info" | cut -d' ' -f1)
        
        if [[ "$start_line" -ne 0 ]]; then
            local current_entries=$(sed -n "${start_line},${$(echo "$section_info" | cut -d' ' -f2)}p" "$hosts_file" | grep "^127.0.0.1" | wc -l)
            local available_entries=$(get_new_microservices_entries | grep "^127.0.0.1" | wc -l)
            
            if [[ "$current_entries" -eq "$available_entries" ]]; then
                print_status "success" "Hosts file appears to be up to date"
                print_status "info" "Use --force to update anyway or --verbose to see details"
                
                if [[ "$opt_verbose" == "true" ]]; then
                    echo ""
                    check_hosts_status "$hosts_file"
                fi
                
                return 0
            fi
        fi
    fi
    
    # Perform the update
    update_hosts_file "$hosts_file"
    
    print_status "success" "Hosts file update completed!"
    
    # Show final status if verbose
    if [[ "$opt_verbose" == "true" ]]; then
        echo ""
        check_hosts_status "$hosts_file"
    fi
}

# =============================================================================
# SCRIPT ENTRY POINT
# =============================================================================

# Only run main if script is executed directly (not sourced)
if [[ "${(%):-%x}" == "${0}" ]]; then
    main "$@"
fi