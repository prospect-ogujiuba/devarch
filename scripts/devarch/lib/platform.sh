#!/usr/bin/env bash
# Platform steps for DevArch bootstraps, delegated to the devarch CLI.
#
# Bootstraps start services, wait for readiness, and register hostnames through
# `devarch` so that logic lives in one tested place. This file has no other
# dependencies so bootstraps can source it next to dotenv.sh.

if [[ ${DEVARCH_PLATFORM_SH_LOADED:-} == 1 ]]; then
    return 0
fi
DEVARCH_PLATFORM_SH_LOADED=1

# Print the devarch executable: $DEVARCH_BIN (set by `devarch new`), else the
# one on PATH.
devarch_bin() {
    if [[ -n ${DEVARCH_BIN:-} ]]; then
        printf '%s\n' "$DEVARCH_BIN"
        return 0
    fi
    if command -v devarch >/dev/null 2>&1; then
        command -v devarch
        return 0
    fi
    printf 'devarch: the devarch CLI is required; install it with: (cd cli && go install ./cmd/devarch)\n' >&2
    return 1
}

_devarch_json_escape() {
    local s=$1
    s=${s//\\/\\\\}
    s=${s//\"/\\\"}
    s=${s//$'\n'/\\n}
    s=${s//$'\r'/\\r}
    s=${s//$'\t'/\\t}
    printf '%s' "${s//[[:cntrl:]]/}"
}

# devarch_progress STEP STATE [MESSAGE]
# STATE is start, done, warn, or fail. Writes one JSON line to the descriptor
# named by DEVARCH_PROGRESS_FD and does nothing when no reader provided one.
devarch_progress() {
    local fd=${DEVARCH_PROGRESS_FD:-}
    [[ $fd =~ ^[0-9]+$ ]] || return 0
    printf '{"step":"%s","state":"%s","message":"%s"}\n' \
        "$(_devarch_json_escape "$1")" "$(_devarch_json_escape "$2")" "$(_devarch_json_escape "${3:-}")" \
        2>/dev/null >&"$fd" || true
}

# devarch_runtime_env [DEVARCH]
# Sets DEVARCH_RUNTIME (podman or docker) and DEVARCH_CONTAINER_USER, the
# uid:gid that container exec calls use so bind-mounted files belong to the
# invoking user. `devarch new` passes both; a bootstrap run directly asks
# `DEVARCH config --env`. Without either, the runtime is podman. The user
# mapping mirrors cli/internal/engine/runtime.go: root inside a rootless Podman
# container is the invoking user, while Docker needs the user's own ids.
devarch_runtime_env() {
    local bin=${1:-} key value
    if [[ -z ${DEVARCH_RUNTIME:-} && -n $bin ]]; then
        while IFS='=' read -r key value; do
            case $key in
                DEVARCH_RUNTIME) DEVARCH_RUNTIME=$value ;;
                DEVARCH_CONTAINER_USER) DEVARCH_CONTAINER_USER=${DEVARCH_CONTAINER_USER:-$value} ;;
            esac
        done < <("$bin" config --env 2>/dev/null || true)
    fi
    DEVARCH_RUNTIME=${DEVARCH_RUNTIME:-podman}
    case $DEVARCH_RUNTIME in
        podman) DEVARCH_CONTAINER_USER=${DEVARCH_CONTAINER_USER:-0:0} ;;
        docker) DEVARCH_CONTAINER_USER=${DEVARCH_CONTAINER_USER:-$(id -u):$(id -g)} ;;
        *)
            printf 'devarch: DEVARCH_RUNTIME must be podman or docker, not %s\n' "$DEVARCH_RUNTIME" >&2
            return 1
            ;;
    esac
    [[ $DEVARCH_CONTAINER_USER =~ ^[0-9]+:[0-9]+$ ]] || {
        printf 'devarch: DEVARCH_CONTAINER_USER must be a numeric uid:gid, not %s\n' "$DEVARCH_CONTAINER_USER" >&2
        return 1
    }
}
