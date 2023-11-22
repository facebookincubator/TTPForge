#!/bin/bash
set -e

cd "$(dirname "$0")"
"$1" test example-ttps/**/*.yaml
