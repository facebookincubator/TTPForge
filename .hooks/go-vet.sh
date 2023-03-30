#!/bin/bash
set -ex

pkg=$(go list .)
for dir in */; do
    if [[ "${dir}" != ".hooks/" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "cmd/" ]] \
                              && [[ "${dir}" != "config/" ]] \
                              && [[ "${dir}" != "deployments/" ]] \
                              && [[ "${dir}" != "files/" ]] \
                              && [[ "${dir}" != "images/" ]] \
                              && [[ "${dir}" != "logs/" ]] \
                              && [[ "${dir}" != "magefiles/" ]] \
                              && [[ "${dir}" != "modules/" ]] \
                              && [[ "${dir}" != "resources/" ]] \
                              && [[ "${dir}" != "templates/" ]]; then
        go vet "${pkg}/${dir}"
    fi
done
