#!/bin/sh

set -e

TOOL_NAME="$1"
TOOL_BINARY="$2"
DOWNLOAD_BASE_URL="$3"
DOWNLOAD_BINARY_FILE="$4"
DOWNLOAD_CHECKSUM_FILE="$5"
SHASUM_IGNORE_MISSING="$6"
SHASUM_ALGO="$7"
TEST_COMMAND="$8"

echo "Installing $TOOL_NAME..."

# Download checksums and binary
curl -L "${DOWNLOAD_BASE_URL}/${DOWNLOAD_CHECKSUM_FILE}" -o checksums.txt
curl -OL "${DOWNLOAD_BASE_URL}/${DOWNLOAD_BINARY_FILE}"

# Verify checksums
shasum "${SHASUM_IGNORE_MISSING}" -a "${SHASUM_ALGO}" -c checksums.txt

if [ "${TOOL_BINARY}" = "terragrunt" ]; then
    chmod +x "${DOWNLOAD_BINARY_FILE}"
    mv "${DOWNLOAD_BINARY_FILE}" /usr/local/bin/"${TOOL_BINARY}"
elif [ "${DOWNLOAD_BINARY_FILE##*.}" = "deb" ]; then
    dpkg -i "${DOWNLOAD_BINARY_FILE}"
elif [ "${DOWNLOAD_BINARY_FILE##*.}" = "gz" ]; then
    tar xzf "${DOWNLOAD_BINARY_FILE}"
    mv "${TOOL_BINARY}" /usr/local/bin/
else
    chmod +x "${DOWNLOAD_BINARY_FILE}"
    mv "${DOWNLOAD_BINARY_FILE}" /usr/local/bin/
fi

# Clean up
rm -f checksums.txt "${DOWNLOAD_BINARY_FILE}"

# Verify binary works
eval "${TEST_COMMAND}"

echo "$TOOL_NAME installation complete."
