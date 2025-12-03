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

# Usage
if [ $# -lt 2 ]; then
  echo "Usage: $0 <ttpforge_binary> <path> [--ignore item1 item2 ...]"
  echo ""
  echo "The --ignore flag accepts both directory names and file names."
  echo "Directories will be excluded from search, files will be skipped during execution."
  echo ""
  echo "Examples:"
  echo "  $0 ./ttpforge ./ttps"
  echo "  $0 ./ttpforge ./ttps --ignore tests broken"
  echo "  $0 ./ttpforge ./ttps --ignore legacy temp.yaml bad.yaml"
  exit 1
fi

TTPFORGE_BINARY="$1"
TTP_PATH="$2"
shift 2

# Parse optional arguments
EXCLUDES=()

while [ $# -gt 0 ]; do
  case "$1" in
    --ignore)
      shift
      ;;
    *)
      EXCLUDES+=("$1")
      shift
      ;;
  esac
done

# Validate binary
if [ ! -f "$TTPFORGE_BINARY" ]; then
  echo "Error: TTPForge binary not found: $TTPFORGE_BINARY"
  exit 1
fi

# Validate path
if [ ! -e "$TTP_PATH" ]; then
  echo "Error: Path not found: $TTP_PATH"
  exit 1
fi

# Build find command arguments with exclusions
FIND_ARGS=()
if [ ${#EXCLUDES[@]} -gt 0 ]; then
  FIND_ARGS+=(\()
  for i in "${!EXCLUDES[@]}"; do
    if [ $i -gt 0 ]; then
      FIND_ARGS+=(-o)
    fi
    FIND_ARGS+=(-path "*/${EXCLUDES[$i]}")
  done
  FIND_ARGS+=(\) -prune -o)
fi
FIND_ARGS+=(\( -type f -name '*.yaml' \) -print)

# Track results
FAILED_TTPS=()
PASSED_TTPS=()
TOTAL_COUNT=0

# Find and test all TTP files
while IFS= read -r ttp_file; do
  TOTAL_COUNT=$((TOTAL_COUNT + 1))
  echo "Testing: $ttp_file"

  if "$TTPFORGE_BINARY" validate "$ttp_file" --run-tests; then
    PASSED_TTPS+=("$ttp_file")
  else
    FAILED_TTPS+=("$ttp_file")
    echo "FAILED: $ttp_file"
  fi
  echo ""
done < <(find "$TTP_PATH" "${FIND_ARGS[@]}")

# Print summary
echo "========================================"
echo "TEST SUMMARY"
echo "========================================"
echo "Total TTPs tested: $TOTAL_COUNT"
echo "Passed: ${#PASSED_TTPS[@]}"
echo "Failed: ${#FAILED_TTPS[@]}"

if [ ${#FAILED_TTPS[@]} -gt 0 ]; then
  echo ""
  echo "Failed TTPs:"
  for ttp in "${FAILED_TTPS[@]}"; do
    echo "  - $ttp"
  done
  exit 1
fi

echo ""
echo "All TTPs passed!"
exit 0
