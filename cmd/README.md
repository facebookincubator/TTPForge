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

### Execute

```go
Execute()
```

Execute adds child commands to the root
command and sets flags appropriately.

---

### NewTTPBuilderCmd

```go
NewTTPBuilderCmd() *cobra.Command
```

NewTTPBuilderCmd creates a new TTP from a template using the
provided input to customize it.

---

### RunTTPCmd

```go
RunTTPCmd() *cobra.Command
```

RunTTPCmd runs an input TTP.

---

### WriteConfigToFile

```go
WriteConfigToFile(string) error
```

WriteConfigToFile writes the configuration data to a YAML file at the specified
filepath. It uses the yaml.Marshal function to convert the configuration struct
into YAML format, and then writes the resulting bytes to the file.

This function is a custom alternative to Viper's built-in WriteConfig method
to provide better control over the formatting of the output YAML file.

Params:
  - filepath: The path of the file where the configuration data will be saved.

Returns:
  - error: An error object if any issues occur during the marshaling or file
    writing process, otherwise nil.

---

## Installation

To use the TTPForge/cmd package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/ttpforge/facebookincubator/cmd
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/ttpforge/facebookincubator/cmd"
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
License - see the [LICENSE](../LICENSE)
file for details.
