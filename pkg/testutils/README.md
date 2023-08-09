# TTPForge/testutils

The `testutils` package is a part of the TTPForge.

---

## Table of contents

- [Functions](#functions)
- [Installation](#installation)
- [Usage](#usage)
- [Tests](#tests)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### MakeAferoTestFs(map[string][]byte)

```go
MakeAferoTestFs(map[string][]byte) afero.Fs, error
```

MakeAferoTestFs is a convenience function that lets you
construct many directories and files in an afero in-memory filesystem
by passing a single path->contents map

---

## Installation

To use the TTPForge/testutils package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/testutils
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/testutils"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/testutils`:

```bash
go test -v
```

---

## Contributing

Pull requests are welcome. For major changes,
please open an issue first to discuss what
you would like to change.

---

## License

This project is licensed under the MIT
License - see the [LICENSE](../LICENSE)
file for details.
