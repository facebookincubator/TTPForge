# TTPForge/targets

The `targets` package is a part of the TTPForge.

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

### ParseAndValidateTargets(TargetSpec)

```go
ParseAndValidateTargets(TargetSpec) map[string]interface{}, error
```

ParseAndValidateTargets takes a TargetSpec and processes the targets specified in it.
It then returns a map containing the processed targets. If the TargetSpec does not have any
valid targets specified, the resulting map will be empty.

Parameters:
targetSpec: The specifications for valid targets, extracted from the YAML configuration.

Returns:
processedTargets: A map of target names (like 'os', 'arch', etc.) to their values.
err: An error if any issues arise during the processing.

---

## Installation

To use the TTPForge/targets package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/targets
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/targets"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/targets`:

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
