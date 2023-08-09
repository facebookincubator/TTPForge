# TTPForge/cmd

The `cmd` package is a part of the TTPForge.

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

### Execute(ExecOptions)

```go
Execute(ExecOptions)
```

Execute adds child commands to the root
command and sets flags appropriately.

---

### RunTTPCmd()

```go
RunTTPCmd() *cobra.Command
```

RunTTPCmd runs an input TTP.

---

## Installation

To use the TTPForge/cmd package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/cmd
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/cmd"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/cmd`:

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
