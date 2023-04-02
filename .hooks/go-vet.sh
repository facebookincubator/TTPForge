#!/bin/bash
set -ex

pkgs=$(go list ./...)

for pkg in $pkgs; do
    dir="$(basename "$pkg")/"
    if [[ "${dir}" != ".hooks/" ]] \
                              && [[ "${dir}" != ".github/" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "docs/" ]] \
                              && [[ "${dir}" != "logging/" ]] \
                              && [[ "${dir}" != "magefiles/" ]] \
                              && [[ "${dir}" != "modules/" ]] \
                              && [[ "${dir}" != "resources/" ]] \
                              && [[ "${dir}" != "templates/" ]]; then
        go vet "${pkg}"
    fi
done
