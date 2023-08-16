FROM golang:1.20.7-alpine

# Description of the image
LABEL org.opencontainers.image.description="Dev Container with dependencies and tools for building and testing the TTPForge project"

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
        openssl-dev \
        python3 \
        py3-pip \
        readline-dev \
        ruby-full \
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
    && curl -sSL "https://github.com/cli/cli/releases/download/v2.32.1/gh_2.32.1_linux_${ARCH}.tar.gz" | tar xz \
    && mv "gh_2.32.1_linux_${ARCH}/bin/gh" /usr/local/bin \
    && gh --version

# Switch to 'ttpforge' user
USER ttpforge
ENV HOME /home/ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin

# Copy the codebase
COPY --chown=ttpforge . /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Create a virtual environment and install pre-commit
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge
RUN /bin/sh -c '\
    python3 -m venv $HOME/venv && \
    . $HOME/venv/bin/activate && \
    pip3 install --upgrade pip && \
    pip3 install "setuptools>=65.5.1" && \
    pip3 install --upgrade pre-commit && \
    pre-commit install'

ENV PATH="/home/ttpforge/venv/bin:$PATH"

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install go dependencies
RUN /bin/sh -c " \
    go install github.com/magefile/mage@latest \
    && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
    && mage installDeps \
    && mage \
    && go mod tidy \
    && go build -o ttpforge \
    && rm ttpforge"

# Remove the copied repository content
RUN rm -rf /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

CMD ["sh"]
