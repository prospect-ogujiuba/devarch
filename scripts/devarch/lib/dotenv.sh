#!/usr/bin/env bash

# Load selected dotenv assignments as data. Values are assigned in the caller's
# shell without changing their export attribute.

_devarch_dotenv_error() {
    printf 'dotenv: %s\n' "$1" >&2
    return 1
}

_devarch_dotenv_trim() {
    local value=$1
    value="${value#"${value%%[![:space:]]*}"}"
    value="${value%"${value##*[![:space:]]}"}"
    printf '%s' "$value"
}

_devarch_dotenv_parse_value() {
    local key=$1 raw quote char remainder value='' escaped=false closed=false i
    raw=$(_devarch_dotenv_trim "$2")

    if [[ $raw == \"* || $raw == \'* ]]; then
        quote=${raw:0:1}
        for ((i = 1; i < ${#raw}; i++)); do
            char=${raw:i:1}
            if [[ $escaped == true ]]; then
                if [[ $char == \\ || $char == "$quote" || ( $quote == '"' && $char == '$' ) ]]; then
                    value+=$char
                else
                    _devarch_dotenv_error "$key contains an unsupported quoted escape"
                    return 1
                fi
                escaped=false
            elif [[ $char == \\ ]]; then
                escaped=true
            elif [[ $char == "$quote" ]]; then
                closed=true
                i=$((i + 1))
                break
            else
                value+=$char
            fi
        done
        if [[ $closed != true ]]; then
            _devarch_dotenv_error "$key has an unterminated quoted value"
            return 1
        fi
        remainder=$(_devarch_dotenv_trim "${raw:i}")
        if [[ -n $remainder && $remainder != \#* ]]; then
            _devarch_dotenv_error "$key has unexpected text after its quoted value"
            return 1
        fi
    else
        for ((i = 1; i < ${#raw}; i++)); do
            if [[ ${raw:i:1} == '#' && ${raw:i-1:1} == [[:space:]] ]]; then
                raw=${raw:0:i}
                break
            fi
        done
        value=$(_devarch_dotenv_trim "$raw")
    fi

    DEVARCH_DOTENV_VALUE=$value
}

devarch_load_dotenv() {
    if (($# < 2)); then
        _devarch_dotenv_error 'loader requires a file and at least one allowed key'
        return 1
    fi

    local file=$1
    shift
    if [[ ! -f $file ]]; then
        _devarch_dotenv_error "env file is not a regular file: $file"
        return 1
    fi

    local allowed_key key line candidate DEVARCH_DOTENV_VALUE=''
    local -a allowed_keys=("$@")
    for key in "${allowed_keys[@]}"; do
        if [[ ! $key =~ ^[A-Z][A-Z0-9_]*$ ]]; then
            _devarch_dotenv_error "invalid allowlist key: $key"
            return 1
        fi

        # Bash strings cannot retain NUL. Inspect only lines belonging to this
        # accepted key before reading them into shell variables.
        if LC_ALL=C grep -aE "^[[:space:]]*(export[[:space:]]+)?${key}[[:space:]]*=" -- "$file" \
            | LC_ALL=C grep -a '[[:cntrl:]]' >/dev/null; then
            _devarch_dotenv_error "$key contains an ASCII control byte"
            return 1
        fi
    done

    while IFS= read -r line || [[ -n $line ]]; do
        line=$(_devarch_dotenv_trim "$line")
        [[ -z $line || $line == \#* ]] && continue
        if [[ $line == export[[:space:]]* ]]; then
            line=$(_devarch_dotenv_trim "${line#export}")
        fi
        [[ $line == *=* ]] || continue
        candidate=$(_devarch_dotenv_trim "${line%%=*}")

        key=''
        for allowed_key in "${allowed_keys[@]}"; do
            if [[ $candidate == "$allowed_key" ]]; then
                key=$allowed_key
                break
            fi
        done
        [[ -n $key ]] || continue

        # The byte preflight above covers NUL, CR, tab, DEL, and other ASCII
        # controls without ever placing the assignment's contents in an error.
        _devarch_dotenv_parse_value "$key" "${line#*=}" || return 1
        printf -v "$key" '%s' "$DEVARCH_DOTENV_VALUE"
    done <"$file"
}
