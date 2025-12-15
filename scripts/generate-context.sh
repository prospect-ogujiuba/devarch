#!/bin/zsh

# Smart Context Sync - Organized project documentation

# Configuration - add/remove folders as needed
FOLDERS_TO_PROCESS=(
    "compose"
    "config" 
    "scripts"

)

# Output directory
CONTEXT_DIR="./context"

# Defaults for optional actions
APPLY_HOSTS=false
# Prefer the correct Windows hosts path; fallback will be handled at runtime
DEFAULT_WINDOWS_HOSTS_TARGET="/mnt/c/Windows/System32/drivers/etc/hosts"
HOSTS_TARGET="${HOSTS_TARGET:-$DEFAULT_WINDOWS_HOSTS_TARGET}"

print_usage() {
    cat <<'USAGE'
Usage: scripts/generate-context.sh [options]

Options:
  --apply-hosts                After generating context/hosts.txt, apply it to the Windows hosts file (requires sudo).
  --hosts-target=<path>        Override the target hosts path (default: /mnt/c/Windows/System32/drivers/etc/hosts).
  -h, --help                   Show this help.

Notes:
  - On Windows via WSL, the typical hosts file is at /mnt/c/Windows/System32/drivers/etc/hosts
USAGE
}

# Apply generated hosts to Windows hosts file (WSL)
apply_hosts_to_windows() {
    local source_file="$CONTEXT_DIR/hosts.txt"
    local target_file="$HOSTS_TARGET"

    # If the default path doesn't exist, try historical/alternate path without etc
    if [[ ! -e "$target_file" ]] && [[ "$HOSTS_TARGET" == "$DEFAULT_WINDOWS_HOSTS_TARGET" ]]; then
        local alt="/mnt/c/Windows/System32/drivers/hosts"
        if [[ -e "$alt" ]]; then
            target_file="$alt"
        fi
    fi

    if [[ ! -f "$source_file" ]]; then
        echo "Error: $source_file not found."
        return 1
    fi

    # Create a backup if target exists
    if [[ -f "$target_file" ]]; then
        local backup="${target_file}.bak.$(date +%Y%m%d%H%M%S)"
        echo "Backing up existing hosts to: $backup"
        sudo cp "$target_file" "$backup" || {
            echo "Warning: could not create backup of $target_file";
        }
    fi

    echo "Applying $source_file -> $target_file (requires sudo)"
    # Use tee to handle permissions and preserve content exactly
    # Using printf to avoid issues with cat and redirection permissions
    if sudo sh -c "cat '$source_file' > '$target_file'"; then
        echo "Hosts file updated: $target_file"
        return 0
    else
        echo "Attempt with direct redirection failed, falling back to tee..."
        if cat "$source_file" | sudo tee "$target_file" >/dev/null; then
            echo "Hosts file updated: $target_file"
            return 0
        else
            echo "Error: Failed to update hosts file at $target_file"
            return 1
        fi
    fi
}

# Helper function to check if file is likely text/readable
is_text_file() {
    local file="$1"
    
    # Skip known binary extensions
    case "${file##*.}" in
        jpg|jpeg|png|gif|bmp|svg|ico|pdf|zip|tar|gz|bz2|xz|7z|rar|exe|dll|so|dylib|bin|dat|db|sqlite|lock)
            return 1
            ;;
    esac
    
    # Use file command if available, otherwise assume text
    if command -v file >/dev/null 2>&1; then
        file "$file" | grep -q -E "(text|JSON|XML|script|shell|Python|JavaScript|HTML|CSS)"
    else
        return 0
    fi
}

# Utilities
# Normalize line endings of a file to LF (remove CR characters)
normalize_line_endings() {
    local f="$1"
    [[ -f "$f" ]] || return 0
    # Use a safe temp file in the same directory
    local tmp="${f}.tmp.$$"
    # Remove carriage returns if any
    tr -d '\r' < "$f" > "$tmp" 2>/dev/null && mv "$tmp" "$f" || rm -f "$tmp"
}

