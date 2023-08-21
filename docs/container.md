# TTPForge Dev Container

We provide a development container hosted on
GitHub Container Registry (ghcr) to serve as a
complete development environment.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Using in VSCode](#using-in-vscode)
- [Using on the Command Line Interface (CLI)](#using-on-the-command-line-interface-cli)
- [Local Build Process](#local-build-process)
- [Run container action locally](#run-container-action-locally)

---

## Prerequisites

- [Create a classic GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token)
  (fine-grained isn't supported yet) with the following permissions
  taken from [here](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry):

      - `read:packages`
      - `write:packages`
      - `delete:packages`

## Using in VSCode

1. If you haven't already, install the
   [Remote - Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
   in Visual Studio Code.

1. Open the command palette in Visual Studio Code by
   pressing Ctrl + Shift + P (or Cmd + Shift + P on macOS)
   and run the `Remote-Containers: Reopen in Container`
   command. Visual Studio Code will build the Docker
   container using the provided Dockerfile and open
   the project inside the container.

   You can now work on your project within the container,
   and Visual Studio Code will utilize the Go, Python,
   and other tools installed in the container for tasks
   such as running, debugging, and linting.

To return to working on your project locally, open the command palette
and run the `Remote-Containers: Reopen Locally` command.

---

## Using on the Command Line Interface (CLI)

1. Pull the latest image from ghcr:

   ```bash
   docker pull ghcr.io/facebookincubator/ttpforge:latest
   ```

1. Run container and mount local project directory:

   ```bash
   docker run -it --rm \
      -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
      ghcr.io/facebookincubator/ttpforge:latest
   ```

---

## Local Build Process

If, for any reason, you need to build the container image
locally, follow the steps below.

Run the following commands to build the image locally:

```bash
docker build \
    -t ghcr.io/facebookincubator/ttpforge \
    -f .devcontainer/bash/Dockerfile .
```

---

## Run container action locally

1. Create a file called `.secrets` with the following:

   ```bash
   export BOT_TOKEN=YOUR_PAT_GOES_HERE
   export GITHUB_USERNAME=YOUR_GITHUB_USERNAME_GOES_HERE
   ```

1. Install [Act](https://github.com/nektos/act)

1. Run one of the actions:

   ```bash
   # Run the container build, push, and test action

   act -j "test_pushed_images" \
      --secret-file .secrets

   # Run the pre-commit action
   act -j "pre-commit" \
      --secret-file .secrets \
      --container-architecture linux/amd64
   ```
