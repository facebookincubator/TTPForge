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
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/outputs"
	"go.uber.org/zap"
)

// These are all the different executors that could run
// our inline command
const (
	ExecutorPython     = "python3"
	ExecutorBash       = "bash"
	ExecutorSh         = "sh"
	ExecutorPowershell = "powershell"
	ExecutorRuby       = "ruby"
	ExecutorBinary     = "binary"
	ExecutorCmd        = "cmd.exe"
)

// BasicStep is a type that represents a basic execution step.
type BasicStep struct {
	actionDefaults `yaml:",inline"`
	Executor       string                  `yaml:"executor,omitempty"`
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

	// Set Executor to "bash" if it is not provided
	if b.Executor == "" && b.Inline != "" {
		logging.L().Debug("defaulting to bash since executor was not provided")
		b.Executor = ExecutorBash
	}

	// Return if Executor is ExecutorBinary
	if b.Executor == ExecutorBinary {
		return nil
	}

	// Check if the executor is in the system path
	if _, err := exec.LookPath(b.Executor); err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	logging.L().Debugw("command found in path", "executor", b.Executor)

	return nil
}

// Execute runs the step and returns an error if one occurs.
func (b *BasicStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()

	if b.Inline == "" {
		return nil, fmt.Errorf("empty inline value in Execute(...)")
	}

	result, err := b.executeBashStdin(ctx, execCtx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *BasicStep) executeBashStdin(ptx context.Context, execCtx TTPExecutionContext) (*ActResult, error) {

	ctx, cancel := context.WithCancel(ptx)
	defer cancel()

	// expand variables in command
	expandedStrs, err := execCtx.ExpandVariables([]string{b.Inline})
	if err != nil {
		return nil, err
	}

	// expand variables in environment
	envAsList := append(FetchEnv(b.Environment), os.Environ()...)
	expandedEnvAsList, err := execCtx.ExpandVariables(envAsList)
	if err != nil {
		return nil, err
	}

	cmd := b.prepareCommand(ctx, execCtx, expandedEnvAsList, expandedStrs[0])

	result, err := streamAndCapture(*cmd, execCtx.Cfg.Stdout, execCtx.Cfg.Stderr)
	if err != nil {
		return nil, err
	}
	result.Outputs, err = outputs.Parse(b.Outputs, result.Stdout)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (b *BasicStep) buildCommand(ctx context.Context, executor string) *exec.Cmd {
	if executor == ExecutorBash {
		return exec.CommandContext(ctx, executor, "-o", "errexit")
	}
	return exec.CommandContext(ctx, executor)
}

func (b *BasicStep) prepareCommand(ctx context.Context, execCtx TTPExecutionContext, envAsList []string, inline string) *exec.Cmd {
	cmd := b.buildCommand(ctx, b.Executor)
	cmd.Env = envAsList
	cmd.Dir = execCtx.WorkDir
	cmd.Stdin = strings.NewReader(inline)

	return cmd
}
