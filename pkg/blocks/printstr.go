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
	actionDefaults `yaml:",inline"`
	Message        string `yaml:"print_str,omitempty"`
}

// NewPrintStrAction creates a new PrintStrAction.
func NewPrintStrAction() *PrintStrAction {
	return &PrintStrAction{}
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

// Validate validates the step, checking for the necessary attributes and dependencies
func (s *PrintStrAction) Validate(_ TTPExecutionContext) error {
	if s.Message == "" {
		return fmt.Errorf("message field cannot be empty")
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (step *PrintStrAction) Template(execCtx TTPExecutionContext) error {
	var err error
	step.Message, err = execCtx.templateStep(step.Message)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
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
