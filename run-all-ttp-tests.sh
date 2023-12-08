#!/bin/bash
set -e

# validate first argument
TTPFORGE_BINARY="$1"
if [ ! -f "${TTPFORGE_BINARY}" ]
then
  echo "Invalid TTPForge Binary Path Specified!"
  exit 1
fi
TTPFORGE_BINARY=$(realpath "${TTPFORGE_BINARY}")

# validate second argument
TTP_DIR="$2"
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
