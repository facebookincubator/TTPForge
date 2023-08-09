# TTPForge/outputs

The `outputs` package is a part of the TTPForge.

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

### JSONFilter.Apply(string)

```go
Apply(string) string, error
```

Apply applies this filters to the target string
and produces a new string

---

### Parse(map[string]Spec, string)

```go
Parse(map[string]Spec, string) map[string]string, error
```

Parse uses provided output specifications to extract output values
from the provided raw stdout string

**Parameters:**

specs: the specs for the outputs to be extracted
inStr: the raw stdout string from the step whose outputs will be extracted

**Returns:**

map[string]string: the output keys and values
error: an error if there is a problem

---

### Spec.Apply(string)

```go
Apply(string) string, error
```

Apply applies all filters in this output spec
to the target string in order, producing a new string

---

### Spec.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML is used to load specs from yaml files

---

## Installation

To use the TTPForge/outputs package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/outputs
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/outputs"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/outputs`:

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
