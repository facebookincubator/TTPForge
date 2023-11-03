# TTPForge/magefiles

`magefiles` provides utilities that would normally be managed
and executed with a `Makefile`. Instead of being written in the make language,
magefiles are crafted in Go and leverage the [Mage](https://magefile.org/) library.

---

## Table of contents

- [Functions](#functions)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### Compile()

```go
Compile() error
```

Compile compiles the Go project using goreleaser. The behavior is
controlled by the 'release' environment variable. If the GOOS and
GOARCH environment variables are not set, the function defaults
to the current system's OS and architecture.

**Environment Variables:**

release: Determines the compilation mode.

If "true", compiles all supported releases for TTPForge.
If "false", compiles only the binary for the specified OS
and architecture (based on GOOS and GOARCH) or the current
system's default if the vars aren't set.

GOOS: Target operating system for compilation. Defaults to the
current system's OS if not set.

GOARCH: Target architecture for compilation. Defaults to the
current system's architecture if not set.

Example usage:

```go
release=true mage compile # Compiles all supported releases for TTPForge
GOOS=darwin GOARCH=arm64 mage compile false # Compiles the binary for darwin/arm64
GOOS=linux GOARCH=amd64 mage compile false # Compiles the binary for linux/amd64
```

**Returns:**

error: An error if any issue occurs during compilation.

---

### InstallDeps()

```go
InstallDeps() error
```

InstallDeps installs the Go dependencies necessary for developing
on the project.

Example usage:

```go
mage installdeps
```

**Returns:**

error: An error if any issue occurs while trying to
install the dependencies.

---

### RunIntegrationTests()

```go
RunIntegrationTests() error
```

RunIntegrationTests executes all integration tests by extracting the commands
described in README files of TTP examples and then executing them.

Example usage:

```go
mage runintegrationtests
```

**Returns:**

error: An error if any issue occurs while running the tests.

---

### RunPreCommit()

```go
RunPreCommit() error
```

RunPreCommit updates, clears, and executes all pre-commit hooks
locally. The function follows a three-step process:

First, it updates the pre-commit hooks.
Next, it clears the pre-commit cache to ensure a clean environment.
Lastly, it executes all pre-commit hooks locally.

Example usage:

```go
mage runprecommit
```

**Returns:**

error: An error if any issue occurs at any of the three stages
of the process.

---

### RunTests()

```go
RunTests() error
```

RunTests executes all unit tests.

Example usage:

```go
mage runtests
```

**Returns:**

error: An error if any issue occurs while running the tests.

---

## Contributing

Pull requests are welcome. For major changes,
please open an issue first to discuss what
you would like to change.

---

## License

This project is licensed under the MIT
License - see the [LICENSE](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
file for details.
