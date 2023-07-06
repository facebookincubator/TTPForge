# TTPForge/files

The `files` package is a part of the TTPForge.

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

### CreateDirIfNotExists(afero.Fs, string)

```go
CreateDirIfNotExists(afero.Fs, string) error
```

CreateDirIfNotExists checks if a directory exists at the
given path and creates it if it does not exist.
It returns an error if the directory could not be created.

**Parameters:**

fsys: An afero.Fs object representing the file system to operate on.

path: A string representing the path to the directory to check and
create if necessary.

**Returns:**

error: An error if the directory could not be created.

---

### ExecuteYAML(string, blocks.TTPExecutionContext)

```go
ExecuteYAML(string, blocks.TTPExecutionContext) *blocks.TTP, error
```

ExecuteYAML is the top-level function for executing a TTP defined
in a YAML file. It is exported for testing purposes,
and the returned TTP is required for assertion checks in tests.

**Parameters:**

yamlFile: A string representing the path to the YAML file containing
the TTP definition.
inventoryPaths: A slice of strings representing the inventory paths
to search for the TTP.

**Returns:**

*blocks.TTP: A pointer to a TTP struct containing the executed TTP
and its related information.

error: An error if the TTP execution fails or if the TTP file cannot be found.

---

### ExpandHomeDir(string)

```go
ExpandHomeDir(string) string
```

ExpandHomeDir expands the tilde character in a path to the user's home
directory. The function takes a string representing a path and checks if the
first character is a tilde (~). If it is, the function replaces the tilde
with the user's home directory. The path is returned unchanged if it does
not start with a tilde or if there's an error retrieving the user's home
directory.

Borrowed from
[here](https://github.com/l50/goutils/blob/e91b7c4e18e23c53e35d04fa7961a5a14ca8ef39/fileutils.go#L283-L318)

**Parameters:**

path: The string containing a path that may start with a tilde (~) character.

**Returns:**

string: The expanded path with the tilde replaced by the user's home
directory, or the original path if it does not start with a tilde or
there's an error retrieving the user's home directory.

---

### MkdirAllFS(afero.Fs, string, os.FileMode)

```go
MkdirAllFS(afero.Fs, string, os.FileMode) error
```

MkdirAllFS is a filesystem-agnostic version of os.MkdirAll.
It creates a directory named path, along with any necessary parents, and
returns nil, or else returns an error. The permission bits perm are used
for all directories that MkdirAll creates.

If path is already a directory, MkdirAll does nothing and returns nil.

**Parameters:**

fsys: An afero.Fs object representing the file system to operate on.

path: A string representing the path to the directory to create, including
any necessary parent directories.

perm: An os.FileMode representing the permission bits for the created
directories.

**Returns:**

error: An error if the directory could not be created.

---

### PathExistsInInventory(afero.Fs, string, []string)

```go
PathExistsInInventory(afero.Fs, string, []string) bool, error
```

PathExistsInInventory checks if a relative file path exists in any of the
inventory directories specified in the inventoryPaths parameter. The function
uses afero.Fs to operate on a filesystem.

**Parameters:**

fsys: An afero.Fs object representing the filesystem to search.

relPath: A string representing the relative path of the file to search
for in the inventory directories.

inventoryPaths: A []string containing the inventory directory paths
to search.

**Returns:**

bool: A boolean value indicating whether the file exists in any of the
inventory directories (true) or not (false).

error: An error if there is an issue checking the file's existence.

---

### TTPExists(afero.Fs, string, []string)

```go
TTPExists(afero.Fs, string, []string) bool, error
```

TTPExists checks if a TTP file exists in any of the inventory directories
specified in the inventoryPaths parameter.

**Parameters:**

fsys: An afero.Fs representing the file system to operate on.

ttpName: A string representing the name of the TTP file to search for
in the inventory directories.

inventoryPaths: A []string containing the inventory directory paths
to search.

**Returns:**

bool: A boolean value indicating whether the TTP file exists in any of
the inventory directories (true) or not (false).

error: An error if there is an issue checking the TTP file's existence.

---

### TemplateExists(afero.Fs, string, []string)

```go
TemplateExists(afero.Fs, string, []string) string, error
```

TemplateExists checks if a template file exists in a 'templates' folder
located in the parent directory of any of the inventory directories specified
in the inventoryPaths parameter. If the template file is found, it returns
the full path to the template file, otherwise, it returns an empty string.

**Parameters:**

fsys: An afero.Fs object representing the file system to operate on.

templatePath: A string representing the path of the template file to search
for in the 'templates' folder of the parent directory of each inventory
directory.

inventoryPaths: A []string containing the inventory directory
paths to search.

**Returns:**

fullPath: A string containing the full path to the template file if it
exists in the 'templates' folder of the parent directory of any of the
inventory directories, or an empty string if not found.

error: An error if there is an issue checking the template file's existence.

---

## Installation

To use the TTPForge/files package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/ttpforge/facebookincubator/files
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/ttpforge/facebookincubator/files"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/files`:

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
