#!/usr/bin/env bash

set -eu -o pipefail

if ! command -v gocritic &> /dev/null; then
    echo "gocritic not installed or available in the PATH" >&2
    echo "please check https://github.com/go-critic/go-critic" >&2
    exit 1
fi

subcommand=$1
shift

failed=false

for file in "$@"; do
    # redirect stderr so that violations and summaries are properly interleaved.
    if ! gocritic "$subcommand" "$file" 2>&1; then
        failed=true
    fi
done

if [[ $failed == "true" ]]; then
    exit 1
fi
