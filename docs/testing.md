# Writing Tests

## Testing Standards

At a high level, the following testing standards are employed for the TTPForge:

1. All exported functions should have a corresponding test.
2. Integration tests must be created to accompany new functionality and ensure that everything works as
   expected going forward and with existing logic.

## Testing Templates

Ensure tests are written using the following template as a starting point:

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

This template employs a separate test package.

It's considered a best practice to place tests in separate packages from the code they're testing.
By doing this, we're adhering to the Go convention of using external test packages.

There are several reasons for this convention:

**Encapsulation and separation of concerns:**

- By keeping test code separate from production code, we maintain a clear
  separation of concerns, making the code easier to understand and maintain.

**Testing exported functionality:**

- Using external test packages encourages testing only the exported functionality,
  which is what consumers of your package will be using. This helps focus on the package's
  public API and ensures it behaves as expected.

**Preventing circular dependencies:**

- External test packages minimize the risk of creating circular dependencies,
  which can occur when you import the package under test within the test code.

Avoiding test-only dependencies: External test packages prevent test-only dependencies from
being included in the compiled binary, reducing the size of the final binary and ensuring
production code does not accidentally use test dependencies.

**Resources:**

- <https://www.red-gate.com/simple-talk/devops/testing/go-unit-tests-tips-from-the-trenches/>
- <https://segment.com/blog/5-advanced-testing-techniques-in-go/>
