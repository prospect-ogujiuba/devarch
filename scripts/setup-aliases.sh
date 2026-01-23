#!/bin/bash

# =============================================================================
# DEVARCH ALIAS SETUP SCRIPT
# =============================================================================
# Automatically detects shells and sets up aliases for service-manager.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVICE_MANAGER="$PROJECT_ROOT/scripts/service-manager.sh"

# Alias definitions
ALIASES=(
    "devarch"
    "dvrc"
    "dv"
    "da"
)

# Shell config files
declare -A SHELL_CONFIGS=(
    ["bash"]="$HOME/.bashrc"
    ["zsh"]="$HOME/.zshrc"
    ["fish"]="$HOME/.config/fish/config.fish"
)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    local type="$1"
    local msg="$2"
    case "$type" in
        success) echo -e "${GREEN}✓${NC} $msg" ;;
        error)   echo -e "${RED}✗${NC} $msg" ;;
        info)    echo -e "${BLUE}ℹ${NC} $msg" ;;
        warn)    echo -e "${YELLOW}⚠${NC} $msg" ;;
    esac
}

detect_shells() {
    local detected=()

    # Check for installed shells
    command -v bash &>/dev/null && [[ -f "${SHELL_CONFIGS[bash]}" ]] && detected+=("bash")
    command -v zsh &>/dev/null && [[ -f "${SHELL_CONFIGS[zsh]}" ]] && detected+=("zsh")
    command -v fish &>/dev/null && {
        mkdir -p "$HOME/.config/fish" 2>/dev/null
        detected+=("fish")
    }

    echo "${detected[@]}"
}

get_current_shell() {
    basename "$SHELL"
}

check_existing_aliases() {
    local config_file="$1"
    local shell_type="$2"

    if [[ ! -f "$config_file" ]]; then
        return 1
    fi

    if [[ "$shell_type" == "fish" ]]; then
        grep -q "alias devarch=" "$config_file" 2>/dev/null || \
        grep -q "function devarch" "$config_file" 2>/dev/null
    else
        grep -q "alias devarch=" "$config_file" 2>/dev/null
    fi
}

generate_aliases_bash() {
    local output=""
    output+="\n# DevArch aliases\n"
    for alias_name in "${ALIASES[@]}"; do
        output+="alias ${alias_name}='${SERVICE_MANAGER}'\n"
    done
    echo -e "$output"
}

generate_aliases_fish() {
    local output=""
    output+="\n# DevArch aliases\n"
    for alias_name in "${ALIASES[@]}"; do
        output+="alias ${alias_name}='${SERVICE_MANAGER}'\n"
    done
    echo -e "$output"
}

add_aliases() {
    local shell_type="$1"
    local config_file="${SHELL_CONFIGS[$shell_type]}"

    if [[ -z "$config_file" ]]; then
        print_status error "Unknown shell: $shell_type"
        return 1
    fi

    # Check if aliases already exist
    if check_existing_aliases "$config_file" "$shell_type"; then
        print_status warn "Aliases already exist in $config_file"
        read -p "Overwrite existing aliases? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status info "Skipped $shell_type"
            return 0
        fi
        # Remove existing devarch aliases
        if [[ "$shell_type" == "fish" ]]; then
            sed -i '/# DevArch aliases/,/^$/d' "$config_file" 2>/dev/null
            sed -i '/alias devarch=/d;/alias dvrc=/d;/alias dv=/d;/alias da=/d' "$config_file" 2>/dev/null
        else
            sed -i '/# DevArch aliases/,/^$/d' "$config_file" 2>/dev/null
            sed -i '/alias devarch=/d;/alias dvrc=/d;/alias dv=/d;/alias da=/d' "$config_file" 2>/dev/null
        fi
    fi

    # Add aliases
    if [[ "$shell_type" == "fish" ]]; then
        generate_aliases_fish >> "$config_file"
    else
        generate_aliases_bash >> "$config_file"
    fi

    print_status success "Added aliases to $config_file"
    return 0
}

show_menu() {
    local shells=($@)
    local current_shell=$(get_current_shell)

    echo ""
    echo "Detected shells:"
    echo ""

    local i=1
    for shell in "${shells[@]}"; do
        local marker=""
        [[ "$shell" == "$current_shell" ]] && marker=" (current)"
        echo "  $i) $shell$marker"
        ((i++))
    done
    echo "  a) All detected shells"
    echo "  q) Quit"
    echo ""
}

main() {
    echo ""
    echo "DevArch Alias Setup"
    echo "==================="
    echo ""
    print_status info "Project: $PROJECT_ROOT"
    print_status info "Script:  $SERVICE_MANAGER"
    echo ""
    print_status info "Aliases to create: ${ALIASES[*]}"

    # Detect shells
    local detected_shells=($(detect_shells))

    if [[ ${#detected_shells[@]} -eq 0 ]]; then
        print_status error "No supported shells detected"
        exit 1
    fi

    show_menu "${detected_shells[@]}"

    read -p "Select shell(s) to configure: " choice

    local selected_shells=()

    case "$choice" in
        q|Q)
            print_status info "Cancelled"
            exit 0
            ;;
        a|A)
            selected_shells=("${detected_shells[@]}")
            ;;
        *)
            if [[ "$choice" =~ ^[0-9]+$ ]] && [[ "$choice" -ge 1 ]] && [[ "$choice" -le ${#detected_shells[@]} ]]; then
                selected_shells=("${detected_shells[$((choice-1))]}")
            else
                print_status error "Invalid selection"
                exit 1
            fi
            ;;
    esac

    echo ""

    local success_count=0
    for shell in "${selected_shells[@]}"; do
        if add_aliases "$shell"; then
            ((success_count++))
        fi
    done

    echo ""

    if [[ $success_count -gt 0 ]]; then
        print_status success "Setup complete!"
        echo ""
        print_status info "To use aliases in current terminal:"
        for shell in "${selected_shells[@]}"; do
            local config="${SHELL_CONFIGS[$shell]}"
            echo "    source $config"
        done
        echo ""
        print_status info "New terminals will load aliases automatically"
        echo ""
        print_status info "Usage: devarch list, dv up postgres, da status"
    fi
}

main "$@"
