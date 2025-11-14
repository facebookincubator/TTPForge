/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
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

package checks

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// CommandCheck is a condition that verifies command execution
// It can check both the exit code and output of a command
type CommandCheck struct {
	Command           string `yaml:"command"`
	ExpectExitCode    *int   `yaml:"expect_exit_code,omitempty"`
	OutputContains    string `yaml:"output_contains,omitempty"`
	OutputNotContains string `yaml:"output_not_contains,omitempty"`
	OutputRegex       string `yaml:"output_regex,omitempty"`
}

// IsNil checks if the condition is empty or uninitialized
func (c *CommandCheck) IsNil() bool {
	return c.Command == ""
}

// Verify executes the command and validates the results
func (c *CommandCheck) Verify(ctx VerificationContext) error {
	if c.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Execute the command using platform-appropriate shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// @lint-ignore G204
		cmd = exec.Command("cmd.exe", "/c", c.Command)
	} else {
		// @lint-ignore G204
		cmd = exec.Command("sh", "-c", c.Command)
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	exitCode := 0

	// Get the actual exit code
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// Command failed to execute at all
			return fmt.Errorf("failed to execute command %q: %w", c.Command, err)
		}
	}

	// Check exit code if specified
	expectedExitCode := 0
	if c.ExpectExitCode != nil {
		expectedExitCode = *c.ExpectExitCode
	}
	if exitCode != expectedExitCode {
		return fmt.Errorf("command %q exited with code %d, expected %d. Output: %s",
			c.Command, exitCode, expectedExitCode, outputStr)
	}

	// Check if output contains expected string
	if c.OutputContains != "" {
		if !strings.Contains(outputStr, c.OutputContains) {
			return fmt.Errorf("command %q output does not contain %q. Output: %s",
				c.Command, c.OutputContains, outputStr)
		}
	}

	// Check if output does not contain specified string
	if c.OutputNotContains != "" {
		if strings.Contains(outputStr, c.OutputNotContains) {
			return fmt.Errorf("command %q output contains %q but should not. Output: %s",
				c.Command, c.OutputNotContains, outputStr)
		}
	}

	// Check if output matches regex pattern
	if c.OutputRegex != "" {
		matched, err := regexp.MatchString(c.OutputRegex, outputStr)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", c.OutputRegex, err)
		}
		if !matched {
			return fmt.Errorf("command %q output does not match regex %q. Output: %s",
				c.Command, c.OutputRegex, outputStr)
		}
	}

	return nil
}
