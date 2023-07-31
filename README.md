# TTPForge

[![Tests](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml)
[![Renovate](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml)
[![Nancy 3p Vulnerability Scan](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml)

This repo hosts the TTPForge tool created by Meta's Purple Team.
It is intended to provide an interface to execute TTPs across various
targets and mediums.

---

## Table of Contents

- [Getting Started - User](#getting-started-as-a-user)
- [Getting Started - Developer](docs/dev.md)
- [Using the TTPForge Dev Container](docs/container.md)
- [Code Standards](docs/code-standards.md)
- [Creating a new release](docs/release.md)
- [TTPForge Building Blocks](docs/building-blocks.md)

---

## Getting started as a user

1. Download and install the [gh cli tool](https://cli.github.com/):

   - [macOS](https://github.com/cli/cli#macos)
   - [Linux](https://github.com/cli/cli/blob/trunk/docs/install_linux.md)
   - [Windows](https://github.com/cli/cli#windows)

1. Get latest TTPForge release:

   ```bash
   # Download utility functions
   bashutils_url="https://raw.githubusercontent.com/l50/dotfiles/main/bashutils"

   # Define the local path of bashutils.sh
   bashutils_path="/tmp/bashutils"

   if [[ ! -f "${bashutils_path}" ]]; then
      # bashutils.sh doesn't exist locally, so download it
      curl -s "${bashutils_url}" -o "${bashutils_path}"
   fi

   # Source bashutils
   # shellcheck source=/dev/null
   source "${bashutils_path}"

   fetchFromGithub "facebookincubator" "TTPForge" "v0.0.5" ttpforge $GITHUB_TOKEN
   ```

   At this point, the latest `ttpforge` release should be in
   `~/.local/bin/ttpforge` and subsequently, the `$USER`'s `$PATH`.

1. Run a basic example:

   ```bash
   ./ttpforge -c config.yaml \
     run examples/variadic-params/variadicParameterExample.yaml \
     --arg name=jimbo \
     --arg password=fakepassword123
   ```
