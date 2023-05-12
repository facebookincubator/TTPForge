# Writing Tests

## Testing Template

Ensure unit and integration tests are written using the following template as a starting point:

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

Several examples of complete tests can be found throughout the TTPForge repo.

---

## Testing Standards

At a high level, the following testing standards are employed for TTPForge:

1. All exported functions should have a corresponding test.

1. Integration tests must be created to accompany new functionality and ensure
   that everything works as expected going forward and with existing logic.

1. If tests touch the filesystem, they should be built to execute in a temporary directory
   specific to that test. Several examples can be found in `file_test.go`.

### Unit Testing

Consider the following criteria when writing unit tests for the TTPForge:

1. **Functionality:** Unit tests should test a single function or method. The
   test should be isolated from dependencies where possible.

1. **Inputs and Outputs:** Describe what type of input the function requires
   and what type of output is expected.

1. **Edge Cases:** Include how to handle edge cases. This can include null
   inputs, empty strings, maximum and minimum values, etc.

1. **Mocking:** If the function depends on other services or functions, you
   may need to mock these dependencies for testing.

1. Error Handling: Explain how to test error handling within the function.

### Integration Testing

Consider the following criteria when writing integration tests for the TTPForge:

1. **Components:** Identify the components that are being tested together. This
   could be multiple functions, methods, or even whole modules or services.

1. **Data Flow:** Describe how data flows between the components and what the expected outcomes are.

1. **Setup and Tear Down:** Include instructions for setting up the necessary
   environment for the test and tearing it down after the test.

1. **Mocking:** Explain when and how to mock dependencies in integration tests.
   Unlike unit tests, integration tests may require fewer mocks and more actual services.

1. **Error Handling:** Explain how to test for errors at the integration level.
   This might include testing how one component handles another component's failure.

---

## Building Tests

Our tests should be built under a separate package from the one that the code under test lives in.

There are several reasons for this convention:

**Encapsulation and separation of concerns:**

- By keeping test code separate from production code, we maintain a clear
  separation of concerns, making the code easier to understand and maintain.

**Testing exported functionality:**

- Using external test packages encourages testing only the exported
  functionality, which is what consumers of your package will be using. This
  helps focus on the package's public API and ensures it behaves as expected.

**Preventing circular dependencies:**

- External test packages minimize the risk of creating circular dependencies,
  which can occur when you import the package under test within the test code.

  Example:

  Assume we have a package named `calculator`:

  ```go
  // calculator/calculator.go
  package calculator

  func Add(a, b int) int {
      return a + b
  }
  ```

  And we want to create a test for it.

  Here's how you might do it with an internal test package:

  ```go
  // calculator/calculator_test.go
  package calculator

  import "testing"

  func TestAdd(t *testing.T) {
      result := Add(2, 3)
      if result != 5 {
          t.Errorf("Expected 5, got %d", result)
      }
  }
  ```

  This works, but the problem is that it's testing the internal workings of the calculator package,
  rather than its exported API. Also, if you import another package that depends on calculator into
  this test file, you might create a circular dependency.

  To avoid this, you can use an external test package:

  ```go
  // calculator/calculator_test.go
  package calculator_test

  import (
      "testing"
      "example.com/myapp/calculator"
  )

  func TestAdd(t *testing.T) {
      result := calculator.Add(2, 3)
      if result != 5 {
          t.Errorf("Expected 5, got %d", result)
      }
  }
  ```

  In this case, we're importing the calculator package as if we were any other consumer,
  and testing only its public API. This encourages good encapsulation practices and prevents
  the possibility of creating a circular dependency.

  The \_test suffix in package calculator_test is a convention in Go for external test
  packages. Note that the test files for external test packages still reside in the same
  directory as the package under test.

**Avoiding test-only dependencies:**

- External test packages prevent test-only dependencies from being included in
  the compiled binary, reducing the size of the final binary and ensuring
  production code does not accidentally use test dependencies.

**Resources:**

- <https://www.red-gate.com/simple-talk/devops/testing/go-unit-tests-tips-from-the-trenches/>
- <https://segment.com/blog/5-advanced-testing-techniques-in-go/>

---

## Updating Existing Tests

If any updates are made to existing library code that change the existing
functionality (add, update, remove, etc.), tests should be updated to capture
these changes.

For example:

If an existing exported function is updated to support `github.com/spf13/
afero`, the associated tests for that function should be updated so that the
existing tests still work properly.
