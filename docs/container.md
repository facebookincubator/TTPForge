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
- [Build base zsh image](#build-base-zsh-image)

---

## Prerequisites

- [Create a classic GitHub Personal Access Token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token)
  (fine-grained isn't supported yet) with the following permissions
  taken from [here](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry):

      - `read:packages`
      - `write:packages`
      - `delete:packages`

## Using in VSCode

**Please note: this functionality requires vanilla VSCode!**

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

1. Login to ghcr:

   ```bash
   docker login ghcr.io -u $GITHUB_USERNAME -p $PAT
   ```

1. Pull the latest image from ghcr:

   - ZSH container:

      ```zsh
      if [[ "$(uname -a | awk '{ print $NF }')" == "arm64" ]]; then
         docker pull --platform linux/x86_64 ghcr.io/facebookincubator/ttpforge-zsh
      else
         docker pull ghcr.io/facebookincubator/ttpforge-zsh:latest
      fi
      ```

   - Bash Container:

      ```bash
      docker pull ghcr.io/facebookincubator/ttpforge-bash:latest
      ```

1. Run container and mount local project directory

   - ZSH container:

      ```zsh
      docker run -it --rm \
         -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
         ghcr.io/facebookincubator/ttpforge-zsh:latest
      ```

   - ZSH container with custom dotfiles:

      ```zsh
      docker run -it --rm \
         -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
         -v "${HOME}/.zshrc:/home/ttpforge/.zshrc" \
         -v "${HOME}/.dotfiles:/home/ttpforge/.dotfiles" \
         ghcr.io/facebookincubator/ttpforge-zsh:latest
      ```

   - Bash Container:

      ```bash
      docker run -it --rm \
         -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
         ghcr.io/facebookincubator/ttpforge-bash:latest
      ```

---

## Local Build Process

If, for any reason, you need to build the container image
locally, follow the steps below.

Run the following commands to build the image locally:

```bash
raw_arch=$(uname -m)
case $raw_arch in
   x86_64)
         ARCH="amd64"
         ;;
   arm64)
         ARCH="arm64"
         ;;
         *)
         echo "Unsupported architecture: $raw_arch"
         exit 1
         ;;
esac
export ARCH

# Change DOCKER_DEFAULT_PLATFORM if we're on an ARM-based system.
if [[ $ARCH == "arm64" ]]; then
      export DOCKER_DEFAULT_PLATFORM=linux/amd64
fi

# Build the ZSH Dockerfile with vncserver - can be used with guacamole
docker build --build-arg USER_ID=$(id -u) \
         --build-arg GROUP_ID=$(id -g) \
         -t facebookincubator/ttpforge-zsh \
         -f .devcontainer/zsh/Dockerfile .

# Build the Bash Dockerfile
docker build --build-arg USER_ID=$(id -u) \
         --build-arg GROUP_ID=$(id -g) \
         -t facebookincubator/ttpforge-bash \
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

---

## Build base zsh image

These instructions are only relevant for repo maintainers that need
to build a new base image or update the existing base zsh image.

1. Download and install the [gh cli tool](https://cli.github.com/).

1. Clone the [ansible-vnc-zsh ansible playbook](https://github.com/CowDogMoo/ansible-vnc-zsh):

   ```bash
   gh repo clone CowDogMoo/ansible-zsh-vnc ~/ansible-zsh-vnc
   ```

1. Clone and compile the
   [ttpforge](https://github.com/facebookincubator/ttpforge) project:

   ```bash
   gh repo clone facebookincubator/TTPForge ~/ttpforge
   cd ~/ttpforge
   go build -o wg
   ```

1. Update the existing `ansible-vnc-zsh` blueprint config:

   ```bash
   cat <<EOM > blueprints/ansible-vnc-zsh/config.yaml
   ---
   blueprint:
     name: ansible-vnc-zsh

   packer_templates:
     - name: ubuntu-vnc-zsh.pkr.hcl
       base:
         name: ubuntu
         version: latest
       systemd: false
       tag:
         name: facebookincubator/pt-ubuntu-vnc-zsh
         version: latest

     - name: ubuntu-systemd-vnc-zsh.pkr.hcl
       base:
         name: geerlingguy/docker-ubuntu2204-ansible
         version: latest
       systemd: true
       tag:
         name: facebookincubator/pt-ubuntu-vnc-zsh-systemd
         version: latest

   container:
     workdir: /home/ubuntu
     entrypoint: "/run/docker-entrypoint.sh ; zsh"
     user: ubuntu
     registry:
       server: ghcr.io
       username: facebookincubator
   EOM
   ```

1. Download [warpgate](https://github.com/CowDogMoo/warpgate/)

1. Build the base image:

   ```bash
   ./wg imageBuilder -b ansible-vnc-zsh -p ~/ansible-vnc-zsh
   ```
