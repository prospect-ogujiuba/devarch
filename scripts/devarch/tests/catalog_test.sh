#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
# shellcheck source-path=SCRIPTDIR
# shellcheck source=test-helper.sh
source "$SCRIPT_DIR/test-helper.sh"
# shellcheck disable=SC1091 # Resolved from the test helper's physical path.
source "$DEVARCH_DIR/lib/catalog.sh"

new_fixture_repo ordered
add_fixture_service zeta shared
add_fixture_service alpha unique
add_fixture_service alpha shared

catalog=$(devarch_catalog_list)
assert_eq $'alpha/shared\nalpha/unique\nzeta/shared' "$catalog" "catalog IDs must be deterministic"
assert_eq "zeta/shared" "$(devarch_catalog_resolve zeta/shared)" "canonical ID must resolve exactly"
assert_eq "alpha/unique" "$(devarch_catalog_resolve unique)" "unique short ID must resolve"
assert_eq "$FIXTURE_REPO/services-library/alpha/unique/compose.yml" \
    "$(devarch_catalog_compose_file unique)" "Compose path must be validated and absolute"

error_file=$TEST_TMP/ambiguous-error
if devarch_catalog_resolve shared 2>"$error_file"; then
    fail "ambiguous short ID unexpectedly resolved"
fi
ambiguous_error=$(<"$error_file")
assert_contains "$ambiguous_error" "alpha/shared" "ambiguous error must list first candidate"
assert_contains "$ambiguous_error" "zeta/shared" "ambiguous error must list second candidate"

new_fixture_repo missing
mkdir -p "$FIXTURE_REPO/services-library/data/redis"
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted a missing Compose file"
fi
assert_contains "$(<"$error_file")" "Compose file is missing" "missing Compose diagnostic"

new_fixture_repo malformed
add_fixture_service data empty ""
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted an empty Compose file"
fi
assert_contains "$(<"$error_file")" "unreadable or empty" "empty Compose diagnostic"

new_fixture_repo malformed-structure
add_fixture_service data broken $'name: not-compose\n'
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted a malformed Compose file"
fi
assert_contains "$(<"$error_file")" "malformed Compose file" "malformed Compose diagnostic"

new_fixture_repo valid-anchor
add_fixture_service data anchored $'services: &service-map\n  app:\n    image: example.invalid/test:latest\n'
assert_eq "data/anchored" "$(devarch_catalog_list)" \
    "catalog must accept an anchored services mapping"

new_fixture_repo malformed-yaml
add_fixture_service data broken $'services:\n  app: [\n'
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted malformed nested YAML"
fi
assert_contains "$(<"$error_file")" "malformed Compose file" "malformed YAML diagnostic"

new_fixture_repo escaped-service
outside=$TEST_TMP/outside-service
mkdir -p "$outside"
printf 'services: {}\n' >"$outside/compose.yml"
mkdir -p "$FIXTURE_REPO/services-library/data"
ln -s "$outside" "$FIXTURE_REPO/services-library/data/escaped"
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted a service path outside the repository"
fi
assert_contains "$(<"$error_file")" "escapes the repository" "escaped service diagnostic"

new_fixture_repo escaped-category
add_fixture_service safe service
outside_category=$TEST_TMP/outside-category
mkdir -p "$outside_category"
ln -s "$outside_category" "$FIXTURE_REPO/services-library/escaped"
if devarch_catalog_list >/dev/null 2>"$error_file"; then
    fail "catalog accepted a category path outside the repository"
fi
assert_contains "$(<"$error_file")" "category escapes the repository" "escaped category diagnostic"

new_fixture_repo escaped-catalog
outside_catalog=$TEST_TMP/outside-catalog
mkdir -p "$outside_catalog/data/service"
printf 'services: {}\n' >"$outside_catalog/data/service/compose.yml"
rmdir "$FIXTURE_REPO/services-library"
ln -s "$outside_catalog" "$FIXTURE_REPO/services-library"
if devarch_catalog_root >/dev/null 2>"$error_file"; then
    fail "catalog accepted services-library outside the repository"
fi
assert_contains "$(<"$error_file")" "services-library escapes" "escaped catalog diagnostic"

printf 'ok - catalog discovery and resolution\n'
