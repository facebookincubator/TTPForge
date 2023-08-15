# TTPForge/args

The `args` package is a part of the TTPForge.

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

### ParseAndValidate([]Spec, []string)

```go
ParseAndValidate([]Spec, []string) map[string]any, error
```

ParseAndValidate checks that the provided arguments
match the argument specifications for this TTP

**Parameters:**

specs: slice of argument Spec values loaded from the TTP yaml
argKvStrs: slice of arguments in "ARG_NAME=ARG_VALUE" format

**Returns:**

map[string]string: the parsed and validated argument key-value pairs
error: an error if there is a problem

---

## Installation

To use the TTPForge/args package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/args
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/args"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/args`:

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
