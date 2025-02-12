# TTPForge

[![License](https://img.shields.io/github/license/facebookincubator/TTPForge?label=License&style=flat&color=blue&logo=github)](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
[![Tests](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/facebookincubator/TTPForge/actions/workflows/semgrep.yaml)
[![Coverage Status](https://coveralls.io/repos/github/facebookincubator/TTPForge/badge.svg)](https://coveralls.io/github/facebookincubator/TTPForge)

TTPForge is a cyber attack simulation platform designed and built by Sam Manzer (@d3sch41n),
Alek Straumann (@CrimsonK1ng), and Geoff Pamerleau (@Sy14r),
and including subsequent contributions from many good folks
in Meta’s Red, Blue, and Purple security teams.
Jayson Grace (@l50) migrated the project to GitHub and
assisted with preparation for the project’s open source release.

This project promotes a Purple
Team approach to cybersecurity with the following goals:

- To help blue teams accurately measure their detection and response
  capabilities through high-fidelity simulations of real attacker activity.
- To help red teams improve the ROI/actionability of their findings by packaging
  their attacks as automated, repeatable simulations.

TTPForge allows you to automate attacker tactics, techniques, and procedures
(TTPs) using a powerful but easy-to-use YAML format. Check out the links below
to learn more!

---

## Table of Contents

- [Installation](#installation)
- [Documentation](docs/foundations/README.md)
- [Getting Started - Developer](docs/dev/README.md)
- [Go Package Documentation](https://pkg.go.dev/github.com/facebookincubator/ttpforge@main)

---

## Installation

1. Get latest TTPForge release:

   ```bash
   curl \
   https://raw.githubusercontent.com/facebookincubator/TTPForge/main/dl-rl.sh \
   | bash
   ```

   At this point, the latest `ttpforge` release should be in
   `$HOME/.local/bin/ttpforge` and subsequently, the `$USER`'s `$PATH`.

   If running in a stripped down system, you can add TTPForge to your `$PATH`
   with the following command:

   ```bash
   export PATH=$HOME/.local/bin:$PATH
   ```

1. Initialize TTPForge configuration

   This command will place a configuration file at the default location
   `~/.ttpforge/config.yaml` and configure the `examples` and `forgearmory` TTP
   repositories:

   ```bash
   ttpforge init
   ```

1. List available TTP repositories (should show `examples` and `forgearmory`)

   ```bash
   ttpforge list repos
   ```

   The `examples` repository contains the TTPForge examples found in this
   repository. The
   [ForgeArmory](https://github.com/facebookincubator/ForgeArmory) repository
   contains our arsenal of attacker TTPs powered by TTPForge.

1. List available TTPs that you can run:

   ```bash
   ttpforge list ttps
   ```

1. Examine an example TTP:

   ```bash
   ttpforge show ttp examples//args/basic.yaml
   ```

1. Run the specified example:

   ```bash
   ttpforge run examples//args/basic.yaml \
     --arg str_to_print=hello \
     --arg run_second_step=true
   ```
