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

EXCEPTIONS_FILE=["kill-process-windows.yaml","kill-process-windows-failure.yaml"]

# Loop over all specified directories and validate all ttps within each.
shift
for TTP_DIR in "$@"; do
  # validate directory
  if [ ! -d "${TTP_DIR}" ]
  then
    echo "Invalid TTP Directory Specified!"
    exit 1
  fi
  TTP_DIR=$(realpath "${TTP_DIR}")

  TTP_FILE_LIST="$(find "${TTP_DIR}" -name "*.yaml")"
  for TTP_FILE in ${TTP_FILE_LIST}
  do
      echo "Running TTP: ${TTP_FILE}"
      if [[ "${EXCEPTIONS_FILE[*]}" =~ ${TTP_FILE##*/} ]]; then
        echo "Skipping TTP: ${TTP_FILE}"
        continue
      fi
      ${TTPFORGE_BINARY} test "${TTP_FILE}"
  done
done
