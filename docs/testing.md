# Writing Go Tests

All exported functions should have a corresponding test.

Use this test harness to begin writing tests for an exported function:

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
