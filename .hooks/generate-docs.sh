#!/bin/bash
#
# This script is a pre-commit hook that checks if the mage command is
# installed and if not, prompts the user to install it. If mage is
# installed, the script changes to the repository root and runs the
# `mage generatepackagedocs` command. This command generates documentation
# for all Go packages in the current directory and its subdirectories by
# traversing the file tree and creating a new README.md file in each
# directory containing a Go package. If the command fails, the commit
# is stopped and an error message is shown.

# Define the lock file
lockfile="/tmp/mage_generatepackagedocs.lock"

# Check if lock file exists, exit if true
if [ -f "${lockfile}" ]; then
    echo "Another instance of this script is running. Exiting."
    exit 1
fi

# Create the lock file
touch "${lockfile}"

# Trap to remove the lock file in case of termination or exit
trap 'rm -f '"${lockfile}" EXIT

set -e

# Define the URL of bashutils.sh
bashutils_url="https://raw.githubusercontent.com/l50/dotfiles/main/bashutils"

# Define the local path of bashutils.sh
bashutils_path="/tmp/bashutils"

# Check if bashutils.sh exists locally
if [[ ! -f "${bashutils_path}" ]]; then
    # bashutils.sh doesn't exist locally, so download it
    curl -s "${bashutils_url}" -o "${bashutils_path}"
fi

# Source bashutils
# shellcheck source=/dev/null
source "${bashutils_path}"

repo_root

mage_bin=$(go env GOPATH)/bin/mage

# Check if mage is installed
if [[ -x "${mage_bin}" ]]; then
    echo "mage is installed"
else
    echo -e "mage is not installed\n"
    echo -e "Please install mage by running the following command:\n"
    echo -e "go install github.com/magefile/mage@latest\n"
    exit 1
fi

# Run the mage generatepackagedocs command
"${mage_bin}" generatepackagedocs
# Catch the exit code of the last command
exit_status=$?

# If the exit code is not zero (i.e., the command failed),
# then stop the commit and show an error message
if [ $exit_status -ne 0 ]; then
    echo "failed to generate package docs"
    exit 1
fi
