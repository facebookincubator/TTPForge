# Finding and Running TTPs

## TTP Repositories

TTPs built with TTPForge are organized into collections called repositories.
Each repository contains multiple TTPs that are grouped into categories. You can
configure TTPForge with the default set of TTP repositories by running:

```bash
ttpforge init
```

You can then view the enabled TTP repositories by running:

```bash
ttpforge list repos
```

## Listing and Examining TTPs in Repositories

You can list the TTPs contained within all your TTP repositories as follows:

```bash
ttpforge list ttps
```

This will print a list of **TTP References**, which have the format
`[repository name]/path/to/ttp`

You can look at the configuration of any given TTP by running
`ttpforge show ttp [ttp-reference]` - for example:

```bash
ttpforge show ttp examples//cleanup/basic.yaml
```

To learn more about the TTPForge YAML configuration format, check out the
relevant [docs](actions.md).

## Running TTPs from Repositories

You can execute any TTP from a TTP repository by passing the appropriate TTP
reference to the `ttpforge run` command - for example:

```bash
ttpforge run examples//cleanup/basic.yaml
```

Note that you can also run TTPs by directly passing an (absolute or relative)
path to the appropriate YAML file - this can be convenient when you are
developing new TTPs and working with a local TTP repository checkout at a
non-standard path.

## Removing and Installing TTP Repositories

You can remove a TTP repository using the `ttpforge remove repo` command - we
will demonstrate this on the `forgearmory` repo:

```bash
ttpforge remove repo forgearmory
```

Notice how the `forgearmory` entry has now disappeared from the output of
`ttpforge list repos`. We can then reinstall that repository using the
`ttpforge install repo` command:

```bash
ttpforge install repo --name forgearmory https://github.com/facebookincubator/forgearmory
```

The `ttpforge install` runs `git clone` under the hood, so it will work with any
valid URL that you could pass to `git clone`.

## Creating Your Own TTP Repository

In order to create your own TTP repository, whether for use at your own company
or for open-source publication, you just need to:

1. Create a Git Repository
2. Add the appropriate
   [TTPForge Repository Configuration](#repository-configuration-files) file.

You can then install your repository with `ttpforge install repo` as
demonstrated above.

## How TTPForge Manages TTP Repos

### The Global TTPForge Configuration File

TTPForge keeps track of the TTP repositories installed on your system by using
the **TTPForge Global Configuration File**, which is stored at
`~/.ttpforge/config.yaml` by default. Take a look inside that file (if it isn't
there, you should run `ttpforge init`) - you should see contents similar to the
following:

```yml
---
repos:
  - name: examples
    path: repos/examples
    git:
      url: https://github.com/facebookincubator/TTPForge
  - name: forgearmory
    path: repos/forgearmory
    git:
      url: https://github.com/facebookincubator/forgearmory
```

Each of the `path:` entries above is a _relative path_ that is interpreted based
on the directory of the TTPForge configuration file. Therefore, in the above
example, these repository paths map to `~/.ttpforge/repos/examples` and
`~/.ttpforge/repos/forgearmory`. Feel free to `ls` those files and look around
the TTP repositories.

**Note**: One can also use absolute repository paths in this configuration file.
This may be useful if your company maintains your own internal TTPForge
repositories and assigns those internal repositories to standardized
installation paths.

### Repository Configuration Files

Each TTP repository contains a `ttpforge-repo-config.yaml` file in the
repository root directory. This file specifies which folders within the
repository contain TTPs that TTPForge should index. For example, you can examine
the repository configuration file for the `examples` repo:

```bash
cat ~/.ttpforge/repos/examples/ttpforge-repo-config.yaml
```

You should see something like this, which tells TTPForge that the TTPs from this
repository live in `example-ttps`:

```yml
---
ttp_search_paths:
  - example-ttps
```

Note that repository owners may add as many `ttp_search_path` entries as they
wish.

### Using a Custom Configuration File

You can override the global configuration file by passing the
`-c [config-file-path]` option to any TTPForge command. You probably won't ever
need to do this, although we do use this feature extensively in the unit tests
for the TTPForge codebase itself.
