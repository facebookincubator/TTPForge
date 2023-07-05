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
set -e

# Define the URL of bashutils.sh
bashutils_url="https://raw.githubusercontent.com/l50/dotfiles/main/bashutils"

# Define the local path of bashutils.sh
bashutils_path="/tmp/bashutils"

# Check if bashutils.sh exists locally
if [[ ! -f "${bashutils_path}" ]]; then
    # bashutils.sh doesn't exist locally, so download it
    mkdir -p "${HOME}/.dotfiles"
    curl -s "${bashutils_url}" -o "${bashutils_path}"
fi

# Source bashutils
# shellcheck source=/dev/null
source "${bashutils_path}"

if command -v mage &> /dev/null; then
    echo "mage is installed"
else
    echo -e "mage is not installed\n"
    echo -e "Please install mage by running the following command:\n"
    echo -e "go install github.com/magefile/mage@latest\n"
    exit 1
fi

repo_root

# Run the mage generatepackagedocs command
$(which mage) generatepackagedocs

# Catch the exit code of the last command
exit_status=$?

# If the exit code is not zero (i.e., the command failed),
# then stop the commit and show an error message
if [ $exit_status -ne 0 ]; then
    echo "failed to generate package docs"
    exit 1
fi
