# TTP Forge

[![Tests](https://github.com/facebookincubator/TTP-Runner/actions/workflows/tests.yaml/badge.svg)](https://github.com/facebookincubator/TTP-Runner/actions/workflows/tests.yaml)
[![🚨 Semgrep Analysis](https://github.com/facebookincubator/TTP-Runner/actions/workflows/semgrep.yaml/badge.svg)](https://github.com/facebookincubator/TTP-Runner/actions/workflows/semgrep.yaml)
[![Renovate](https://github.com/facebookincubator/TTP-Runner/actions/workflows/renovate.yaml/badge.svg)](https://github.com/facebookincubator/TTP-Runner/actions/workflows/renovate.yaml)

This repo hosts the TTP Forge tool created by Meta's Purple Team.
It is intended to provide an interface to execute TTPs across various
targets and mediums.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Developer Environment Setup](docs/dev.md)

---

## Getting Started

1. Download and install the [gh cli tool](https://cli.github.com/).

1. Get latest TTP Forge release:

   ```bash
   OS="$(uname | python3 -c 'print(open(0).read().lower().strip())')"
   ARCH="$(uname -a | awk '{ print $NF }')"
   gh release download -p "*${OS}_${ARCH}.tar.gz"
   tar -xvf *tar.gz
   ```
