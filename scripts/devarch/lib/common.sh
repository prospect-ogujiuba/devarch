#!/usr/bin/env bash
# Shared DevArch command helpers. This library intentionally does not abstract
# container runtimes or individual Podman subcommands.

if [[ ${DEVARCH_COMMON_SH_LOADED:-} == 1 ]]; then
    return 0
fi
readonly DEVARCH_COMMON_SH_LOADED=1

devarch_die() {
    printf 'devarch: %s\n' "$*" >&2
    return 1
}

_devarch_real_dir() {
    local directory=${1:?directory is required}
    [[ -d $directory ]] || devarch_die "directory does not exist: $directory" || return
    (cd -P -- "$directory" && pwd -P)
}

# Print the physical repository root. DEVARCH_REPO_ROOT is supported for tests
# and callers operating on another DevArch checkout.
devarch_repo_root() {
    local candidate library_dir root

    if [[ -n ${DEVARCH_REPO_ROOT:-} ]]; then
        candidate=$DEVARCH_REPO_ROOT
    else
        library_dir=$(_devarch_real_dir "$(dirname -- "${BASH_SOURCE[0]}")") || return
        candidate=$library_dir/../../..
    fi

    root=$(_devarch_real_dir "$candidate") || return
    [[ -d $root/services-library ]] ||
        devarch_die "not a DevArch repository (services-library is missing): $root" || return
    printf '%s\n' "$root"
}

# Fail with actionable installation guidance. There is deliberately no Docker
# fallback.
devarch_require_podman() {
    command -v podman >/dev/null 2>&1 ||
        devarch_die "podman is required; install Podman and ensure 'podman' is on PATH"
}

# Keep capability probes quiet when they succeed, but replay both native output
# channels unchanged when they fail.
_devarch_quiet_success() {
    local output_directory status

    output_directory=$(mktemp -d "${TMPDIR:-/tmp}/devarch-check.XXXXXX") ||
        devarch_die "could not create temporary command output directory" || return
    if command "$@" >"$output_directory/stdout" 2>"$output_directory/stderr"; then
        rm -rf -- "$output_directory"
        return 0
    else
        status=$?
    fi

    cat "$output_directory/stdout" || true
    cat "$output_directory/stderr" >&2 || true
    rm -rf -- "$output_directory"
    return "$status"
}

# Let Podman select and diagnose its configured external Compose provider.
devarch_require_compose() {
    devarch_require_podman || return
    _devarch_quiet_success podman compose --help
}

# Log an argv vector without evaluating it, then run it unchanged. Pass --exec
# for a final native process when no post-processing is required.
devarch_run() {
    local use_exec=0 monitor_mode=0 argument child_pid status signal
    local received_signal=''
    local -a argv=()
    local -A previous_traps=()

    if [[ ${1:-} == --exec ]]; then
        use_exec=1
        shift
    fi
    [[ ${1:-} == -- ]] || devarch_die "devarch_run usage: devarch_run [--exec] -- command [args...]" || return
    shift
    (($# > 0)) || devarch_die "devarch_run requires a command" || return
    argv=("$@")

    printf '+' >&2
    for argument in "${argv[@]}"; do
        printf ' %q' "$argument" >&2
    done
    printf '\n' >&2

    if ((use_exec)); then
        exec "${argv[@]}"
    fi

    for signal in HUP INT QUIT TERM; do
        previous_traps[$signal]=$(trap -p "$signal")
    done

    if [[ $- == *m* ]]; then
        monitor_mode=1
        set +m
    fi
    command "${argv[@]}" <&0 &
    child_pid=$!
    trap 'received_signal=HUP; kill -HUP "$child_pid" 2>/dev/null || true' HUP
    trap 'received_signal=INT; kill -INT "$child_pid" 2>/dev/null || true' INT
    trap 'received_signal=QUIT; kill -QUIT "$child_pid" 2>/dev/null || true' QUIT
    trap 'received_signal=TERM; kill -TERM "$child_pid" 2>/dev/null || true' TERM

    while :; do
        if wait "$child_pid"; then
            status=0
        else
            status=$?
        fi
        kill -0 "$child_pid" 2>/dev/null || break
    done

    if ((monitor_mode)); then
        set -m
    fi
    for signal in HUP INT QUIT TERM; do
        if [[ -n ${previous_traps[$signal]} ]]; then
            eval "${previous_traps[$signal]}"
        else
            trap - "$signal"
        fi
    done

    if [[ -n $received_signal ]]; then
        kill -s "$received_signal" "$BASHPID"
    fi
    return "$status"
}
