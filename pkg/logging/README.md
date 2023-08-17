# TTPForge/logging

The `logging` package is a part of the TTPForge.

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

### InitLog(Config)

```go
InitLog(Config) error
```

InitLog initializes TTPForge global logger

---

### L()

```go
L() *zap.SugaredLogger
```

L returns the global logger for ttpforge

---

## Installation

To use the TTPForge/logging package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/logging
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/logging"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/logging`:

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
License - see the [LICENSE](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
file for details.
