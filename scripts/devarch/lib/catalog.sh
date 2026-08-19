#!/usr/bin/env bash
# DevArch service catalog discovery and canonical ID resolution.

if [[ ${DEVARCH_CATALOG_SH_LOADED:-} == 1 ]]; then
    return 0
fi
readonly DEVARCH_CATALOG_SH_LOADED=1

_devarch_catalog_library_dir=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P) || return
# shellcheck source-path=SCRIPTDIR
# shellcheck source=common.sh
source "$_devarch_catalog_library_dir/common.sh"
unset _devarch_catalog_library_dir

_devarch_catalog_valid_segment() {
    [[ $1 =~ ^[a-z0-9][a-z0-9._-]*$ ]]
}

_devarch_catalog_path_is_within() {
    local path=$1 parent=$2
    [[ $path == "$parent" || $path == "$parent"/* ]]
}

# Print the physical services-library path after proving it remains inside the
# physical repository root.
devarch_catalog_root() {
    local repository catalog
    repository=$(devarch_repo_root) || return
    catalog=$(_devarch_real_dir "$repository/services-library") || return
    _devarch_catalog_path_is_within "$catalog" "$repository" ||
        devarch_die "services-library escapes the repository: $repository/services-library" || return
    printf '%s\n' "$catalog"
}

_devarch_catalog_validate_category() {
    local catalog=$1 category=$2 category_path

    _devarch_catalog_valid_segment "$category" ||
        devarch_die "malformed catalog category: $category" || return
    category_path=$catalog/$category
    [[ -d $category_path ]] || devarch_die "catalog category is missing: $category" || return
    [[ ! -L $category_path ]] ||
        devarch_die "catalog category escapes the repository: $category" || return
}

_devarch_catalog_validate_service_path() {
    local catalog=$1 category=$2 name=$3 service compose

    _devarch_catalog_validate_category "$catalog" "$category" || return
    _devarch_catalog_valid_segment "$name" ||
        devarch_die "malformed catalog service name: $name" || return

    service=$catalog/$category/$name
    compose=$service/compose.yml
    [[ -d $service ]] || devarch_die "service directory is missing: $category/$name" || return
    [[ ! -L $service ]] ||
        devarch_die "service path escapes the repository: $category/$name" || return
    [[ ! -L $compose ]] || devarch_die "Compose file may not be a symbolic link: $compose" || return
    [[ -f $compose ]] || devarch_die "Compose file is missing: $compose" || return
    [[ -r $compose && -s $compose ]] || devarch_die "Compose file is unreadable or empty: $compose" || return
}

_devarch_catalog_has_services_key() {
    local compose=$1 line
    local services_key_regex="^(services|'services'|\"services\")[[:space:]]*:"

    while IFS= read -r line || [[ -n $line ]]; do
        [[ $line =~ $services_key_regex ]] && return 0
    done <"$compose"
    return 1
}

_devarch_catalog_validate_compose_native() {
    local compose=$1

    _devarch_quiet_success podman compose -f "$compose" config --quiet || {
        devarch_die "malformed Compose file: $compose"
        return
    }
    _devarch_catalog_has_services_key "$compose" ||
        devarch_die "malformed Compose file (top-level services mapping is missing): $compose" || return
}

# A compatible yq can parse the full catalog in one process. Fall back to the
# configured Compose provider for portability and for focused diagnostics.
_devarch_catalog_validate_composes() {
    local compose
    local yaml_filter='all(.[]; type == "object" and (.services | type == "object"))'

    devarch_require_podman || return
    if command -v yq >/dev/null 2>&1 &&
        yq -s -e "$yaml_filter" "$@" >/dev/null 2>&1; then
        return 0
    fi

    for compose in "$@"; do
        _devarch_catalog_validate_compose_native "$compose" || return
    done
}

_devarch_catalog_validate_service() {
    local catalog=$1 category=$2 name=$3

    _devarch_catalog_validate_service_path "$catalog" "$category" "$name" || return
    _devarch_catalog_validate_composes "$catalog/$category/$name/compose.yml"
}

# Emit every canonical category/name ID in bytewise deterministic order.
devarch_catalog_list() {
    local catalog category_path service_path category name
    local -a ids=() compose_files=()

    catalog=$(devarch_catalog_root) || return
    for category_path in "$catalog"/*; do
        [[ -d $category_path ]] || continue
        category=${category_path##*/}
        _devarch_catalog_validate_category "$catalog" "$category" || return
        for service_path in "$category_path"/*; do
            [[ -d $service_path ]] || continue
            name=${service_path##*/}
            _devarch_catalog_validate_service_path "$catalog" "$category" "$name" || return
            ids+=("$category/$name")
            compose_files+=("$service_path/compose.yml")
        done
    done

    ((${#ids[@]} > 0)) || devarch_die "no services found under $catalog" || return
    _devarch_catalog_validate_composes "${compose_files[@]}" || return
    printf '%s\n' "${ids[@]}" | LC_ALL=C sort
}

# Resolve an exact canonical ID or a unique short service name. Exact IDs are
# never broadened to fuzzy matches.
devarch_catalog_resolve() {
    local query=${1:-} catalog category name id category_path service_path
    local LC_ALL=C
    local -a matches=()

    [[ -n $query ]] || devarch_die "a service ID is required" || return
    catalog=$(devarch_catalog_root) || return

    if [[ $query == */* ]]; then
        [[ $query != */*/* ]] || devarch_die "invalid canonical service ID: $query" || return
        category=${query%%/*}
        name=${query#*/}
        _devarch_catalog_validate_service "$catalog" "$category" "$name" || return
        printf '%s\n' "$query"
        return
    fi

    _devarch_catalog_valid_segment "$query" || devarch_die "invalid service name: $query" || return
    for category_path in "$catalog"/*; do
        [[ -d $category_path ]] || continue
        category=${category_path##*/}
        _devarch_catalog_validate_category "$catalog" "$category" || return
        for service_path in "$category_path"/*; do
            [[ -d $service_path ]] || continue
            name=${service_path##*/}
            _devarch_catalog_validate_service_path "$catalog" "$category" "$name" || return
            [[ $name == "$query" ]] && matches+=("$category/$name")
        done
    done

    for id in "${matches[@]}"; do
        _devarch_catalog_validate_composes "$catalog/$id/compose.yml" || return
    done

    case ${#matches[@]} in
        1)
            printf '%s\n' "${matches[0]}"
            ;;
        0)
            devarch_die "unknown service: $query"
            ;;
        *)
            printf 'devarch: ambiguous service name %q; candidates:\n' "$query" >&2
            printf '  %s\n' "${matches[@]}" >&2
            return 1
            ;;
    esac
}

# Print the validated Compose file for an exact or unique service ID.
devarch_catalog_compose_file() {
    local id catalog
    id=$(devarch_catalog_resolve "${1:-}") || return
    catalog=$(devarch_catalog_root) || return
    printf '%s/%s/compose.yml\n' "$catalog" "$id"
}
