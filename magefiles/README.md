# TTPForge/magefiles

The `magefiles` are a part of TTPForge that provide the functionality that
a traditional `Makefile` does in a project. It provides the functionality to
build, test, and document the project. The `magefiles` are written in Go and
use the [Mage](https://magefile.org/) library to provide the functionality.

---

## Table of contents

- [Functions](#functions)
  - [Compile](#compile)
  - [InstallDeps](#installdeps)
  - [FindExportedFuncsWithoutTests](#findexportedfuncswithouttests)
  - [GeneratePackageDocs](#generatepackagedocs)
  - [RunPreCommit](#runprecommit)
  - [RunTests](#runtests)
  - [RunIntegrationTests](#runintegrationtests)
- [Installation](#installation)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### Compile

```go
Compile(release bool) error
```

`Compile` is used to compile the Go project with the help of `goreleaser`. The
function checks for the `GOOS` and `GOARCH` environment variables to determine
the operating system and architecture, respectively, for the compilation. If
these variables are not set, it defaults to the current system's OS and
architecture.

By using the `release` flag, you can control the behavior of the compilation
process:

- If `release` is set to `true`, it will compile all supported releases for TTPForge.

- If set to `false`, it will compile only the binary for the specified OS and
  architecture (based on the environment variables) or for the current system's
  OS and architecture (if the environment variables aren't set).

Here are some examples of how to use the supported environment variables to
specify the OS, architecture, and invoke the `Compile` function:

```bash
# Compile all supported OS and architectures and output to the dist directory.
mage Compile true

# Compile for macOS on arm64
GOOS=darwin \
GOARCH=arm64 \
mage Compile false

# Compile for Linux on amd64
GOOS=linux \
GOARCH=amd64 \
mage Compile false

# Compile for Windows on amd64
GOOS=windows \
GOARCH=amd64 \
mage Compile false
```

---

### InstallDeps

```go
InstallDeps() error
```

Installs the TTPForge's Go dependencies necessary for developing on the project.

---

### FindExportedFuncsWithoutTests

```go
FindExportedFuncsWithoutTests(pkg string) ([]string, error)
```

Identifies exported functions within a package that lack corresponding test functions.

---

### GeneratePackageDocs

```go
GeneratePackageDocs() error
```

Creates documentation for the various packages in TTPForge.

---

### RunPreCommit

```go
RunPreCommit() error
```

Updates, clears, and executes all pre-commit hooks locally.

---

### RunTests

```go
RunTests() error
```

Executes all TTPForge unit and integration tests.

---

### RunIntegrationTests

```go
RunIntegrationTests() error
```

Executes all TTPForge integration tests by extracting commands described in
README files of TTP examples and then executing them. Before running the tests,
it compiles the project's binary and ensures that it's accessible from the
system's PATH.

---

## Installation

To use the functions from `TTPForge/magefiles`, ensure you have Mage installed
and then simply invoke them using the mage command in the TTPForge directory.

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first to
discuss what you would like to change.

---

## License

This project is licensed under the MIT License - see the
[LICENSE](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
file for details.
