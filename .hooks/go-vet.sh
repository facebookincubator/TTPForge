#!/bin/bash
set -ex

pkg=$(go list ./...)
for dir in */; do
    if [[ "${dir}" != ".hooks/" ]] \
                              && [[ "${dir}" != ".github" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "docs/" ]] \
                              && [[ "${dir}" != "logging/" ]] \
                              && [[ "${dir}" != "magefiles/" ]] \
                              && [[ "${dir}" != "modules/" ]] \
                              && [[ "${dir}" != "resources/" ]] \
                              && [[ "${dir}" != "templates/" ]]; then
        go vet "${pkg}/${dir}"
    fi
done
