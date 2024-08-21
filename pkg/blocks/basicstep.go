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
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/outputs"
	"go.uber.org/zap"
)

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
	switch {
	case b.Inline == "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies.
func (b *BasicStep) Validate(execCtx TTPExecutionContext) error {
	// Check if Inline is provided
	if b.Inline == "" {
		err := errors.New("inline must be provided")
		logging.L().Error(zap.Error(err))
		return err
	}

	// Set ExecutorName to "bash" if it is not provided
	if b.ExecutorName == "" {
		logging.L().Debug("defaulting to bash since executor was not provided")
		b.ExecutorName = ExecutorBash
		return nil
	}

	// Return if ExecutorName is ExecutorBinary
	if b.ExecutorName == ExecutorBinary {
		return nil
	}

	// Check if the executor is in the system path
	if _, err := exec.LookPath(b.ExecutorName); err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	logging.L().Debugw("command found in path", "executor", b.ExecutorName)

	return nil
}

// Execute runs the step and returns an error if one occurs.
func (b *BasicStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()

	if b.Inline == "" {
		return nil, fmt.Errorf("empty inline value in Execute(...)")
	}

	executor := NewExecutor(b.ExecutorName, b.Inline, b.Environment)
	result, err := executor.Execute(ctx, execCtx)
	if err != nil {
		return nil, err
	}
	result.Outputs, err = outputs.Parse(b.Outputs, result.Stdout)
	if err != nil {
		return nil, err
	}
	return result, nil
}
