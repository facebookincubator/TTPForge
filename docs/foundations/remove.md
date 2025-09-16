# Removing TTPs and Repositories

## Overview

The `ttpforge remove` command allows you to remove TTPs and repositories from
TTPForge. This command provides two subcommands:

1. `repo` for removing entire repositories and their associated configuration
2. `ttp` for removing individual TTP files with automatic dependency checking

## Remove a Repository

To remove an entire repository:

```bash
ttpforge remove repo [repo_name]
```

When removing a repository, TTPForge will locate the repository by name in your
configuration, delete all files in its file system path, remove the repository
entry from your TTPForge configuration file, and save the updated configuration.

**Warning:** Repository removal is irreversible. Ensure you have backups of any
important TTPs before removing a repository.

## Remove a TTP

To remove a specific TTP file:

```bash
ttpforge remove ttp [repo_name//path/to/ttp]
```

If dependencies are found, the command will list all files that reference the
TTP and exit without deleting it. This prevents accidental removal of TTPs that
are still in use.

### Unsafe Removal

Use the `--unsafe` flag to skip dependency checking and force deletion:

```bash
ttpforge remove ttp --unsafe examples//basic.yaml
```

**Warning:** Using `--unsafe` can break other TTPs that depend on the
removed TTP. Only use this flag when you're certain no other TTPs
depend on the target file.
