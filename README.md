# TTPForge

[![License](https://img.shields.io/github/license/facebookincubator/TTPForge?label=License&style=flat&color=blue&logo=github)](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
[![Tests](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml)
[![🚨 CodeQL Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/codeql-analysis.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/codeql-analysis.yaml)
[![🚨 Nancy 3p Vulnerability Scan](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml)
[![Renovate](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml)
[![Coverage Status](https://coveralls.io/repos/github/facebookincubator/TTPForge/badge.svg)](https://coveralls.io/github/facebookincubator/TTPForge)

This repo hosts the TTPForge tool created by Meta's Purple Team.
It is intended to provide an interface to execute TTPs across various
targets and mediums.

---

## Table of Contents

- [Getting Started - User](#getting-started-as-a-user)
- [Getting Started - Developer](docs/dev/README.md)
- [Using the TTPForge Dev Container](docs/container.md)
- [Code Standards](docs/code-standards.md)
- [Creating a new release](docs/release.md)
- [TTPForge Building Blocks](docs/building-blocks.md)

---

## Getting started as a user

1. Get latest TTPForge release:

   ```bash
   bashutils_url="https://raw.githubusercontent.com/l50/dotfiles/main/bashutils"

   bashutils_path="/tmp/bashutils"

   if [[ ! -f "${bashutils_path}" ]]; then
      curl -s "${bashutils_url}" -o "${bashutils_path}"
   fi

   source "${bashutils_path}"

   fetchFromGithub "facebookincubator" "TTPForge" "v1.0.3" ttpforge

   # Optionally, if you are using the `gh` cli:
   fetchFromGithub "facebookincubator" "TTPForge" "v1.0.3" ttpforge $GITHUB_TOKEN
   ```

   At this point, the latest `ttpforge` release should be in
   `~/.local/bin/ttpforge` and subsequently, the `$USER`'s `$PATH`.

   If running in a stripped down system, you can add TTPForge to your `$PATH`
   with the following command:

   ```bash
   export PATH=$HOME/.local/bin:$PATH
   ```

1. Initialize TTPForge configuration

   This command will place a configuration file at the default location
   `~/.ttpforge/config.yaml` and download the
   [ForgeArmory](https://github.com/facebookincubator/ForgeArmory)
   TTPs repository:

   ```bash
   ttpforge init
   ```

1. List available TTP repositories (should show `forgearmory`)

   ```bash
   ttpforge list repos
   ```

1. List available TTPs that you can run:

   ```bash
   ttpforge list ttps
   ```

1. Examine an example TTP:

   ```bash
   ttpforge show ttp forgearmory//examples/args/define-args.yaml
   ```

1. Run the specified example:

   ```bash
   ttpforge run \
     forgearmory//examples/args/define-args.yaml \
     --arg a_message="hello" \
     --arg a_number=1337
   ```
