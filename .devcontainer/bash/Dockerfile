FROM golang:1.21.0-alpine

LABEL org.opencontainers.image.description="Dev Container with dependencies and tools for building and testing the TTPForge project"

ARG GITHUB_CLI_VERSION=2.32.1
ARG SETUP_TOOLS_VER=65.5.1 # Avoid vulnerable versions

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

# Copy mage_completion.sh
COPY .devcontainer/bash/files/mage_completion.sh /etc/bash_completion.d/

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

# Add the completion script to the user's .bashrc
RUN echo "source /etc/bash_completion.d/mage_completion.sh" >> $HOME/.bashrc

WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge
COPY --chown=ttpforge . .

# Setup Python environment, Go dependencies, and build
RUN python3 -m venv $HOME/venv \
    && . $HOME/venv/bin/activate \
    && pip3 install --upgrade pip \
    "setuptools>=${SETUP_TOOLS_VER}" \
    pre-commit \
    && pre-commit install-hooks \
    && go install github.com/magefile/mage@latest \
    && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
    && PATH=$HOME/venv/bin:$GOPATH/bin:$PATH \
    && mage installDeps \
    && mage \
    && go mod tidy \
    && go build -o ttpforge \
    && rm ttpforge

# Update PATH to include virtual environment and Go binaries
ENV PATH=$HOME/venv/bin:$GOPATH/bin:$PATH

# Add the completion script to the user's .bashrc
RUN echo '### Mage Autocomplete ###' >> $HOME/.bashrc \
    && echo '# This section will not run during a bats test.' >> $HOME/.bashrc \
    && echo 'if [[ $RUNNING_BATS_TEST != 1 ]]; then' >> $HOME/.bashrc \
    && echo '    # The functions below are required for mage autocomplete' >> $HOME/.bashrc \
    && echo '    _get_comp_words_by_ref() { :; }' >> $HOME/.bashrc \
    && echo '    __ltrim_colon_completions() { :; }' >> $HOME/.bashrc \
    && echo "    source \"/etc/bash_completion.d/mage_completion.sh\"" >> $HOME/.bashrc \
    && echo 'fi' >> $HOME/.bashrc

CMD ["bash"]
