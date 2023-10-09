#!/bin/bash

set -e  # Stop on any error

echo "Starting the test of magefile commands..."

# Compile for all supported OS and architectures
echo "Testing 'Compile' for all supported OS and architectures..."
mage Compile true

# Compile for specific OS and architectures
echo "Testing 'Compile' for macOS on arm64..."
GOOS=darwin GOARCH=arm64 mage Compile false

echo "Testing 'Compile' for Linux on amd64..."
GOOS=linux GOARCH=amd64 mage Compile false

echo "Testing 'Compile' for Windows on amd64..."
GOOS=windows GOARCH=amd64 mage Compile false

# Install Dependencies
echo "Testing 'InstallDeps'..."
mage InstallDeps

# Generate Package Docs
echo "Testing 'GeneratePackageDocs'..."
mage GeneratePackageDocs

# Run PreCommit
echo "Testing 'RunPreCommit'..."
mage RunPreCommit

# Run Tests
echo "Testing 'RunTests'..."
mage RunTests

# Run Integration Tests
echo "Testing 'RunIntegrationTests'..."
mage RunIntegrationTests

echo "All magefile commands executed successfully!"
