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

package blocks

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// PrintStrAction is used to print a string to the console
type PrintStrAction struct {
	actionDefaults `yaml:"-"`
	Message        string `yaml:"print_str,omitempty"`
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *PrintStrAction) IsNil() bool {
	switch {
	case s.Message == "":
		return true
	default:
		return false
	}
}

// Execute runs the step and returns an error if any occur.
func (s *PrintStrAction) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	// needs to be overwritable to capture output during testing
	stdout := execCtx.Cfg.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	expandedStrs, err := execCtx.ExpandVariables([]string{s.Message})
	if err != nil {
		return nil, err
	}
	var stdoutBuf bytes.Buffer
	multi := io.MultiWriter(stdout, &stdoutBuf)
	fmt.Fprintln(multi, expandedStrs[0])
	result := &ActResult{
		Stdout: stdoutBuf.String(),
	}
	return result, nil
}

// Validate validates the step
//
// **Returns:**
//
// error: An error if any validation checks fail.
func (s *PrintStrAction) Validate(execCtx TTPExecutionContext) error {
	if s.Message == "" {
		return fmt.Errorf("message field cannot be empty")
	}
	return nil
}
