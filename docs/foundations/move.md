# Moving TTPs

## Overview

The `ttpforge move` command allows you to move or rename TTP files within
TTPForge repositories while automatically updating all references to a moved TTP.
This ensures any files referencing a TTP work correctly after a move operation.

**Note:** The `move` command does not support updating references that use
variable substitution (e.g., `ttp: //{{.Arg.here}}/ttp.yaml`). Only static
references are automatically updated. Be sure to manually update any dynamic
references after moving TTPs.

## Basic Usage

To move a TTP, use the command shown below:

```bash
ttpforge move [source] [destination]
```

## Cross-Repository Moves

You can move TTPs between repositories by specifying the source and destination.
For example, to move a TTP from the `examples` repository to the `forgearmory`
repository, you can use the following command:

```bash
ttpforge move examples//basic.yaml forgearmory//imported/basic.yaml
```

Any TTP files that reference a moved TTP will be updated to point to its new
location using relative references they are within the same repository, or
absolute references if they are in different repositories.

**Note:** These repository names are defined in your TTPForge configuration
file, so take caution when performing cross-repository moves as your config file
may not match the config files of other users which can create inconsistencies
in repository naming. Performing moves without a config file may also create
naming inconsistencies across repositories.

### Unsafe Moves

Use the `--unsafe` flag to skip dependency updates and force a move operation:

```bash
ttpforge move ttp examples//basic.yaml examples//basic-new.yaml --unsafe
```

**Warning:** Using `--unsafe` can break other TTPs that depend on the moved TTP.
Only use this flag when you're certain no other TTPs depend on the target file.

## Argument Formats

The `move` command accepts source and destination arguments in several formats.

**Note:** The source and destination paths should fall under a configured
repository path, but the destination will be created if it does not exist.

### 1. Repository Reference Format (Recommended)

You can use repository references to specify the source and destination:

```bash
ttpforge move examples//actions/inline/basic.yaml examples//actions/inline/basic-new.yaml
```

Format: `repo_name//path/to/ttp.yaml`

- `repo_name` - The name of the repository as defined in your TTPForge configuration
- `//` - The repository separator
- `path/to/ttp.yaml` - The path to the TTP within the repository's search paths

If moving within the same repository, you can omit the destination `repo_name`.

```bash
ttpforge move examples//actions/inline/basic.yaml //actions/inline/basic-new.yaml
```

### 2. Absolute Path Format

You can specify the full absolute path to TTP files on your filesystem:

```bash
ttpforge move /home/user/repo/ttps/basic.yaml /home/user/repo/ttps/basic-new.yaml
```

### 3. Mixed Format Support

You can mix and match formats between source and destination:

```bash
# Repository reference to absolute path
ttpforge move examples//basic.yaml /home/user/repo/ttps/moved.yaml

# Absolute path to repository reference
ttpforge move /home/user/repo/ttps/basic.yaml examples//moved.yaml
```

## Default Output Location

**Important**: When moving to a repository using the repository
reference format (`repo//path`), the TTP will be placed in the
**first path** defined in that repository's `ttpforge-repo-config.yaml` file.

For example, if your repository configuration contains:

```yaml
ttp_search_paths:
  - ttps
  - templates
  - examples
```

A move to `myrepo//basic.yaml` places the file at `<repo_root>/ttps/basic.yaml`.

## Examples

### Basic Move Within Repository

```bash
ttpforge move examples//basic.yaml examples//basic-renamed.yaml
```

### Move to Nested Directory

```bash
ttpforge move examples//basic.yaml examples//actions/basic/renamed.yaml
```

### Cross-Repository Move

```bash
ttpforge move examples//basic.yaml forgearmory//imported/basic.yaml
```

### Mixed Format Move

```bash
ttpforge move examples//basic.yaml /home/user/myrepo/ttps/basic.yaml
```

### Using Absolute Paths

```bash
ttpforge move /home/user/repo/ttps/old.yaml /home/user/repo/ttps/new.yaml
```
