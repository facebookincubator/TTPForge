# Enumeration

## Enumerating TTPs Based on Provided Filters

To enumerate the TTPs in a repo, use the command shown below:

```bash
ttpforge enum ttps --platform <platform1>,<platform2> --repo <repo> \
  --tactic <tactic> --technique <technique> --sub-tech <subtechnique> \
  --author <author> --verbose
```

Please note:

1. For platforms, you can filter by a single item or multiple
comma-separated values. These work similarly to an OR search where you can
enumerate the files that include any of the provided values.
    - Allowlist of platforms = [linux, windows, darwin, any]
2. Repo details are present in `~/.ttpforge/config.yaml` or you specify it
using `--config <path>`. TTPForge will be able to find and enumerate all TTPs
in the config file.

The output is a platform-wise count of TTPs in the repo along with other
information like total count and total match count after applying filters.

## Enumerating TTP Dependencies

To enumerate the dependencies of a TTP, use the command shown below:

```bash
ttpforge enum dependencies [repo_name//path/to/ttp] --verbose
```

The output lists all dependencies that rely on the TTP as well as a total count.

## Enumerating Authors

To enumerate TTP contributions by author, use the command shown below:

```bash
ttpforge enum authors --repo <repo>
```

The output displays each author with their TTP contribution count and percentage,
sorted by count in descending order. Authors with more contributions appear first.
Unattributed TTPs (those without an author field) are listed at the end.

Use the `--verbose` flag to see the specific TTPs attributed to each author.
