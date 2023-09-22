# TTPForge

[![License](https://img.shields.io/github/license/facebookincubator/TTPForge?label=License&style=flat&color=blue&logo=github)](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
[![Tests](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml)
[![🚨 CodeQL Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/codeql-analysis.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/codeql-analysis.yaml)
[![🚨 Nancy 3p Vulnerability Scan](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/nancy.yaml)
[![Renovate](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/renovate.yaml)
[![Coverage Status](https://coveralls.io/repos/github/facebookincubator/TTPForge/badge.svg)](https://coveralls.io/github/facebookincubator/TTPForge)

TTPForge is a cyber attack simulation platform. This project promotes a
Purple Team approach to cybersecurity with the following goals:

* To help blue teams accurately measure their detection and response capabilities
  through high-fidelity simulations of real attacker activity.
* To help red teams improve the ROI/actionability of their findings by packaging
  their attacks as automated, repeatable simulations.

TTPForge allows you to automate  attacker tactics, techniques, and procedures (TTPs)
using a powerful but easy-to-use YAML format. Check out an [example](#examples) or
head straight to the [installation](#installation) instructions.

TTPForge is written in [Go](https://go.dev/) and was originally created
by @d3sch41n, @CrimsonK1ng, and @l50, members of Meta's Offensive Security Group.

## Installation

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

TTPForge is now ready to use - check out our [tutorial](#tutorial) to
start exploring its capabilities.

## Examples

Here's an example of TTPForge's YAML format; this
[TTP](https://github.com/facebookincubator/ForgeArmory/blob/main/ttps/defense-evasion/macos/disable-system-updates/disable-system-updates.yaml)
disables macOS security updates:

```yaml
---
name: Disable system security updates
description: |
  This TTP disables the automatic installation of macOS security updates.
mitre:
  tactics:
    - TA0005 Defense Evasion
  techniques:
    - T1562 Impair Defenses
  subtechniques:
    - "T1562.001 Impair Defenses: Disable or Modify Tools"

steps:
  - name: disable-updates
    inline: |
      echo -e "===> Disabling automatic installation of security updates..."
      sudo defaults write /Library/Preferences/com.apple.SoftwareUpdate.plist CriticalUpdateInstall -bool NO
      echo "[+] DONE!"

    cleanup:
      inline: |
        echo -e "===> Enabling automatic installation of security updates..."
        sudo defaults write /Library/Preferences/com.apple.SoftwareUpdate.plist CriticalUpdateInstall -bool YES
        echo "[+] DONE!"
```

Head over to the [ForgeArmory](https://github.com/facebookincubator/ForgeArmory),
our collection of open-source TTPs powered by TTPForge,
to check out more examples!

---

## Tutorial

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

## Documentation

Check out our documentation on the following topics:

- [Developer Documentation](docs/dev/README.md)
- [Using the TTPForge Dev Container](docs/container.md)
- [Code Standards](docs/code-standards.md)
- [Creating a new release](docs/release.md)
- [TTPForge Building Blocks](docs/building-blocks.md)
