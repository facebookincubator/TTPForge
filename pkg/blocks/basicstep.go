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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/outputs"
)

// DefaultExecutionTimeout is the default timeout for step execution.
const DefaultExecutionTimeout = 100 * time.Minute

// BasicStep is a type that represents a basic execution step.
type BasicStep struct {
	actionDefaults `yaml:",inline"`
	ExecutorName   string                  `yaml:"executor,omitempty"`
	Inline         string                  `yaml:"inline,flow"`
	Environment    map[string]string       `yaml:"env,omitempty"`
	Outputs        map[string]outputs.Spec `yaml:"outputs,omitempty"`
}

// NewBasicStep creates a new BasicStep instance with an initialized Act struct.
func NewBasicStep() *BasicStep {
	return &BasicStep{}
}

// IsNil checks if a step is considered empty or uninitialized.
func (b *BasicStep) IsNil() bool {
	switch b.Inline {
	case "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies.
func (b *BasicStep) Validate(execCtx TTPExecutionContext) error {
	// Check if Inline is provided
	if b.Inline == "" {
		return fmt.Errorf("inline must be provided")
	}

	// Set ExecutorName to "bash" if it is not provided
	if b.ExecutorName == "" {
		logging.L().Debug("defaulting to bash since executor was not provided")
		b.ExecutorName = ExecutorBash
	}

	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
func (b *BasicStep) Template(execCtx TTPExecutionContext) error {
	var err error
	b.Inline, err = execCtx.templateStep(b.Inline)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (b *BasicStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultExecutionTimeout)
	defer cancel()

	if b.Inline == "" {
		return nil, fmt.Errorf("empty inline value in Execute(...)")
	}

	executor := NewExecutor(b.ExecutorName, b.Inline, "", nil, b.Environment)
	result, err := executor.Execute(ctx, execCtx)
	if err != nil {
		return nil, err
	}
	result.Outputs, err = outputs.Parse(b.Outputs, result.Stdout)
	if err != nil {
		return nil, err
	}
	// Send stdout to the output variable
	if b.OutputVar != "" {
		execCtx.Vars.StepVars[b.OutputVar] = strings.TrimSuffix(result.Stdout, "\n")
	}
	return result, nil
}
