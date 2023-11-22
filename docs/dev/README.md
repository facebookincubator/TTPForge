# Getting started as a developer

To get involved with this project,
[create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
and follow along.

---

## Install Golang

We recommend building and testing TTPForge using Golang version `1.21.1`,
although older versions are also supported for compatibility reasons.
You can install this Golang version from the official
Golang [website](https://go.dev/doc/install).

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

Finally, you can run our integration tests against your binary
with the command:

```bash
./integration-tests.sh ./ttpforge
```
