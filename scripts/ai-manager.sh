#!/usr/bin/env bash

# =============================================================================
# DEVARCH AI - LLM Assistant Manager
# =============================================================================

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DEVARCH_API_URL="${DEVARCH_API_URL:-http://localhost:8550}"
DEVARCH_API_KEY="${DEVARCH_API_KEY:-test}"

api_call() {
    local method="$1" endpoint="$2" data="$3"
    local url="${DEVARCH_API_URL}/api/v1${endpoint}"
    local args=(-s -X "$method" -H "X-API-Key: ${DEVARCH_API_KEY}" -H "Content-Type: application/json")
    if [[ -n "$data" ]]; then
        args+=(-d "$data")
    fi
    curl "${args[@]}" "$url"
}

api_stream() {
    local endpoint="$1" data="$2"
    local url="${DEVARCH_API_URL}/api/v1${endpoint}"
    curl -sN -X POST -H "X-API-Key: ${DEVARCH_API_KEY}" -H "Content-Type: application/json" -d "$data" "$url"
}

cmd_status() {
    local resp
    resp=$(api_call GET /ai/status)

    local running model port gpu uptime
    running=$(echo "$resp" | jq -r '.data.running')
    model=$(echo "$resp" | jq -r '.data.model')
    port=$(echo "$resp" | jq -r '.data.port')
    gpu=$(echo "$resp" | jq -r '.data.gpu')
    uptime=$(echo "$resp" | jq -r '.data.uptime // "n/a"')

    echo "DevArch AI Status"
    echo "─────────────────"
    if [[ "$running" == "true" ]]; then
        echo "  Status:  running"
        echo "  Model:   $model"
        echo "  Port:    $port"
        echo "  GPU:     $gpu"
        echo "  Uptime:  $uptime"
    else
        echo "  Status:  stopped"
        echo "  Model:   $model (configured)"
        echo "  GPU:     $gpu"
    fi
}

cmd_stop() {
    echo "Stopping LLM container..."
    local resp
    resp=$(api_call POST /ai/stop)
    echo "$resp" | jq -r '.data.message // .error.message'
}

cmd_generate() {
    local description="$1"
    if [[ -z "$description" ]]; then
        echo "Usage: devarch ai generate \"description\""
        return 1
    fi

    echo "Generating service template..."
    local resp
    resp=$(api_call POST /ai/generate "{\"description\":$(echo "$description" | jq -Rs .)}")

    # Check for error
    local err
    err=$(echo "$resp" | jq -r '.error.message // empty')
    if [[ -n "$err" ]]; then
        echo "Error: $err"
        return 1
    fi

    # Check if we got parsed service or raw
    local service raw
    service=$(echo "$resp" | jq '.data.service // empty')
    raw=$(echo "$resp" | jq -r '.data.raw // empty')

    if [[ -n "$service" && "$service" != "null" ]]; then
        echo ""
        echo "$service" | jq .
        echo ""
        read -r -p "Create this service? [y/N] " confirm
        if [[ "$confirm" =~ ^[yY]$ ]]; then
            local name
            name=$(echo "$service" | jq -r '.name')
            local create_resp
            create_resp=$(api_call POST /services "$service")
            local create_err
            create_err=$(echo "$create_resp" | jq -r '.error.message // empty')
            if [[ -n "$create_err" ]]; then
                echo "Error creating service: $create_err"
                return 1
            fi
            echo "Service '$name' created successfully"
        else
            echo "Cancelled"
        fi
    elif [[ -n "$raw" ]]; then
        echo ""
        echo "LLM response (could not parse as service JSON):"
        echo "$raw"
    else
        echo "Unexpected response:"
        echo "$resp" | jq .
    fi
}

