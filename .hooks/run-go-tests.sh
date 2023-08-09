#!/bin/bash

set -e

TESTS_TO_RUN=$1
RETURN_CODE=0

if [[ -z "${TESTS_TO_RUN}" ]]; then
    echo "No tests input"
    echo "Example - Run all tests: bash run-go-tests.sh all"
    echo "Example - Run all tests and generate coverage report: bash run-go-tests.sh coverage"
    exit 1
fi

if [[ "${TESTS_TO_RUN}" == 'coverage' ]]; then
    go test -v -race -failfast \
        -tags=integration -coverprofile=coverage-all.out ./...
    RETURN_CODE=$?
elif [[ "${TESTS_TO_RUN}" == 'all' ]]; then
    go test -v -race -failfast ./...
    RETURN_CODE=$?
elif [[ "${TESTS_TO_RUN}" == 'short' ]] \
                                        && [[ "${GITHUB_ACTIONS}" != "true" ]]; then
    go test -v -short -failfast -race ./...
    RETURN_CODE=$?
else
    if [[ "${GITHUB_ACTIONS}" != 'true' ]]; then
        go test -v -race -failfast "./.../${TESTS_TO_RUN}"
        RETURN_CODE=$?
    fi
fi

if [[ "${RETURN_CODE}" -ne 0 ]]; then
    echo "unit tests failed"
    exit 1
fi
