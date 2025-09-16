# Enumeration

## Enumerating TTPs Based on Provided Filters

To enumerate the TTPs in a repo, use the command shown below:

```bash
ttpforge enum ttps --platforms <platform1>,<platform2> --repo <repo> --tactic <tactic>
--technique <technique> --sub-tech <subtechnique> --verbose
```

Please note:

1) Allowlist of platforms = [linux, windows, darwin]
2) Repo details are present in ~/.ttpforge/config.yaml or you specify it
using `--config <path>`
TTPForge will be able to find and enumerate all TTPs in the config file.

The output is a platform-wise count of TTPs in the repo along with other
information like total count and total match count after applying filters.

## Enumerating TTP Dependencies

To enumerate the dependencies of a TTP, use the command shown below:

```bash
ttpforge enum dependencies [repo_name//path/to/ttp] --verbose
```

The output lists all dependencies that rely on the TTP as well as a total count.
