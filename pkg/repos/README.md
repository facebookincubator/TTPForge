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

### NewRepoCollection(afero.Fs, []Spec, string)

```go
NewRepoCollection(afero.Fs, []Spec, string) RepoCollection, error
```

NewRepoCollection validates the provided repo specs
and assembles them into a RepoCollection

**Parameters:**

fsys: base file system (used for unit testing)
specs: a list of repo.Spec entries (usually from the config file)

**Returns:**

RepoCollection: assembled RepoCollection, or nil if there was an error
error: an error if there is a problem

---

### Spec.Load(afero.Fs, string)

```go
Load(afero.Fs, string) Repo, error
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

### repo.GetFullPath()

```go
GetFullPath() string
```

GetFullPath returns the repos full path
including the basePath that was passed
when it was constructed

---

### repo.GetName()

```go
GetName() string
```

GetName returns the repos name

---

### repo.ListTTPs()

```go
ListTTPs() []string, error
```

ListsTTPs lists the TTPs in this repo

---

### repoCollection.GetRepo(string)

```go
GetRepo(string) Repo, error
```

GetRepo retrieves a Repo reference
for a repo of the specified name,
or returns an error
if the repo is not in this collection

**Parameters:**

repoName: the repoistory name

**Returns:**

Repo: the located repo
error: an error if there is a problem

---

### repoCollection.ListTTPs()

```go
ListTTPs() []string, error
```

ListTTPs lists all TTPs in the RepoCollection

**Returns:**

[]string: the list of TTPs
error: an error if there is a problem

---

### repoCollection.ResolveTTPRef(string)

```go
ResolveTTPRef(string) Repo, string, error
```

ResolveTTPRef turns a provided TTP reference into
a Repo and absolute TTP file path

**Parameters:**

ttpRef: one of two things:

1. a reference of the form repo//path/to/ttp
2. an absolute or relative file path

**Returns:**

Repo: the located repo
string: the absolute path to the specified TTP
error: an error if there is a problem

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
License - see the [LICENSE](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
file for details.
