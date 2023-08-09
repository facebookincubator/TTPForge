#!/bin/bash

REPO=$(grep "url" .git/config)

if [[ "${REPO}" == *https* ]]; then
    echo "${REPO}" | awk -F '://' '{print $2}' | awk -F '.git' '{print $1}'
else
    echo "${REPO}" | awk -F '@' '{print $2}' | tr : / | awk -F '.git' '{print $1}'
fi

if grep "replace ${REPO}" "$@" 2>&1; then
    echo "ERROR: Don't commit a replacement in go.mod!"
    exit 1
fi
