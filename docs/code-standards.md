# Code Standards

---

## Table of Contents

- [Testing](#testing)
  - [General Testing Guidelines](#general-testing-guidelines)
  - [Unit Testing](#unit-testing)
  - [Integration Testing](#integration-testing)
  - [Example Tests](#example-tests)
  - [Test Architecture](#test-architecture)
  - [Test Maintenance](#test-maintenance)
- [Documentation](#documentation)
  - [Documenting Exported Functions](#documenting-exported-functions)
  - [Documenting Exported Structs](#documenting-exported-structs)
  - [Documenting Exported Interfaces](#documenting-exported-interfaces)

---

## Testing

We employ [table-driven tests](https://semaphoreci.com/blog/table-driven-unit-tests-go) for the TTPForge.
This format facilitates testing a function with a variety of cases.

```go
package some_test

import (
 "testing"
)

func TestSomething(t *testing.T) {
 tests := []struct{
  name string
  // input and output go here
 }{
  {
   name: "something"
   // input and output go here
  },
  {
   name: "another thing"
   // input and output go here
  },
 }

 for _, tc := range testCases {
  t.Run(tc.name, func(t *testing.T) {
   // call function or method being tested
   // check outcome
  })
 }
}
```

Numerous examples of comprehensive tests can be found throughout the TTPForge repo.

### General Testing Guidelines

In a nutshell, TTPForge adheres to the following testing standards:

1. All exported functions should be paired with a corresponding test.

1. Any new functionality must be accompanied by integration tests to ensure
   compatibility with existing logic and future code.

1. If tests interact with the filesystem, they should execute within a temporary
   directory specific to that test. Refer to `pkg/file/file_test.go` for examples.

1. Test packages should be separate from the package under test to prevent
   test code from being included in the compiled binary. E.g., tests for
   `pkg/foo` should be `in pkg/foo_test`.

### Unit Testing

When crafting unit tests for the TTPForge, please heed the following criteria:

1. **Functionality:** Unit tests should be focused on a single function or method, isolating
   it from dependencies wherever possible.

1. **Inputs and Outputs:** Specify the function's required input types and
   expected output types.

1. **Edge Cases:** Provide handling strategies for edge cases, including null
   inputs, empty strings, extreme values, etc.

1. **Mocking:** Mock dependencies if the function under test relies on
   other services or functions.

1. **Error Handling:** Elucidate strategies for testing error handling within the function.

### Integration Testing

When formulating integration tests for TTPForge, keep the following points in mind:

1. **Components:** Identify the components being tested together. This could be a combination
   of functions, methods, or even modules or services.

1. **Data Flow:** Describe how data flows between the components and what the expected outcomes are.

1. **Setup and Tear Down:** Provide instructions for the setup and tear down of the test environment.

1. **Mocking:** Explain when and how to mock dependencies. Unlike unit tests, integration tests
   may demand more actual services and fewer mocks.

1. **Error Handling:** Explain how to test for errors at the integration level.
   This might include testing how one component handles another component's failure.

### Example Tests

When creating your tests, it's beneficial to provide examples of how to utilize your code.
This can be achieved by constructing a test prefixed with "Example".

In the following example, we demonstrate how to use a function called `FixCodeBlocks`:

````go
package docs_test

import (
  "fmt"
  "log"
  "os"
  "strings"

  "github.com/example/docs"
  "github.com/example/fileutils"
)

func ExampleFixCodeBlocks() {
  input := `Driver represents an interface to Google Chrome using go.

It contains a context.Context associated with this Driver and
Options for the execution of Google Chrome.

` + "```go" + `
browser, err := cdpchrome.Init(true, true)

if err != nil {
    fmt.Printf("failed to initialize a chrome browser: %v", err)
    return
}
` + "```"
 language := "go"

 // Create a temporary file
 tmpfile, err := os.CreateTemp("", "example.*.md")
 if err != nil {
    fmt.Printf("failed to create temp file: %v", err)
    return
 }

 defer os.Remove(tmpfile.Name()) // clean up

 // Write the input to the temp file
 if _, err := tmpfile.Write([]byte(input)); err != nil {
   fmt.Printf("failed to write to temp file: %v", err)
   return
 }

 if err := tmpfile.Close(); err != nil {
   fmt.Printf("failed to close temp file: %v", err)
   return
 }

 // Run the function
 file := fileutils.RealFile(tmpfile.Name())
 if err := docs.FixCodeBlocks(file, language); err != nil {
   fmt.Printf("failed to fix code blocks: %v", err)
   return
 }

 // Read the modified content
 content, err := os.ReadFile(tmpfile.Name())
 if err != nil {
   fmt.Printf("failed to read file: %v", err)
   return
 }

 // Print the result
 fmt.Println(strings.TrimSpace(string(content)))
 // Output:
 // Driver represents an interface to Google Chrome using go.
 //
 // It contains a context.Context associated with this Driver and
 // Options for the execution of Google Chrome.
 //
 // ```go
 // browser, err := cdpchrome.Init(true, true)
 //
 // if err != nil {
 //     log.Fatalf("failed to initialize a chrome browser: %v", err)
 // }
 // ```
}
````

Note: The use of the "Example" prefix in these tests indicates that their
primary purpose is to serve as executable documentation for your code. They
are designed to provide developers with clear, practical examples of how to
use your functions. When these Example tests are run, the actual output from
the function is compared with the expected output, as specified in the `// Output:`
comment. This comparison ensures that your examples stay accurate and up-to-date
as your API evolves over time, preventing documentation from becoming obsolete or
misleading. Remember, Example tests complement, but do not replace, traditional
unit and integration tests.

**Resource:** <https://go.dev/blog/examples>

### Test Architecture

Tests should reside in a separate package from the one containing the code under test.
We advocate for this convention due to several reasons:

**Encapsulation and separation of concerns:**

- Keeping test code distinct from production code simplifies comprehension and maintenance.

**Testing exported functionality:**

- External test packages promote testing of only the exported functionality,
  mirroring the package consumers' perspective. This emphasizes the public API's behavior.

**Avoiding test-only dependencies:**

- External test packages eliminate test-only dependencies from the compiled binary,
  reducing its size and warding off accidental usage in production code.

**Resources:**

- <https://www.red-gate.com/simple-talk/devops/testing/go-unit-tests-tips-from-the-trenches/>
- <https://segment.com/blog/5-advanced-testing-techniques-in-go/>

### Test Maintenance

Test maintenance is pivotal. If library code is updated in a way that modifies
functionality (add, update, remove, etc.), the corresponding tests should also be updated.

For instance, if an exported function is updated to support `github.com/spf13/afero`,
its associated tests should also be updated to ensure they still function correctly.

---

## Documentation

Automated package documentation is an essential part of our work. It benefits us in
terms of reducing manual effort and standardizing documentation. That's why we utilize
the following templates. They're designed to support autogen package docs, enabling
us to parse this format and create READMEs for each package automatically.

### Documenting Exported Functions

```go
// FindExportedFunctionsInPackage locates all exported functions in a specific Go
// package by parsing all non-test Go files in the package directory. It returns
// a slice of FuncInfo structs. Each struct contains the file path and the name of an
// exported function. If the package contains no exported functions, an
// error is returned.
//
// **Parameters:**
//
// pkgPath: A string representing the path to the directory containing the package
// to search for exported functions.
//
// **Returns:**
//
// []FuncInfo: A slice of FuncInfo structs, each containing the file path and the
// name of an exported function found in the package.
// error: An error if no exported functions are found.
```

### Documenting Exported Structs

```go
// LogInfo encapsulates parameters that aid in managing logging throughout
// a program.
//
// **Attributes:**
//
// Dir: A string denoting the directory where the log file is located.
// File: An afero.File object that represents the log file.
// FileName: A string denoting the log file's name.
// Path: A string indicating the full path to the log file.
type LogInfo struct {
  Dir      string
  File     afero.File
  FileName string
  Path     string
}
```

### Documenting Exported Interfaces

```go
// File is an interface representing a system file.
//
// **Methods:**
//
// Open: Opens the file, returns a io.ReadCloser and an error.
// Write: Writes contents to the file, returns an error.
// RemoveAll: Removes a file or directory at the specified path, returns an error.
// Stat: Retrieves the FileInfo for the specified file or directory, returns an os.FileInfo and an error.
// Remove: Removes the specified file or directory, returns an error.
type File interface {
  Open() (io.ReadCloser, error)
  Write(contents []byte, perm os.FileMode) error
  RemoveAll() error
  Stat() (os.FileInfo, error)
  Remove() error
}
```
