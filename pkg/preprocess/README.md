# TTPForge/preprocess

The `preprocess` package is a part of the TTPForge.

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

### Parse([]byte)

```go
Parse([]byte) *Result, error
```

Parse handles early-stage processing of a TTP. It does two main tasks:
1) It lints the TTP.
2) It segments the TTP into "not steps" and "steps" sections, both crucial
for YAML unmarshalling and templating.

**Parameters:**

ttpBytes: A byte slice with the raw TTP for processing.

**Returns:**

*Result: Pointer to Result with parsed preamble and steps.
error: Error for parsing issues or top-level key arrangement problems.

---

## Installation

To use the TTPForge/preprocess package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/preprocess
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/preprocess"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/preprocess`:

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
