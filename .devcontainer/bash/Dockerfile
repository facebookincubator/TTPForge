FROM golang:1.20.7-alpine

LABEL org.opencontainers.image.description="Dev Container with dependencies and tools for building and testing the TTPForge project"

ARG GITHUB_CLI_VERSION=2.32.1
ARG SETUP_TOOLS_VER=65.5.1 # Necessary to remediate vulnerabilities in older versions

# Install dependencies
RUN apk add --no-cache \
    bash \
    binutils-gold \
    build-base \
    curl \
    gcc \
    git \
    gpgme \
    libc-dev \
    make \
    nodejs \
    npm \
    openssl-dev \
    python3 \
    python3-dev \
    py3-pip \
    readline-dev \
    ruby-dev \
    ruby-full \
    shellcheck \
    sqlite-dev \
    unzip \
    vim \
    yaml-dev \
    zlib-dev

# Create the ttpforge user and change ownership of the necessary directories
RUN adduser -D -s /bin/sh ttpforge \
    && mkdir -p /home/ttpforge/go/src/github.com/facebookincubator/TTPForge \
    && chown -R ttpforge:ttpforge /home/ttpforge

# Install GitHub CLI
RUN export ARCH="$(uname -m)" \
    && case ${ARCH} in \
       x86_64) ARCH='amd64' ;; \
       aarch64) ARCH='arm64' ;; \
       arm*) ARCH='armv6' ;; \
       *) echo "Unsupported architecture: ${ARCH}" && exit 1 ;; \
       esac \
    && curl -sSL "https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_linux_${ARCH}.tar.gz" | tar xz \
    && mv "gh_${GITHUB_CLI_VERSION}_linux_${ARCH}/bin/gh" /usr/local/bin \
    && gh --version

USER ttpforge
ENV HOME=/home/ttpforge \
    GOPATH=/home/ttpforge/go \
    PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin


WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge
COPY --chown=ttpforge . .

# Setup Python environment
RUN python3 -m venv $HOME/venv \
    && . $HOME/venv/bin/activate \
    && pip3 install --upgrade pip \
    "setuptools>=${SETUP_TOOLS_VER}" \
    pre-commit

# Update PATH to include virtual environment
ENV PATH=$HOME/venv/bin:$PATH

# Install Go dependencies and build
RUN go install github.com/magefile/mage@latest \
    && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
    && export PATH=$PATH:$GOPATH/bin \
    && mage installDeps \
    && mage \
    && go mod tidy \
    && go build -o ttpforge \
    && rm ttpforge

# Update PATH to include Go binaries
ENV PATH=$GOPATH/bin:$PATH

CMD ["sh"]
