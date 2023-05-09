FROM ghcr.io/facebookincubator/pt-ubuntu-vnc-zsh:latest

# Set build-time arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000

# Add the ubuntu user to sudoers
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' | sudo tee -a /etc/sudoers

# Install necessary dependencies
RUN DEBIAN_FRONTEND="noninteractive" sudo apt-get update \
    && sudo apt-get install -y bison curl expect git python3 python3-pip ruby-full shellcheck unzip vim

# Install GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && sudo apt-get update \
    && sudo apt-get install -y gh

# Install Terragrunt
RUN TERRAGRUNT_VERSION="0.45.2" \
    && curl -sL https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64 -o /tmp/terragrunt \
    && sudo install -o root -g root -m 0755 /tmp/terragrunt /usr/local/bin/terragrunt \
    && sudo rm /tmp/terragrunt

# Switch to root user
USER root

# Change the ubuntu user to the ttpforge user
RUN groupmod -n ttpforge ubuntu \
    && usermod -l ttpforge -g ttpforge ubuntu \
    && usermod -d /home/ttpforge -m ttpforge \
    && mkdir -p /home/ttpforge/go/src/github.com/csdXrStkMjREnjZGU/ttpforge \
    && chown -R ttpforge:ttpforge /home/ttpforge/go \
    && usermod -a -G sudo ttpforge

# Create the VNC password and configuration for the ttpforge user
RUN mkdir -p /home/ttpforge/.vnc \
    && chown ttpforge:ttpforge /home/ttpforge/.vnc \
    && chmod 0700 /home/ttpforge/.vnc \
    && su ttpforge -c "touch /home/ttpforge/.vnc/passwd" \
    && su ttpforge -c "chmod 0600 /home/ttpforge/.vnc/passwd"

# Update the VNC password for ttpforge user
RUN su ttpforge -c "echo 'ttpforge' | vncpasswd -f > /home/ttpforge/.vnc/passwd"

# Create the new docker-entrypoint.sh script with the correct username and home directory
RUN echo '#!/usr/bin/env zsh\n\
# docker-entrypoint.sh for docker image creation.\n\
set -x\n\
\n\
. /home/ttpforge/.zshrc\n\
start_vnc\n\
\n\
exec "$@"' > /run/docker-entrypoint.sh \
    && chmod +x /run/docker-entrypoint.sh

COPY --chown=ttpforge .. /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install pre-commit
RUN python3 -m pip install --upgrade pip \
    && python3 -m pip install --upgrade pre-commit \
    && pre-commit install

# Install gvm
RUN curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer | bash

# Install Go 1.20 using gvm
RUN /usr/bin/zsh -c "source .gvm \
    # Install go dependencies
    && go install github.com/magefile/mage@latest \
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

# Add the gvm source to the .zshrc file
RUN echo "\nsource /home/ttpforge/go/src/github.com/facebookincubator/ttpforge/.gvm" | tee -a /home/ttpforge/.zshrc

CMD ["zsh"]
