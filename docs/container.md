# Dev Container

## Using in VSCode

**Please note:** this functionality requires vanilla VSCode!

1. Install the [Remote - Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
   in VSCode if you haven't already.

2. Open the command palette in VSCode by pressing Ctrl + Shift + P
   (or Cmd + Shift + P on macOS) and run the
   `Remote-Containers: Reopen in Container` command.
   VSCode will build the Docker container using the provided
   Dockerfile and open the project inside the container.

   You can now work on your project inside the container, and VSCode will
   use the Go, Python, and other tools installed in the container for
   tasks like running, debugging, and linting.

If you want to go back to working on your project locally, you can open
the command palette and run the `Remote-Containers: Reopen Locally` command.

---

## Using on the CLI

1. Build the image

   ```bash
   export ARCH="$(uname -a | awk '{ print $NF }')"

   # Change DOCKER_DEFAULT_PLATFORM if we're on an ARM-based system.
   if [[ $ARCH == "arm64" ]]; then
       export DOCKER_DEFAULT_PLATFORM=linux/amd64
   fi

   docker build --build-arg USER_ID=$(id -u) \
       --build-arg GROUP_ID=$(id -g) -t facebookincubator/ttpforge .
   ```

2. Run container and mount local project directory

   ```bash
   docker run -it --rm \
       -v "$(pwd)":/home/ttpforge/go/src/github.com/facebookincubator/ttpforge \
       facebookincubator/ttpforge
   ```
