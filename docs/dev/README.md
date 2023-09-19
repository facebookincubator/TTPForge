# Getting started as a developer

To get involved with this project,
[create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
and follow along.

---

## Install Golang

We recommend building and testing TTPForge using Golang version `1.20.8`,
although older versions are also supported for compatibility reasons.
You can install this Golang version from the official Golang [website](https://go.dev/doc/install).
Alternatively, you can use `asdf` to manage your tool versions as described below -
this is highly recommended if you will be juggling multiple tool versions across
various projects.

## Using asdf to manage tool versions

The tool versions recommended for use with TTPForge are specified
in the `.tool-versions` file found in the repository root.

Setup asdf for usage with TTPForge as follows (these commands are distilled from the asdf [docs](https://asdf-vm.com/) - check there for updates if needed):

```bash
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.13.1
# we recommend adding the below line to your ~/.bashrc, ~/.zshrc etc
. "$HOME/.asdf/asdf.sh"
asdf plugin add golang https://github.com/asdf-community/asdf-golang.git
```

You can then install the correct Golang version for this project by
running the following command in the repo root:

```bash
asdf install golang
```

You now may need to run `rehash` (zsh) or `hash -r` (bash) - after that, you
can verify that your asdf version of Go is being used:

```bash
which go
go version
```

## Testing and Building TTPForge

With the appropriate Golang version installed as per the instructions above, you 
can now run our unit tests

```bash
go test ./...
```

and subsequently build your own copy of the TTPForge binary:

```bash
go build -o ttpforge
```

## Running Pre-Commit Locally (Optional)

This step is not required to build and run your own copy of TTPForge, 
but may help you iterate more quickly when responding to Pull Request 
checks that our repository automatically runs to test/lint new code.
Otherwise, you'll need to wait for Github Actions to vet your PR.

### Install pre-commit Dependencies

- [Install pre-commit](https://pre-commit.com/):

  ```bash
  python3 -m pip install --upgrade pip
  python3 -m pip install pre-commit
  ```

- [Install Mage](https://magefile.org/):

  ```bash
  go install github.com/magefile/mage@latest
  ```

- [Install Docker](https://docs.docker.com/get-docker/)

---

### Configure environment

1. Install dependencies:

   ```bash
   mage installDeps
   ```

1. Update and run pre-commit hooks locally:

   ```bash
   mage runPreCommit
   ```

1. Compile ttpforge:

   ```bash
   go build -o ttpforge
   ```

### Uninstall pre-commit

If you want to hold off on pre-commit until your code is at a certain point,
you can disable it locally:

```bash
pre-commit uninstall
```

Once you want to get feedback from pre-commit, reinstall it with:

```bash
pre-commit install
```