# Generate hosts file entries dynamically from compose files
generate_hosts_file() {
    local hosts_file="$CONTEXT_DIR/hosts.txt"
    local compose_dir="compose"

    echo "Generating dynamic hosts file entries from compose files and apps..."

    cat > "$hosts_file" << 'EOF'
# Microservices Host File Entries
# Dynamically generated from Docker Compose and App directories
# Add these to your system hosts file (/etc/hosts on Linux/Mac, C:\Windows\System32\drivers\etc\hosts on Windows)
EOF
    echo "# Generated on: $(date)" >> "$hosts_file"
    echo "# Source: Parsed from $compose_dir/ and apps/ directory" >> "$hosts_file"
    echo "" >> "$hosts_file"

    local compose_domains=()
    local app_domains=()

    # Parse Compose domains (infrastructure services - keep .test)
    for compose_file in $(find "$compose_dir" -name "*.yml" -o -name "*.yaml" 2>/dev/null | sort); do
        echo "  Parsing $compose_file..."
        local domains=$(grep -o 'Host(`[^`]*`)' "$compose_file" | sed 's/Host(`//g' | sed 's/`)//g')
        for domain in $domains; do
            [[ "$domain" == *.test && "$domain" != *'$'* && "$domain" != *'['* ]] && compose_domains+=("$domain")
        done

        if [[ -z "$domains" ]]; then
            local filename=$(basename "$compose_file" .yml)
            case "$filename" in mariadb|mysql|postgres|mongodb|redis|elasticsearch|logstash) ;; 
            *) compose_domains+=("${filename}.test") ;; 
            esac
        fi
    done

    # Parse App domains (development projects - NOW USE .test instead of .dev)
    echo "  Scanning apps directory for development projects..."
    for app_dir in apps/*/; do
        [[ -d "$app_dir" ]] || continue
        local app_name=$(basename "$app_dir")
        [[ "$app_name" != .* && "$app_name" != "node_modules" && "$app_name" != "vendor" ]] && app_domains+=("${app_name}.test")
    done

    # Output Infrastructure Services (Compose domains)
    if [[ ${#compose_domains[@]} -gt 0 ]]; then
        echo "# Infrastructure Services (.test domains)" >> "$hosts_file"
        local line="127.0.0.1" count=0
        for domain in $(printf '%s\n' "${compose_domains[@]}" | sort -u); do
            line="$line $domain"
            count=$((count + 1))
            [[ $count -eq 6 ]] && { echo "$line" >> "$hosts_file"; line="127.0.0.1"; count=0; }
        done
        [[ $count -gt 0 ]] && echo "$line" >> "$hosts_file"
        echo "" >> "$hosts_file"
    fi

    # Output Development Projects (App domains - NOW ALL .test)
    if [[ ${#app_domains[@]} -gt 0 ]]; then
        echo "# Development Projects (.test domains)" >> "$hosts_file"
        local line="127.0.0.1" count=0
        for domain in $(printf '%s\n' "${app_domains[@]}" | sort -u); do
            line="$line $domain"
            count=$((count + 1))
            [[ $count -eq 6 ]] && { echo "$line" >> "$hosts_file"; line="127.0.0.1"; count=0; }
        done
        [[ $count -gt 0 ]] && echo "$line" >> "$hosts_file"
        echo "" >> "$hosts_file"
    fi

    # Add special development domains (UPDATED)
    echo "# Development Tools" >> "$hosts_file"
    echo "127.0.0.1 projects.test" >> "$hosts_file"
    echo "" >> "$hosts_file"

    echo "# Summary: $(( ${#compose_domains[@]} + ${#app_domains[@]} + 1 )) domains found" >> "$hosts_file"
    echo "# Infrastructure (.test): ${#compose_domains[@]} domains" >> "$hosts_file"
    echo "# Development (.test): $(( ${#app_domains[@]} + 1 )) domains" >> "$hosts_file"
    
    # Normalize line endings to LF
    normalize_line_endings "$hosts_file"

    echo "Generated dynamic hosts file: $hosts_file"
}

# Process a single folder
process_folder() {
    local folder="$1"
    local output_file="$CONTEXT_DIR/${folder}.txt"
    local file_count=0
    
    echo "Processing folder: $folder"
    
    # If folder doesn't exist, skip it
    if [[ ! -d "$folder" ]]; then
        echo "Warning: Folder '$folder' not found, skipping"
        return
    fi
    
    # Header
    echo "# $(basename "$(pwd)") - $folder" > "$output_file"
    echo "Generated: $(date)" >> "$output_file"
    echo "Folder: $folder" >> "$output_file"
    echo "" >> "$output_file"
    
    # Get folder structure
    echo "## Folder Structure" >> "$output_file"
    if command -v tree >/dev/null 2>&1; then
        tree "$folder" | sed 's/^/- /' >> "$output_file"
    else
        find "$folder" -type f | sort | sed 's/^/- /' >> "$output_file"
    fi
    echo "" >> "$output_file"
    
    # Process all files in the folder
    echo "## Files" >> "$output_file"
    
    for file in $(find "$folder" -type f 2>/dev/null | sort); do
        # Check if it's a readable text file
        if is_text_file "$file"; then
            echo "" >> "$output_file"
            echo "### $file" >> "$output_file"
            echo '```' >> "$output_file"
            
            # Add the file content (limit to 500 lines to prevent massive files)
            if [[ -r "$file" ]]; then
                head -500 "$file" >> "$output_file"
                
                # Add note if file was truncated
                local total_lines=$(wc -l < "$file" 2>/dev/null || echo "0")
                if [[ $total_lines -gt 500 ]]; then
                    echo "" >> "$output_file"
                    echo "... (file truncated - showing first 500 of $total_lines lines)" >> "$output_file"
                fi
            else
                echo "(file not readable)" >> "$output_file"
            fi
            
            echo '```' >> "$output_file"
            file_count=$((file_count + 1))
        fi
    done
    
    # Footer
    echo "" >> "$output_file"
    echo "---" >> "$output_file"
    echo "Files processed: $file_count" >> "$output_file"
    
    # Normalize line endings to LF
    normalize_line_endings "$output_file"

    local size=$(wc -c < "$output_file")
    echo "Created $output_file ($file_count files, $(( size / 1024 ))KB)"
}

# Main function
main() {
    # Parse args
    for arg in "$@"; do
        case "$arg" in
            --apply-hosts)
                APPLY_HOSTS=true
                shift
                ;;
            --hosts-target=*)
                HOSTS_TARGET="${arg#*=}"
                shift
                ;;
            -h|--help)
                print_usage
                return 0
                ;;
        esac
    done

    local project_name=$(basename "$(pwd)")
    
    echo "Smart Context Sync for: $project_name"
    echo "=================================="
    
    # If we're in a scripts folder, go up one level
    if [[ "$(basename "$(pwd)")" == "scripts" ]] || [[ "$(basename "$(pwd)")" == "bin" ]] || [[ "$(basename "$(pwd)")" == "tools" ]]; then
        cd ..
        echo "Moved up to project root: $(pwd)"
    fi
    
    # Create context directory if it doesn't exist
    mkdir -p "$CONTEXT_DIR"
    
    # Generate hosts file entries
    generate_hosts_file

    # Optionally apply to Windows hosts file
    if [[ "$APPLY_HOSTS" == true ]]; then
        apply_hosts_to_windows || {
            echo "Failed to apply hosts file. You may need to run this script with sudo inside WSL."
        }
    fi
    
    # Create master index file
    local index_file="$CONTEXT_DIR/index.txt"
    echo "# $project_name - Context Index" > "$index_file"
    echo "Generated: $(date)" >> "$index_file"
    echo "" >> "$index_file"
    
    # Add git context if available
    if git rev-parse --git-dir >/dev/null 2>&1; then
        echo "## Git Status" >> "$index_file"
        echo "Branch: $(git branch --show-current 2>/dev/null || echo 'unknown')" >> "$index_file"
        echo "" >> "$index_file"
        
        echo "Recent commits:" >> "$index_file"
        git log --oneline -20 2>/dev/null | sed 's/^/- /' >> "$index_file" || echo "- No git history" >> "$index_file"
        echo "" >> "$index_file"
        
        # Working directory status
        echo "Working directory:" >> "$index_file"
        if git diff --quiet && git diff --cached --quiet; then
            echo "- Clean (no changes)" >> "$index_file"
        else
            echo "Has uncommitted changes:" >> "$index_file"
            git status --porcelain 2>/dev/null | head -10 | sed 's/^/- /' >> "$index_file"
            if [[ $(git status --porcelain 2>/dev/null | wc -l) -gt 10 ]]; then
                echo "- ... (showing first 10 of $(git status --porcelain 2>/dev/null | wc -l) changes)" >> "$index_file"
            fi
        fi
        echo "" >> "$index_file"
        
        # Remote information
        echo "Remote info:" >> "$index_file"
        local remote_url=$(git remote get-url origin 2>/dev/null || echo "No remote origin")
        echo "- Origin: $remote_url" >> "$index_file"
        
        # Ahead/behind status
        local ahead_behind=$(git rev-list --left-right --count origin/$(git branch --show-current)...HEAD 2>/dev/null || echo "0	0")
        local behind=$(echo "$ahead_behind" | cut -f1)
        local ahead=$(echo "$ahead_behind" | cut -f2)
        if [[ "$ahead" != "0" || "$behind" != "0" ]]; then
            echo "- Sync status: $ahead commits ahead, $behind commits behind origin" >> "$index_file"
        else
            echo "- Sync status: Up to date with origin" >> "$index_file"
        fi
        echo "" >> "$index_file"
        
        # Recent tags
        local recent_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "No tags")
        echo "- Latest tag: $recent_tag" >> "$index_file"
        
        # Stash info
        local stash_count=$(git stash list 2>/dev/null | wc -l)
        if [[ "$stash_count" -gt 0 ]]; then
            echo "Stashes: $stash_count stashed changes" >> "$index_file"
            git stash list --oneline 2>/dev/null | head -3 | sed 's/^/- /' >> "$index_file"
        else
            echo "- Stashes: None" >> "$index_file"
        fi
        echo "" >> "$index_file"
    fi

    # Add environment context if .env file exists
    if [[ -f ".env" ]]; then
        echo "## Environment Configuration" >> "$index_file"
        echo "Environment file: .env" >> "$index_file"
        echo "" >> "$index_file"
        
        # Show .env contents with sensitive data partially masked
        echo "Environment variables:" >> "$index_file"
        echo '```' >> "$index_file"
        
        # Read .env file and mask sensitive values (passwords, tokens, etc.)
        while IFS= read -r line; do
            # Skip empty lines and comments
            if [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]]; then
                echo "$line" >> "$index_file"
                continue
            fi
            
            # Check if line contains sensitive data patterns
            if [[ "$line" =~ (PASSWORD|TOKEN|SECRET|KEY|PASS)= ]]; then
                # Extract variable name and mask the value
                local var_name="${line%%=*}"
                local var_value="${line#*=}"
                if [[ ${#var_value} -gt 0 ]]; then
                    echo "${var_name}=***masked***" >> "$index_file"
                else
                    echo "$line" >> "$index_file"
                fi
            else
                echo "$line" >> "$index_file"
            fi
        done < ".env"
        
        echo '```' >> "$index_file"
        echo "" >> "$index_file"
        
        # Add environment statistics
        local total_vars=$(grep -c "^[^#]*=" ".env" 2>/dev/null || echo "0")
        local masked_vars=$(grep -c -E "(PASSWORD|TOKEN|SECRET|KEY|PASS)=" ".env" 2>/dev/null || echo "0")
        echo "Statistics: $total_vars total variables, $masked_vars sensitive values masked" >> "$index_file"
        echo "" >> "$index_file"
    else
        echo "## Environment Configuration" >> "$index_file"
        echo "No .env file found in project root" >> "$index_file"
        echo "" >> "$index_file"
    fi
    
    # Add project structure to index
    echo "## Project Structure" >> "$index_file"
    if command -v tree >/dev/null 2>&1; then
        # Use tree but limit depth and exclude deep app contents
        tree -I 'node_modules|.git|dist|build|venv|__pycache__|target|context' -L 2 | sed 's/^/- /' >> "$index_file"
    else
        # Manual listing - show top level and apps folder contents
        echo "- ." >> "$index_file"
        find . -maxdepth 1 -type d -not -path '*/.*' -not -path '.' | sort | sed 's/^/- /' >> "$index_file"
        
        # Show apps folder contents specifically
        if [[ -d "./apps" ]]; then
            echo "- apps/" >> "$index_file"
            find ./apps -maxdepth 1 -type d -not -path './apps' | sort | sed 's|./apps/|  - |' >> "$index_file"
        fi
    fi
    echo "" >> "$index_file"

    # Detailed folder structures (full depth) for compose, config, and scripts
    # Using the same logic as individual context files, with markdown bullets
    for section_folder in compose config scripts; do
        if [[ -d "$section_folder" ]]; then
            echo "### ${section_folder} - Folder Structure" >> "$index_file"
            if command -v tree >/dev/null 2>&1; then
                tree "$section_folder" | sed 's/^/- /' >> "$index_file"
            else
                find "$section_folder" -type f | sort | sed 's/^/- /' >> "$index_file"
            fi
            echo "" >> "$index_file"
        fi
    done
    
    # List context files that will be generated
    echo "## Context Files" >> "$index_file"
    for folder in "${FOLDERS_TO_PROCESS[@]}"; do
        echo "- ${folder}.txt - Contents of $folder/ directory" >> "$index_file"
    done
    echo "" >> "$index_file"
    
    # Process each configured folder
    local total_files=0
    local total_size=0
    
    for folder in "${FOLDERS_TO_PROCESS[@]}"; do
        process_folder "$folder"
        
        # Update totals if file was created
        local output_file="$CONTEXT_DIR/${folder}.txt"
        if [[ -f "$output_file" ]]; then
            local file_count=$(tail -1 "$output_file" | grep -o '[0-9]*' | head -1)
            local size=$(wc -c < "$output_file")
            total_files=$((total_files + file_count))
            total_size=$((total_size + size))
        fi
    done
    
    # Update index with summary
    echo "## Summary" >> "$index_file"
    echo "- Total files processed: $total_files" >> "$index_file"
    echo "- Total context size: $(( total_size / 1024 ))KB" >> "$index_file"
    echo "- Folders processed: ${FOLDERS_TO_PROCESS[*]}" >> "$index_file"

    # Normalize line endings to LF
    normalize_line_endings "$index_file"
    
    echo ""
    echo "=================================="
    echo "Context generation complete!"
    echo "Output directory: $CONTEXT_DIR"
    echo "Total: $total_files files, $(( total_size / 1024 ))KB"
    echo ""
    echo "Generated files:"
    ls -la "$CONTEXT_DIR"
    echo ""
    echo "To add services to your system hosts file:"
    echo "   sudo cat $CONTEXT_DIR/hosts.txt >> /etc/hosts         # Linux/Mac"
    echo "   cat $CONTEXT_DIR/hosts.txt >> C:\\Windows\\System32\\drivers\\etc\\hosts  # Windows"
}

# Show usage if help requested
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    echo "Smart Context Sync"
    echo "=================="
    echo ""
    echo "Generates organized context files from specified project folders."
    echo ""
    echo "Usage: $0"
    echo ""
    echo "Configuration (edit script to modify):"
    echo "- FOLDERS_TO_PROCESS: Array of folders to process"
    echo "- CONTEXT_DIR: Output directory for context files"
    echo ""
    echo "Current folders configured:"
    for folder in "${FOLDERS_TO_PROCESS[@]}"; do
        echo "  - $folder"
    done
    echo ""
    echo "Output files:"
    echo "  - ./context/index.txt - Project overview and git status"
    echo "  - ./context/{folder}.txt - Contents of each folder"
    echo "  - ./context/hosts.txt - Dynamic hosts file from compose files"
    echo ""
    exit 0
fi

# Run main function
main "$@"