FROM golang:1.20.7

# Description of the image
LABEL org.opencontainers.image.description="Dev Container with dependencies and tools for building and testing the TTPForge project"

# Set build-time arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000

# Install dependencies
RUN DEBIAN_FRONTEND="noninteractive" apt-get update \
    && apt-get install -y \
        bison \
        curl \
        git \
        shellcheck \
        unzip \
        vim \
        autoconf \
        automake \
        libtool \
        dirmngr \
        gpg \
        libreadline-dev \
        libsqlite3-dev \
        libssl-dev \
        python3 \
        python3-pip \
        python3.11-venv \
        zlib1g-dev \
        libyaml-dev \
        patch

# Create the ttpforge user and change ownership of the necessary directories
RUN useradd -m -s /bin/bash ttpforge \
    && mkdir -p /home/ttpforge/go/src/github.com/facebookincubator/TTPForge \
    && chown -R ttpforge:ttpforge /home/ttpforge

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh

# Switch to 'ttpforge' user
USER ttpforge
ENV HOME /home/ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin

# Install ASDF
RUN git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.12.0
RUN echo '. $HOME/.asdf/asdf.sh' >> ~/.bashrc

# Install Ruby using ASDF
RUN bash -c '\
    . $HOME/.asdf/asdf.sh && \
    asdf plugin-add ruby https://github.com/asdf-vm/asdf-ruby.git && \
    asdf install ruby 3.2.2 && \
    asdf global ruby 3.2.2'

# Copy the codebase
COPY --chown=ttpforge . /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Create a virtual environment and install pre-commit
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge
RUN bash -c '\
    python3 -m venv $HOME/venv && \
    . $HOME/venv/bin/activate && \
    pip3 install --upgrade pre-commit && \
    pre-commit install'

ENV PATH="/opt/venv/bin:$PATH"

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install go dependencies
RUN /bin/bash -c " \
    go install github.com/magefile/mage@latest \
    && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
    && mage installDeps \
    && mage \
    && go mod tidy \
    && go build -o ttpforge \
    && rm ttpforge"

# Remove the copied repository content
RUN rm -rf /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

CMD ["bash"]
