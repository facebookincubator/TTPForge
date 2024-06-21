#!/bin/bash

set -e

TESTS_TO_RUN=$1
PROJECT=TTPFORGE
RETURN_CODE=0

TIMESTAMP=$(date +"%Y%m%d%H%M%S")
LOGFILE="/tmp/$PROJECT-unit-test-results-$TIMESTAMP.log"
MODULE_ROOT=$(go list -m -f "{{.Dir}}")

if [[ -z "${TESTS_TO_RUN}" ]]; then
    echo "No tests input" | tee -a "$LOGFILE"
    echo "Example - Run all shorter collection of tests: bash run-go-tests.sh short" | tee -a "$LOGFILE"
    echo "Example - Run all tests: bash run-go-tests.sh all" | tee -a "$LOGFILE"
    echo "Example - Run coverage for a specific version: bash run-go-tests.sh coverage" | tee -a "$LOGFILE"
    echo "Example - Run tests for modified files: bash run-go-tests.sh modified" | tee -a "$LOGFILE"
    exit 1
fi

run_tests() {
    local coverage_file=$1
    repo_root=$(git rev-parse --show-toplevel 2> /dev/null) || exit
    pushd "${repo_root}" || exit
    echo "Logging output to ${LOGFILE}" | tee -a "$LOGFILE"
    echo "Run the following command to see the output in real time:" | tee -a "$LOGFILE"
    echo "tail -f ${LOGFILE}" | tee -a "$LOGFILE"
    echo "Running tests..." | tee -a "$LOGFILE"

    # Check if go.mod and go.sum exist
    if [[ -f "go.mod" && -f "go.sum" ]]; then
        # Check if `go mod tidy` is necessary
        MOD_TMP=$(mktemp)
        SUM_TMP=$(mktemp)
        cp go.mod "$MOD_TMP"
        cp go.sum "$SUM_TMP"
        go mod tidy
        if ! cmp -s go.mod "$MOD_TMP" || ! cmp -s go.sum "$SUM_TMP"; then
            echo "Running 'go mod tidy' to clean up module dependencies..." | tee -a "$LOGFILE"
            go mod tidy 2>&1 | tee -a "$LOGFILE"
        fi
        rm "$MOD_TMP" "$SUM_TMP"
    fi

    if [[ "${TESTS_TO_RUN}" == 'coverage' ]]; then
        go test -v -race -failfast -tags=integration -coverprofile="${coverage_file}" ./... 2>&1 | tee -a "$LOGFILE"
    elif [[ "${TESTS_TO_RUN}" == 'all' ]]; then
        go test -v -race -failfast ./... 2>&1 | tee -a "$LOGFILE"
    elif [[ "${TESTS_TO_RUN}" == 'short' ]] && [[ "${GITHUB_ACTIONS}" != "true" ]]; then
        go test -v -short -failfast -race ./... 2>&1 | tee -a "$LOGFILE"
    elif [[ "${TESTS_TO_RUN}" == 'modified' ]]; then
        # Run tests for modified files
        local modified_files
        IFS=$'\n' read -r -a modified_files <<< "$(git diff --name-only --cached | grep '\.go$')"

        local pkg_dirs=()

        for file in "${modified_files[@]}"; do
            local pkg_dir
            pkg_dir=$(dirname "$file")
            pkg_dir=${pkg_dir#"$MODULE_ROOT/"}
            pkg_dirs+=("$pkg_dir")
        done

        # Remove duplicate package directories
        IFS=$'\n' read -r -a pkg_dirs <<< "$(sort -u <<< "${pkg_dirs[*]}")"
        unset IFS

        for dir in "${pkg_dirs[@]}"; do
            go test -v -race -failfast "./$dir/..." 2>&1 | tee -a "$LOGFILE"
        done
    else
        if [[ "${GITHUB_ACTIONS}" != 'true' ]]; then
            go test -v -failfast -race "./.../${TESTS_TO_RUN}" 2>&1 | tee -a "$LOGFILE"
        fi
    fi

    RETURN_CODE=$?
}

if [[ "${TESTS_TO_RUN}" == 'coverage' ]]; then
    run_tests 'coverage-all.out'
else
    run_tests
fi

if [[ "${RETURN_CODE}" -ne 0 ]]; then
    echo "unit tests failed" | tee -a "$LOGFILE"
    exit 1
fi
