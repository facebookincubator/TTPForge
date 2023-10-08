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
	"gopkg.in/yaml.v3"
)

// BasicStep is a type that represents a basic execution step.
type BasicStep struct {
	*Act        `yaml:",inline"`
	Executor    string                  `yaml:"executor,omitempty"`
	Inline      string                  `yaml:"inline,flow"`
	Environment map[string]string       `yaml:"env,omitempty"`
	Outputs     map[string]outputs.Spec `yaml:"outputs,omitempty"`
	CleanupStep CleanupAct              `yaml:"cleanup,omitempty"`
}

// NewBasicStep creates a new BasicStep instance with an initialized Act struct.
func NewBasicStep() *BasicStep {
	return &BasicStep{
		Act: &Act{
			Type: StepBasic,
		},
	}
}

// UnmarshalYAML custom unmarshaler for BasicStep to handle decoding from YAML.
func (b *BasicStep) UnmarshalYAML(node *yaml.Node) error {
	type BasicStepTmpl struct {
		Act         `yaml:",inline"`
		Executor    string                  `yaml:"executor,omitempty"`
		Inline      string                  `yaml:"inline,flow"`
		Environment map[string]string       `yaml:"env,omitempty"`
		Outputs     map[string]outputs.Spec `yaml:"outputs,omitempty"`
		CleanupStep yaml.Node               `yaml:"cleanup,omitempty"`
	}

	var tmpl BasicStepTmpl
	// there is an issue with strict fields not being managed https://github.com/go-yaml/yaml/issues/460
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	b.Act = &tmpl.Act
	b.Executor = tmpl.Executor
	b.Inline = tmpl.Inline
	b.Environment = tmpl.Environment
	b.Outputs = tmpl.Outputs

	if b.IsNil() {
		return b.ExplainInvalid()
	}

	// we do it piecemiel to build our struct
	if tmpl.CleanupStep.IsZero() || b.Type == StepCleanup {
		return nil
	}

	logging.L().Debugw("step", "name", tmpl.Name)
	cleanup, err := b.MakeCleanupStep(&tmpl.CleanupStep)
	logging.L().Debugw("step", "err", err)
	if err != nil {
		return err
	}

	b.CleanupStep = cleanup

	return nil
}

// Cleanup is an implementation of the CleanupAct interface's Cleanup method.
func (b *BasicStep) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	result, err := b.Execute(execCtx)
	if err != nil {
		return nil, err
	}
	return &result.ActResult, err
}

// GetCleanup returns the cleanup steps for a BasicStep.
func (b *BasicStep) GetCleanup() []CleanupAct {
	if b.CleanupStep != nil {
		return []CleanupAct{b.CleanupStep}
	}
	return []CleanupAct{}
}

// GetType returns the step type for a BasicStep.
func (b *BasicStep) GetType() StepType {
	return b.Type
}

// ExplainInvalid returns an error with an explanation of why a BasicStep is invalid.
func (b *BasicStep) ExplainInvalid() error {
	var err error
	if b.Inline == "" {
		err = fmt.Errorf("(inline) empty")
	}
	if b.Name != "" && err != nil {
		return fmt.Errorf("[!] invalid basicstep: [%s] %w", b.Name, err)
	}
	return err
}

// IsNil checks if a BasicStep is considered empty or uninitialized.
func (b *BasicStep) IsNil() bool {
	switch {
	case b.Act.IsNil():
		return true
	case b.Inline == "":
		return true
	default:
		return false
	}
}

// Validate validates the BasicStep, checking for the necessary attributes and dependencies.
func (b *BasicStep) Validate(execCtx TTPExecutionContext) error {
	// Validate Act
	if err := b.Act.Validate(); err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	// Check if Inline is provided
	if b.Inline == "" {
		err := errors.New("inline must be provided")
		logging.L().Error(zap.Error(err))
		return err
	}

	// Set Executor to "bash" if it is not provided
	if b.Executor == "" && b.Inline != "" {
		logging.L().Debug("defaulting to bash since executor was not provided")
		b.Executor = "bash"
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

	// Validate CleanupStep if it is not nil
	if b.CleanupStep != nil {
		if err := b.CleanupStep.Validate(execCtx); err != nil {
			logging.L().Errorw("error validating cleanup step", zap.Error(err))
			return err
		}
	}

	logging.L().Debugw("command found in path", "executor", b.Executor)

	return nil
}

// Execute runs the BasicStep and returns an error if any occur.
func (b *BasicStep) Execute(execCtx TTPExecutionContext) (*ExecutionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()

	logging.L().Info("========= Executing ==========")

	if b.Inline == "" {
		return nil, fmt.Errorf("empty inline value in Execute(...)")
	}

	result, err := b.executeBashStdin(ctx, execCtx)
	if err != nil {
		return nil, err
	}

	logging.L().Info("========= Done ==========")

	return result, nil
}

func (b *BasicStep) executeBashStdin(ptx context.Context, execCtx TTPExecutionContext) (*ExecutionResult, error) {

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

	cmd := b.prepareCommand(ctx, expandedEnvAsList, expandedStrs[0])

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

func (b *BasicStep) prepareCommand(ctx context.Context, envAsList []string, inline string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, b.Executor)
	cmd.Env = envAsList
	cmd.Dir = b.WorkDir
	cmd.Stdin = strings.NewReader(inline)

	return cmd
}
