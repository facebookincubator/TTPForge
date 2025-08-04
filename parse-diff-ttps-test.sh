#!/bin/bash

# Copyright © 2023-present, Meta Platforms, Inc. and affiliates
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

set -e

# Validate path to TTPForge binary
TTPFORGE_BINARY="$1"
if [ ! -f "${TTPFORGE_BINARY}" ]
then
  echo "Invalid TTPForge Binary Path Specified!"
  exit 1
fi
TTPFORGE_BINARY=$(realpath "${TTPFORGE_BINARY}")

TTP_BASE_DIR="$2"

shift 2

full_paths=()
for file in "$@"; do
  echo "Processing input: $file"
  ttp_ref="${TTP_BASE_DIR}/${file}"
  echo "TTP Path: $ttp_ref"
  full_paths+=("$ttp_ref")
done

echo "Processing following TTPs:" "${full_paths[@]}"

${TTPFORGE_BINARY} parse-yaml "${full_paths[@]}"
