/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package blocks

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	expect "github.com/l50/go-expect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func expectNoError(t *testing.T) expect.ConsoleOpt {
	return expect.WithExpectObserver(
		func(matchers []expect.Matcher, buf string, err error) {
			if err == nil {
				return
			}
			if len(matchers) == 0 {
				t.Fatalf("Error occurred while matching %q: %s\n%s", buf, err, string(debug.Stack()))
			} else {
				var criteria []string
				for _, matcher := range matchers {
					if crit, ok := matcher.Criteria().([]string); ok {
						criteria = append(criteria, crit...)
					} else {
						criteria = append(criteria, "unknown criteria")
					}
				}
				t.Fatalf("Failed to find [%s] in %q: %s\n%s", strings.Join(criteria, ", "), buf, err, string(debug.Stack()))
			}
		},
	)
}

func createTestScript(t *testing.T) (string, string) {
	tempDir, err := os.MkdirTemp("", "python-script-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	scriptContent := `
print("Enter your name:")
name = input()
print(f"Hello, {name}!")
print("Enter your age:")
age = input()
print(f"You are {age} years old.")
`
	scriptPath := filepath.Join(tempDir, "interactive.py")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err)
	return scriptPath, tempDir
}

func sendNoError(t *testing.T) expect.ConsoleOpt {
	return expect.WithSendObserver(
		func(msg string, n int, err error) {
			if err != nil {
				t.Fatalf("Failed to send %q: %s\n%s", msg, err, string(debug.Stack()))
			}
			if len(msg) != n {
				t.Fatalf("Only sent %d of %d bytes for %q\n%s", n, len(msg), msg, string(debug.Stack()))
			}
		},
	)
}

func NewTestTTPExecutionContext(workDir string) TTPExecutionContext {
	return TTPExecutionContext{
		WorkDir: workDir,
	}
}

func TestExpectStep(t *testing.T) {
	scriptPath, tempDir := createTestScript(t)

	testCases := []struct {
		name               string
		content            string
		wantUnmarshalError bool
		wantValidateError  bool
		wantExecuteError   bool
		expectedErrTxt     string
	}{
		{
			name: "Test Unmarshal Expect Valid",
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      inline: |
        python3 interactive.py
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
		},
		{
			name: "Test Unmarshal Expect No Inline",
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
			wantValidateError: true,
			expectedErrTxt:    "inline must be provided",
		},
		{
			name: "Test ExpectStep Execute With Output",
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      inline: |
        python3 interactive.py
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var steps struct {
				Steps []struct {
					Name        string      `yaml:"name"`
					Description string      `yaml:"description"`
					Expect      *ExpectStep `yaml:"expect"`
				} `yaml:"steps"`
			}

			err := yaml.Unmarshal([]byte(tc.content), &steps)
			if tc.wantUnmarshalError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if len(steps.Steps) == 0 || steps.Steps[0].Expect == nil {
				assert.Fail(t, "Failed to unmarshal test case content")
				return
			}

			expectStep := steps.Steps[0].Expect

			err = expectStep.Validate(NewTestTTPExecutionContext(tempDir))
			if tc.wantValidateError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			if tc.name == "Test ExpectStep Execute With Output" {
				// Mock the command execution
				execCtx := NewTestTTPExecutionContext(tempDir)

				// prepare the console
				console, err := expect.NewConsole(expectNoError(t), sendNoError(t), expect.WithStdout(os.Stdout), expect.WithStdin(os.Stdin))
				require.NoError(t, err)
				defer console.Close()

				cmd := exec.Command("sh", "-c", "python3 "+scriptPath)
				cmd.Stdin = console.Tty()
				cmd.Stdout = console.Tty()
				cmd.Stderr = console.Tty()

				err = cmd.Start()
				require.NoError(t, err)

				done := make(chan struct{})

				// simulate console input
				go func() {
					defer close(done)
					time.Sleep(1 * time.Second)
					console.SendLine("John")
					time.Sleep(1 * time.Second)
					console.SendLine("30")
					time.Sleep(1 * time.Second)
					console.Tty().Close() // Close the TTY to signal EOF
				}()

				// execute the step and check result
				result, err := expectStep.Execute(execCtx)
				if tc.wantExecuteError {
					assert.Equal(t, tc.expectedErrTxt, err.Error())
					return
				}
				require.NoError(t, err)
				assert.NotNil(t, result)

				<-done // wait for the goroutine to finish

				// Check the output of the command execution
				output, err := console.ExpectEOF()
				require.NoError(t, err)

				// Normalize line endings for comparison
				normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")
				expectedSubstring1 := "Hello, John!\n"
				expectedSubstring2 := "You are 30 years old.\n"
				assert.Contains(t, normalizedOutput, expectedSubstring1)
				assert.Contains(t, normalizedOutput, expectedSubstring2)
			}
		})
	}
}
