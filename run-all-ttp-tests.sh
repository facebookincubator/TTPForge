#!/bin/bash
set -e

# Validate path to TTPForge binary
TTPFORGE_BINARY="$1"
if [ ! -f "${TTPFORGE_BINARY}" ]
then
  echo "Invalid TTPForge Binary Path Specified!"
  exit 1
fi
TTPFORGE_BINARY=$(realpath "${TTPFORGE_BINARY}")

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
      ${TTPFORGE_BINARY} test "${TTP_FILE}"
  done
done
