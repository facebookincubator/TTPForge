FROM golang:1.20.7

# Description of the image
LABEL org.opencontainers.image.description="Dev Container with dependencies and tools for building and testing the TTPForge project"

# Set build-time arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000

# Install dependencies required for RVM and Ruby
RUN DEBIAN_FRONTEND="noninteractive" apt-get update \
    && apt-get install -y bison curl git python3 python3-pip python3-venv shellcheck unzip vim \
    gpg2 autoconf automake libtool

# Install RVM
RUN gpg --keyserver hkp://pool.sks-keyservers.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3 7D2BAF1CF37B13E2069D6956105BD0E739499BDB \
    && \curl -sSL https://get.rvm.io | bash -s stable

# Use RVM to install Ruby 3.2.2
RUN /bin/bash -l -c "rvm requirements" \
    && /bin/bash -l -c "rvm install 3.2.2" \
    && /bin/bash -l -c "rvm use 3.2.2 --default"

# Ensure that the installed Ruby version is available system-wide
ENV PATH /usr/local/rvm/rubies/ruby-3.2.2/bin:$PATH

# Copy the codebase
COPY --chown=ttpforge . /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Create a virtual environment and install pre-commit
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge
RUN python3 -m venv /opt/venv || exit 1 \
    && . /opt/venv/bin/activate \
    && pip install --upgrade pre-commit || exit 1 \
    && pre-commit install || exit 1

ENV PATH="/opt/venv/bin:$PATH"

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh

# Install Terragrunt
RUN TERRAGRUNT_VERSION="0.45.2" \
    && curl -sL https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64 -o /tmp/terragrunt \
    && install -o root -g root -m 0755 /tmp/terragrunt /usr/local/bin/terragrunt \
    && rm /tmp/terragrunt

# Create the ttpforge user and change ownership of the necessary directories
RUN useradd -m -s /bin/bash ttpforge \
    && mkdir -p /home/ttpforge/go/src/github.com/facebookincubator/TTPForge \
    && chown -R ttpforge:ttpforge /home/ttpforge

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install go dependencies
RUN /bin/bash -c " \
    go install github.com/magefile/mage@latest \
    && go install mvdan.cc/sh/v3/cmd/shfmt@latest \
    # Install project dependencies
    && mage installDeps \
    && mage \
    # Run go mod tidy and build the project
    # to speed up builds
    && go mod tidy \
    && go build -o ttpforge \
    # Clean up the compiled bin
    && rm ttpforge"

# Remove the copied repository content
RUN rm -rf /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

CMD ["bash"]
