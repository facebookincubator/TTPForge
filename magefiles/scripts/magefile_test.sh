#!/bin/bash

set -e  # Stop on any error

echo "Starting the test of magefile commands..."

# Compile for all supported OS and architectures
echo "Testing 'Compile' for all supported OS and architectures..."
release=true mage Compile

# Compile for specific OS and architectures
echo "Testing 'Compile' for macOS on arm64..."
GOOS=darwin GOARCH=arm64 release=false mage compile

echo "Testing 'Compile' for Linux on amd64..."
GOOS=linux GOARCH=amd64 release=false mage compile

echo "Testing 'Compile' for Windows on amd64..."
GOOS=windows GOARCH=amd64 release=false mage compile

# Install Dependencies
echo "Testing 'InstallDeps'..."
mage InstallDeps

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
