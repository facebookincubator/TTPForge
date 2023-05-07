FROM golang:1.20.4

# Set build-time arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000

# Install necessary dependencies
RUN DEBIAN_FRONTEND="noninteractive" apt-get update \
    && apt-get install -y curl git python3 python3-pip ruby-full shellcheck unzip vim

# Create the ttpforge user and change ownership of the necessary directories
RUN useradd -m -s /bin/bash ttpforge \
    && mkdir -p /home/ttpforge/go/src/github.com/facebookincubator/TTPForge \
    && chown -R ttpforge:ttpforge /home/ttpforge/go

# Install go dependencies
RUN go install github.com/magefile/mage@latest \
	&& go install mvdan.cc/sh/v3/cmd/shfmt@latest

# Install tools function
COPY .devcontainer/install-gh-release.sh /usr/local/bin/install-gh-release
RUN chmod +x /usr/local/bin/install-gh-release

# Install gh cli and terragrunt with their respective releases on GitHub.
ARG TARGETPLATFORM
RUN export OS="$(uname | python3 -c 'print(open(0).read().lower().strip())')" \
    && if [ "$TARGETPLATFORM" = "linux/amd64" ]; then ARCH=amd64; elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then ARCH=arm64; else exit 1; fi \
    && GITHUB_CLI_VERSION=2.27.0 \
    && install-gh-release \
    "GitHub CLI" \
    "gh" \
    "https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}" \
    "gh_${GITHUB_CLI_VERSION}_${OS}_${ARCH}.deb" \
    "gh_${GITHUB_CLI_VERSION}_checksums.txt" \
    "--ignore-missing" \
    "512" \
    "gh --version" \
    && mkdir -p /home/ttpforge/.local/bin/ \
    && mv "$(command -v gh)" /home/ttpforge/.local/bin/gh \
    && chown ttpforge:ttpforge -R /home/ttpforge/.local \
    && TERRAGRUNT_VERSION=0.45.2 \
    && install-gh-release \
    "Terragrunt" \
    "terragrunt" \
    "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}" \
    "terragrunt_${OS}_${ARCH}" \
    "SHA256SUMS" \
    "--ignore-missing" \
    "256" \
    "terragrunt -v"

COPY --chown=ttpforge . /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install pre-commit
RUN python3 -m pip install --upgrade pip \
    && python3 -m pip install --upgrade pre-commit \
    && pre-commit install

# Install project dependencies
RUN mage installDeps \
    && mage \
    && mage runPreCommit

# Run go mod tidy and build the project
# to speed up builds
RUN go mod tidy && go build -o ttpforge

# Remove the copied repository content
RUN rm -rf /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

CMD ["bash"]
