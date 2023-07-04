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
set -ex

# Change to the repository root
cd "$(git rev-parse --show-toplevel)"

# Determine the location of the mage binary
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

# repo_root

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
