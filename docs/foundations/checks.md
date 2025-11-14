# TTPForge Checks

## Overview

Checks allow TTP authors to verify that their steps executed correctly and
weren't silently blocked by security tools like EDR/AV software.

## Why Use Checks?

When executing TTPs (Tactics, Techniques, and Procedures), it's critical to
know whether the operations actually succeeded. Security tools may silently
block malicious actions without returning error codes. Checks provide
confidence that:

1. **Files were actually created** - Not just that the command ran,
   but that the file exists on disk
2. **Commands produced expected output** - Verify services are running,
   processes exist, etc.
3. **Exit codes match expectations** - Detect when operations fail
   silently
4. **Content integrity** - Ensure files weren't modified or corrupted
   by security tools

## Available Check Types

### 1. Path Exists Check

Verifies that a file exists at a specified path and optionally validates
its contents using SHA256 checksums, content patterns, and
permissions.

**Fields:**

- `path_exists` (required): Path to the file to verify
- `checksum.sha256` (optional): SHA256 hash to verify file contents
- `content_contains` (optional): String that must appear in file content
- `content_not_contains` (optional): String that must NOT appear in file
- `content_regex` (optional): Regex pattern to match against file content
- `permissions` (optional): File permissions in octal format (e.g., "0755")

**Example:**

```yaml
checks:
  # Basic file existence check
  - msg: "Payload file was not created"
    path_exists: /tmp/payload.txt

  # File existence with SHA256 checksum verification
  - msg: "File exists but contents were modified"
    path_exists: /tmp/payload.txt
    checksum:
      sha256: 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae

  # Check file contains expected content
  - msg: "Config file should contain setting"
    path_exists: /etc/app.conf
    content_contains: "enabled=true"

  # Check file does NOT contain unwanted content
  - msg: "File should not contain debug flags"
    path_exists: /etc/app.conf
    content_not_contains: "DEBUG=1"

  # Check file content matches regex
  - msg: "Version file should have proper format"
    path_exists: /tmp/version.txt
    content_regex: "v[0-9]+\\.[0-9]+\\.[0-9]+"

  # Check file permissions (Unix/Linux/macOS)
  - msg: "Script should be executable"
    path_exists: /tmp/payload.sh
    permissions: "0755"

  # Combine multiple checks
  - msg: "Malware file should exist with correct perms and content"
    path_exists: /tmp/malware.bin
    permissions: "0644"
    content_contains: "malicious_code"
    checksum:
      sha256: abc123...
```

### 2. Command Check

Executes a command and verifies its exit code and/or output. This check is
**cross-platform compatible** - commands execute via `sh -c` on
Linux/macOS and `cmd.exe /c` on Windows.

**Fields:**

- `command` (required): Command to execute
- `expect_exit_code` (optional): Expected exit code (defaults to 0)
- `output_contains` (optional): String that must appear in output
- `output_not_contains` (optional): String that must NOT appear in output
- `output_regex` (optional): Regex pattern to match against output

**Example:**

```yaml
checks:
  # Basic command success check
  - msg: "Command should succeed"
    command: "echo hello"

  # Check specific exit code
  - msg: "File should not exist"
    command: "test -f /tmp/should-not-exist"
    expect_exit_code: 1

  # Check output contains string
  - msg: "Service should be running"
    command: "systemctl is-active myservice"
    output_contains: "active"

  # Check output does NOT contain string
  - msg: "No errors in log"
    command: "cat /var/log/app.log"
    output_not_contains: "ERROR"

  # Check output matches regex
  - msg: "Version should match semantic versioning"
    command: "myapp --version"
    output_regex: "[0-9]+\\.[0-9]+\\.[0-9]+"

  # Combine multiple conditions
  - msg: "Deployment successful with timestamp"
    command: "cat /tmp/deploy-status.txt"
    expect_exit_code: 0
    output_contains: "SUCCESS"
    output_regex: "[0-9]{4}-[0-9]{2}-[0-9]{2}"
```

## Using Checks in TTP YAML

Checks are added to the `checks` field of a step. Multiple checks can be
defined, and they execute sequentially after the step completes.

### Basic Usage

```yaml
steps:
  - name: create_malicious_file
    create_file: /tmp/malware.bin
    contents: "malicious payload"
    cleanup: default
    checks:
      - msg: "Malware file should exist"
        path_exists: /tmp/malware.bin
```

### Multiple Checks

**All checks must pass for the step to succeed.** You can mix different
check types:

```yaml
steps:
  - name: install_persistence
    inline: |
      echo "payload" > /tmp/payload.sh
      chmod +x /tmp/payload.sh
    cleanup:
      inline: rm -f /tmp/payload.sh
    checks:
      # Verify file exists
      - msg: "Payload file should be created"
        path_exists: /tmp/payload.sh

      # Verify file is executable
      - msg: "Payload should be executable"
        command: "test -x /tmp/payload.sh"

      # Verify file contains expected content
      - msg: "Payload should contain correct code"
        command: "cat /tmp/payload.sh"
        output_contains: "payload"
```

### Using Checks as Prerequisites

Checks can also verify prerequisites before executing potentially dangerous
operations:

```yaml
steps:
  - name: verify_target_exists
    description: "Verify target file exists before modification"
    print_str: "Checking for target file..."
    checks:
      - msg: "Target file must exist"
        path_exists: ~/.bashrc

  - name: modify_target
    description: "Now safe to modify since we verified it exists"
    inline: |
      echo 'export MALICIOUS_VAR=true' >> ~/.bashrc
```

### Cross-Platform Checks

When writing cross-platform TTPs, remember that command checks
automatically use the appropriate shell:

```yaml
requirements:
  platforms:
    - os: linux
    - os: darwin
    - os: windows

steps:
  - name: cross_platform_check
    inline: echo "test"
    checks:
      # This works on all platforms
      - msg: "Echo should work"
        command: "echo hello"
        output_contains: "hello"
```
