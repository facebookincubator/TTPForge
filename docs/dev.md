# Developer Environment Setup

To get involved with this project,
[create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
and follow along.

---

## Dependencies

- [Install pre-commit](https://pre-commit.com/):

  ```bash
  pip3 install pre-commit
  ```

- [Install gvm](https://github.com/moovweb/gvm):

  ```bash
  bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
  source "${GVM_BIN}"
  ```

- [Install golang](https://go.dev/):

  ```bash
  source .gvm
  ```

- [Install Mage](https://magefile.org/):

  ```bash
  go install github.com/magefile/mage@latest
  ```

- [Install Docker](https://docs.docker.com/get-docker/)

---

## Configure environment

1. Setup go environment:

   ```bash
   source .gvm
   ```

1. Install pre-commit hooks:

   ```bash
   mage installPreCommitHooks
   ```

1. Run pre-commit hooks locally:

   ```bash
   mage runPreCommit
   ```

1. Compile warpgate:

   ```bash
   go build -o ttpforge
   ```
