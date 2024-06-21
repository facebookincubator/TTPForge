#!/bin/bash
set -e

pkgs=$(go list ./...)

for pkg in $pkgs; do
    dir="$(basename "$pkg")/"
    if [[ "${dir}" != .*/ ]]; then
        go vet "${pkg}"
    fi
done
