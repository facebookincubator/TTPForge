# Getting started as a developer

To get involved with this project,
[create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
and follow along.

---

## Install Golang

TTPForge is build and tested in
[Github Actions](https://github.com/features/actions) using the Golang version
from this configuration file:

https://github.com/facebookincubator/TTPForge/blob/main/.github/workflows/tests.yaml

It is recommended to use the same version when developing locally. You can
install this Golang version from the official Golang
[website](https://go.dev/doc/install).

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

Finally, you can run our integration tests against your binary with the command:

```bash
./integration_tests.sh ./ttpforge
```

## Github Actions CI/CD

When you submit your change as a pull request to our repository, a variety of
linting and testing workflows will be triggered. If you wish to run any of these
workflows locally to fix a failure, you can do so with the
[act](https://github.com/nektos/act) tool. For example, you can run the
markdownlint action as follows:

```bash
act -W .github/workflows/markdownlint.yaml
```

## Running Pre-Commit Locally

Several of the linters in this project may be used as pre-commit hooks if
desired - you can install and setup pre-commit according to the
[official instructions](https://pre-commit.com/).

For quick ad hoc runs, you may with to run pre-commit in a virtual environment:

```bash
python3 -m venv venv
. venv/bin/activate
pip3 install pre-commit
pre-commit run --all-files
```

You can also run pre-commit locally using `act`, as described in the previous
section.