cmd_chat() {
    echo "DevArch AI Assistant (type 'exit' to quit, 'exec' to run last command)"
    echo "───"

    local conv_id=""
    local last_command=""

    while true; do
        printf "> "
        read -r input
        [[ -z "$input" ]] && continue

        if [[ "$input" == "exit" || "$input" == "quit" ]]; then
            echo "Goodbye"
            break
        fi

        if [[ "$input" == "exec" ]]; then
            if [[ -z "$last_command" ]]; then
                echo "No command to execute"
                continue
            fi
            read -r -p "Execute: $last_command ? [y/N] " confirm
            if [[ "$confirm" =~ ^[yY]$ ]]; then
                eval "$last_command"
            fi
            continue
        fi

        local payload="{\"message\":$(echo "$input" | jq -Rs .)"
        if [[ -n "$conv_id" ]]; then
            payload="${payload%\}},\"conversation_id\":\"$conv_id\"}"
        else
            payload="${payload}}"
        fi

        local resp
        resp=$(api_call POST /ai/chat "$payload")

        local err
        err=$(echo "$resp" | jq -r '.error.message // empty')
        if [[ -n "$err" ]]; then
            echo "Error: $err"
            continue
        fi

        local message
        message=$(echo "$resp" | jq -r '.data.message')
        conv_id=$(echo "$resp" | jq -r '.data.conversation_id')

        echo ""
        echo "$message"
        echo ""

        # Extract last command suggestion (line starting with $)
        local cmd
        cmd=$(echo "$message" | grep '^\$ ' | tail -1 | sed 's/^\$ //')
        if [[ -n "$cmd" ]]; then
            last_command="$cmd"
        fi
    done
}

cmd_diagnose() {
    local target="$1"
    if [[ -z "$target" ]]; then
        echo "Usage: devarch ai diagnose <service|stack|instance>"
        return 1
    fi

    echo "Diagnosing $target..."
    local resp
    resp=$(api_call POST /ai/diagnose "{\"target\":$(echo "$target" | jq -Rs .)}")

    local err
    err=$(echo "$resp" | jq -r '.error.message // empty')
    if [[ -n "$err" ]]; then
        echo "Error: $err"
        return 1
    fi

    echo ""
    echo "$resp" | jq -r '.data.diagnosis'
}

cmd_model_pull() {
    local model="$1"
    if [[ -z "$model" ]]; then
        echo "Usage: devarch ai model pull <model-name>"
        return 1
    fi

    echo "Pulling model $model..."
    local resp
    resp=$(api_call POST /ai/model/pull "{\"model\":$(echo "$model" | jq -Rs .)}")

    local err
    err=$(echo "$resp" | jq -r '.error.message // empty')
    if [[ -n "$err" ]]; then
        echo "Error: $err"
        return 1
    fi

    echo "$resp" | jq -r '.data.message'
}

cmd_model_list() {
    local resp
    resp=$(api_call GET /ai/models)

    local err
    err=$(echo "$resp" | jq -r '.error.message // empty')
    if [[ -n "$err" ]]; then
        echo "Error: $err"
        return 1
    fi

    echo "Available models:"
    echo "$resp" | jq -r '.data.models[]? // empty'
}

show_help() {
    cat << 'EOF'
devarch ai - AI Assistant

USAGE:
  devarch ai <command> [args...]

COMMANDS:
  status                  Show LLM runtime status
  generate "description"  Generate service template from description
  chat                    Interactive chat session
  diagnose <target>       Diagnose service/stack/instance issues
  model pull <name>       Pull a model
  model list              List available models
  stop                    Stop LLM container

EXAMPLES:
  devarch ai status
  devarch ai generate "redis cache with persistence"
  devarch ai chat
  devarch ai diagnose nginx
  devarch ai model pull granite3.2:8b
  devarch ai stop
EOF
}

main() {
    if [[ $# -eq 0 || "$1" == "-h" || "$1" == "--help" || "$1" == "help" ]]; then
        show_help
        return 0
    fi

    local cmd="$1"
    shift

    case "$cmd" in
        status)   cmd_status "$@" ;;
        generate) cmd_generate "$@" ;;
        chat)     cmd_chat "$@" ;;
        diagnose) cmd_diagnose "$@" ;;
        stop)     cmd_stop "$@" ;;
        model)
            local subcmd="${1:-help}"
            shift 2>/dev/null || true
            case "$subcmd" in
                pull) cmd_model_pull "$@" ;;
                list) cmd_model_list "$@" ;;
                *)    echo "Usage: devarch ai model <pull|list>" ;;
            esac
            ;;
        *)
            echo "Unknown command: $cmd"
            show_help
            return 1
            ;;
    esac
}

main "$@"
