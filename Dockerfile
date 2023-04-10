FROM golang:1.20.2

# Set build-time arguments for user and group IDs
ARG USER_ID=1000
ARG GROUP_ID=1000

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y curl git python3 python3-pip ruby-full shellcheck vim

# Create the ttpforge user and change ownership of the necessary directories
RUN useradd -m -s /bin/bash ttpforge && \
    mkdir -p /home/ttpforge/go/src/github.com/facebookincubator/TTPForge && \
    chown -R ttpforge:ttpforge /home/ttpforge/go

# Install go dependencies
RUN go install github.com/magefile/mage@latest && \
	go install mvdan.cc/sh/v3/cmd/shfmt@latest

# Copy the Go project files into the container
COPY --chown=ttpforge . /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Set the 'ttpforge' user as the default user
USER ttpforge
ENV GOPATH=/home/ttpforge/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:/home/ttpforge/.local/bin
WORKDIR /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

# Install pre-commit and update path with user-specific
# bin directory and run pre-commit install
RUN pip3 install --user pre-commit && \
    pre-commit install && \
    pre-commit

# Install project dependencies
RUN mage installDeps && \
    mage && \
    go mod tidy && \
    # Build the project to speed up builds in a new container
    go build -o ttpforge

# Remove the copied repository content
RUN rm -rf /home/ttpforge/go/src/github.com/facebookincubator/ttpforge

CMD ["bash"]
