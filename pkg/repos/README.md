# TTPForge/repos

The `repos` package is a part of the TTPForge.

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

### Spec.Load(afero.Fs)

```go
Load(afero.Fs) Repo, error
```

Load will clone a repository if necessary and valdiate
its configuration, making it usable to lookup TTPs

---

### repo.FindTTP(string)

```go
FindTTP(string) string, error
```

FindTTP locates a TTP if it exists in this repo

---

### repo.FindTemplate(string)

```go
FindTemplate(string) string, error
```

FindTemplate locates a template if it exists in this repo

---

### repo.GetFs()

```go
GetFs() afero.Fs
```

GetFs is a convenience function principally used by SubTTPs

---

## Installation

To use the TTPForge/repos package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/repos
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/repos"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/repos`:

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
