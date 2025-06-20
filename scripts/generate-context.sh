#!/bin/zsh

# Smart Context Sync - Organized project documentation

# Configuration - add/remove folders as needed
FOLDERS_TO_PROCESS=(
    "compose"
    "config" 
    "scripts"
    "docs"
)

# Output directory
CONTEXT_DIR="./context"

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

# Generate hosts file entries dynamically from compose files
generate_hosts_file() {
    local hosts_file="$CONTEXT_DIR/hosts.txt"
    local compose_dir="compose"
    
    echo "Generating dynamic hosts file entries from compose files..."
    
    # Create hosts file header
    cat > "$hosts_file" << 'EOF'
# Microservices Host File Entries
# Dynamically generated from Docker Compose files
# Add these to your system hosts file (/etc/hosts on Linux/Mac, C:\Windows\System32\drivers\etc\hosts on Windows)
EOF
    echo "# Generated on: $(date)" >> "$hosts_file"
    echo "# Source: Parsed from $compose_dir/ directory" >> "$hosts_file"
    echo "" >> "$hosts_file"
    
    # Check if compose directory exists
    if [[ ! -d "$compose_dir" ]]; then
        echo "# ERROR: $compose_dir directory not found!" >> "$hosts_file"
        echo "Warning: $compose_dir directory not found, cannot generate dynamic hosts"
        return 1
    fi
    
    # Simple approach - just collect all domains
    local all_domains=()
    
    # Parse all compose files
    for compose_file in $(find "$compose_dir" -name "*.yml" -o -name "*.yaml" 2>/dev/null | sort); do
        echo "  Parsing $compose_file..."
        
        # Extract Host rules
        local domains=$(grep -o 'Host(`[^`]*`)' "$compose_file" 2>/dev/null | sed 's/Host(`//g' | sed 's/`)//g')
        
        # Add found domains
        for domain in $domains; do
            if [[ "$domain" == *.test ]]; then
                if [[ "$domain" != *'$'* ]]; then
                    if [[ "$domain" != *'['* ]]; then
                        all_domains+=("$domain")
                    fi
                fi
            fi
        done
        
        # If no domains found, use filename
        if [[ -z "$domains" ]]; then
            local filename=$(basename "$compose_file" .yml)
            case "$filename" in
                mariadb|mysql|postgres|mongodb|redis|elasticsearch|logstash)
                    # Skip database services
                    ;;
                *)
                    all_domains+=("${filename}.test")
                    ;;
            esac
        fi
    done

    # if [[ -d "apps" ]]; then
        echo "  Scanning apps directory for projects..."
        for app_dir in apps/*/; do
            if [[ -d "$app_dir" ]]; then
                local app_name=$(basename "$app_dir")
                # Skip hidden directories and common non-project folders
                if [[ "$app_name" != .* && "$app_name" != "node_modules" && "$app_name" != "vendor" ]]; then
                    all_domains+=("${app_name}.test")
                    echo "    Found app: ${app_name}.test"
                fi
            fi
        done
    # fi
    
    # Remove duplicates and sort
    local unique_domains=($(printf '%s\n' "${all_domains[@]}" | sort -u))
    
    # Output domains grouped by lines
    echo "# All Services" >> "$hosts_file"
    local line="127.0.0.1"
    local count=0
    
    for domain in "${unique_domains[@]}"; do
        line="$line $domain"
        count=$((count + 1))
        
        if [[ $count -eq 6 ]]; then
            echo "$line" >> "$hosts_file"
            line="127.0.0.1"
            count=0
        fi
    done
    
    # Output remaining domains
    if [[ $count -gt 0 ]]; then
        echo "$line" >> "$hosts_file"
    fi
    
    echo "" >> "$hosts_file"
    echo "# Summary: ${#unique_domains[@]} domains found" >> "$hosts_file"
    
    echo "Generated dynamic hosts file: $hosts_file (${#unique_domains[@]} unique domains)"
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
        tree "$folder" >> "$output_file"
    else
        find "$folder" -type f | sort >> "$output_file"
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
    
    local size=$(wc -c < "$output_file")
    echo "Created $output_file ($file_count files, $(( size / 1024 ))KB)"
}

# Main function
main() {
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
        git log --oneline -20 2>/dev/null >> "$index_file" || echo "No git history" >> "$index_file"
        echo "" >> "$index_file"
        
        # Working directory status
        echo "Working directory:" >> "$index_file"
        if git diff --quiet && git diff --cached --quiet; then
            echo "Clean (no changes)" >> "$index_file"
        else
            echo "Has uncommitted changes:" >> "$index_file"
            git status --porcelain 2>/dev/null | head -10 >> "$index_file"
            if [[ $(git status --porcelain 2>/dev/null | wc -l) -gt 10 ]]; then
                echo "... (showing first 10 of $(git status --porcelain 2>/dev/null | wc -l) changes)" >> "$index_file"
            fi
        fi
        echo "" >> "$index_file"
        
        # Remote information
        echo "Remote info:" >> "$index_file"
        local remote_url=$(git remote get-url origin 2>/dev/null || echo "No remote origin")
        echo "Origin: $remote_url" >> "$index_file"
        
        # Ahead/behind status
        local ahead_behind=$(git rev-list --left-right --count origin/$(git branch --show-current)...HEAD 2>/dev/null || echo "0	0")
        local behind=$(echo "$ahead_behind" | cut -f1)
        local ahead=$(echo "$ahead_behind" | cut -f2)
        if [[ "$ahead" != "0" || "$behind" != "0" ]]; then
            echo "Sync status: $ahead commits ahead, $behind commits behind origin" >> "$index_file"
        else
            echo "Sync status: Up to date with origin" >> "$index_file"
        fi
        echo "" >> "$index_file"
        
        # Recent tags
        local recent_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "No tags")
        echo "Latest tag: $recent_tag" >> "$index_file"
        
        # Stash info
        local stash_count=$(git stash list 2>/dev/null | wc -l)
        if [[ "$stash_count" -gt 0 ]]; then
            echo "Stashes: $stash_count stashed changes" >> "$index_file"
            git stash list --oneline 2>/dev/null | head -3 >> "$index_file"
        else
            echo "Stashes: None" >> "$index_file"
        fi
        echo "" >> "$index_file"
    fi
    
    # Add project structure to index
    echo "## Project Structure" >> "$index_file"
    if command -v tree >/dev/null 2>&1; then
        # Use tree but limit depth and exclude deep app contents
        tree -I 'node_modules|.git|dist|build|venv|__pycache__|target|context' -L 2 >> "$index_file"
    else
        # Manual listing - show top level and apps folder contents
        echo "." >> "$index_file"
        find . -maxdepth 1 -type d -not -path '*/.*' -not -path '.' | sort >> "$index_file"
        
        # Show apps folder contents specifically
        if [[ -d "./apps" ]]; then
            echo "apps/" >> "$index_file"
            find ./apps -maxdepth 1 -type d -not -path './apps' | sed 's|./apps/|  ├── |' | sort >> "$index_file"
        fi
    fi
    echo "" >> "$index_file"
    
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