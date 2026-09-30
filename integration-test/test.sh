#!/bin/bash

########################################################
# Integration test entry
#
# Usage:
#   ./test.sh <TestName>   run one Test* function, e.g. ./test.sh TestLogin
#   ./test.sh TestAll      run every Test* function
#   ./test.sh list         list the Test* functions
#
# Each Test* runs against a fresh compose of the image built by
# `make dockertest`: the compose is started before it and removed after it,
# and its db lives under $WT_TEST_ROOT/<TestName> (default /tmp).
#
# Environment:
#   WT_TEST_IMAGE  image to test          (default alonza0314/work-tracker:test)
#   WT_TEST_PORT   host port of the app   (default 18888)
#   WT_TEST_ROOT   where per-case data go (default /tmp/wt-integration-test)
########################################################

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO_TEST_DIR="${SCRIPT_DIR}/goTest"

export WT_TEST_IMAGE="${WT_TEST_IMAGE:-alonza0314/work-tracker:test}"
export WT_TEST_PORT="${WT_TEST_PORT:-18888}"
export WT_TEST_UID="$(id -u)"
export WT_TEST_GID="$(id -g)"
WT_TEST_ROOT="${WT_TEST_ROOT:-/tmp/wt-integration-test}"

COMPOSE=(docker compose -f "${SCRIPT_DIR}/docker-compose.yaml")

usage() {
    echo "Usage: $0 <TestName|TestAll|list>"
    echo "Test cases:"
    list_cases | sed 's/^/  /'
}

# every top-level Test* function of goTest, in file order
list_cases() {
    grep -h -o '^func Test[A-Za-z0-9_]*(t \*testing.T)' "${GO_TEST_DIR}"/*_test.go \
        | sed -E 's/^func (Test[A-Za-z0-9_]*).*/\1/'
}

compose_down() {
    "${COMPOSE[@]}" down --volumes --remove-orphans >/dev/null 2>&1
}

wait_ready() {
    local url="http://localhost:${WT_TEST_PORT}/login"
    for _ in $(seq 1 60); do
        if curl -sf -o /dev/null "${url}"; then
            return 0
        fi
        sleep 0.5
    done
    echo "the app did not answer on ${url}"
    return 1
}

run_case() {
    local name=$1
    export WT_TEST_DATA="${WT_TEST_ROOT}/${name}"

    echo "================ ${name} ================"
    compose_down
    rm -rf "${WT_TEST_DATA}"
    mkdir -p "${WT_TEST_DATA}"

    if ! "${COMPOSE[@]}" up -d >/dev/null 2>&1 || ! wait_ready; then
        echo "failed to start the compose for ${name}"
        "${COMPOSE[@]}" logs --no-color | tail -30
        compose_down
        return 1
    fi

    local result=0
    (cd "${GO_TEST_DIR}" && WT_BASE_URL="http://localhost:${WT_TEST_PORT}" \
        go test -count=1 -v -run "^${name}\$" ./...) || result=1

    if [ ${result} -ne 0 ]; then
        echo "---------------- ${name}: app log ----------------"
        "${COMPOSE[@]}" logs --no-color | tail -50
    fi
    compose_down
    rm -rf "${WT_TEST_DATA}"
    return ${result}
}

main() {
    if [ $# -ne 1 ]; then
        usage
        return 1
    fi
    if [ "$1" = "list" ]; then
        list_cases
        return 0
    fi
    if ! docker image inspect "${WT_TEST_IMAGE}" >/dev/null 2>&1; then
        echo "image ${WT_TEST_IMAGE} not found; build it with: make dockertest"
        return 1
    fi

    local cases
    if [ "$1" = "TestAll" ]; then
        cases=$(list_cases)
    elif list_cases | grep -qx "$1"; then
        cases=$1
    else
        echo "unknown test case: $1"
        usage
        return 1
    fi

    trap compose_down EXIT
    local passed=() failed=()
    for name in ${cases}; do
        if run_case "${name}"; then
            passed+=("${name}")
        else
            failed+=("${name}")
        fi
    done

    echo "================ summary ================"
    for name in "${passed[@]}"; do echo "PASS ${name}"; done
    for name in "${failed[@]}"; do echo "FAIL ${name}"; done
    echo "${#passed[@]} passed, ${#failed[@]} failed"
    [ ${#failed[@]} -eq 0 ]
}

main "$@"
