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

- [Getting Started](#getting-started)
- [Using the TTPForge Dev Container](docs/container.md)
- [Developer Environment Setup](docs/dev.md)
- [Writing Tests](docs/testing.md)
- [Creating a new release](docs/release.md)
- [TTPForge Building Blocks](docs/building-blocks.md)

---

## Getting Started

1. Download and install the [gh cli tool](https://cli.github.com/).

1. Get latest TTPForge release:

   ```bash
   OS="$(uname | python3 -c 'print(open(0).read().lower().strip())')"
   ARCH="$(uname -a | awk '{ print $NF }')"
   gh release download -p "*${OS}_${ARCH}.tar.gz"
   tar -xvf *tar.gz
   ```

1. Run the Hello World example:

   ```bash
   ./ttpforge -c config.yaml run ttps/privilege-escalation/credential-theft/hello-world/ttp.yaml
   ```
