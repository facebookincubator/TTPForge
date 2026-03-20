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
	"fmt"
	"regexp"
	"strings"
)

// OutputCheck is a condition that inspects the combined stdout+stderr
// output from the step that just ran.
type OutputCheck struct {
	// Command is captured here to detect when output fields are used with
	// a command check (CommandCheck should handle that case, not OutputCheck)
	Command           string `yaml:"command,omitempty"`
	OutputContains    string `yaml:"output_contains,omitempty"`
	OutputNotContains string `yaml:"output_not_contains,omitempty"`
	OutputRegex       string `yaml:"output_regex,omitempty"`
}

// IsNil returns true when all fields are empty or when a command is present
// (if a command is present, CommandCheck should handle this instead)
func (o *OutputCheck) IsNil() bool {
	// If a command is present, this should be handled by CommandCheck
	if o.Command != "" {
		return true
	}
	return o.OutputContains == "" && o.OutputNotContains == "" && o.OutputRegex == ""
}

// Verify checks the step output against the configured conditions
func (o *OutputCheck) Verify(ctx VerificationContext) error {
	output := ctx.StepOutput

	if o.OutputContains != "" {
		if !strings.Contains(output, o.OutputContains) {
			return fmt.Errorf("step output does not contain %q",
				o.OutputContains)
		}
	}

	if o.OutputNotContains != "" {
		if strings.Contains(output, o.OutputNotContains) {
			return fmt.Errorf("step output contains %q but should not",
				o.OutputNotContains)
		}
	}

	if o.OutputRegex != "" {
		matched, err := regexp.MatchString(o.OutputRegex, output)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", o.OutputRegex, err)
		}
		if !matched {
			return fmt.Errorf("step output does not match regex %q",
				o.OutputRegex)
		}
	}

	return nil
}
