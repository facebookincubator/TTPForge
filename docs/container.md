# TTPForge Dev Container

We provide a development container hosted on
GitHub Container Registry (ghcr) to serve as a
complete development environment.

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
   docker login ghcr.io -u $GITHUB_USERNAME -p $PERSONAL_ACCESS_TOKEN
   ```

1. Pull the latest image from ghcr:

   ```bash
   docker pull ghcr.io/facebookincubator/ttpforge:latest
   ```

1. Run container and mount local project directory

   ```bash
   docker run -it --rm \
       -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
       facebookincubator/ttpforge:latest
   ```

---

## Manual Build Process

If, for any reason, you need to build the container image
locally, follow the steps below.

Run the following commands to build the image locally:

```bash
export ARCH="$(uname -a | awk '{ print $NF }')"

# Change DOCKER_DEFAULT_PLATFORM if we're on an ARM-based system.
if [[ $ARCH == "arm64" ]]; then
      export DOCKER_DEFAULT_PLATFORM=linux/amd64
fi

docker build --build-arg USER_ID=$(id -u) \
      --build-arg GROUP_ID=$(id -g) -t facebookincubator/ttpforge .
```
