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
	"os"
	"os/exec"
	"strings"
)

// These are all the different executors that could run
// our inline command
const (
	ExecutorPython            = "python3"
	ExecutorBash              = "bash"
	ExecutorSh                = "sh"
	ExecutorPowershell        = "powershell"
	ExecutorPowershellOnLinux = "pwsh"
	ExecutorRuby              = "ruby"
	ExecutorBinary            = "binary"
	ExecutorCmd               = "cmd.exe"
)

// Executor is an interface that defines the Execute method.
type Executor interface {
	Execute(ctx context.Context, execCtx TTPExecutionContext) (*ActResult, error)
}

// DefaultExecutor encapsulates logic to execute TTP steps
type DefaultExecutor struct {
	Name        string
	Inline      string
	Environment map[string]string
}

// NewExecutor creates a new DefaultExecutor
func NewExecutor(executorName string, inline string, environment map[string]string) Executor {
	return &DefaultExecutor{Name: executorName, Inline: inline, Environment: environment}
}

func (e *DefaultExecutor) buildCommand(ctx context.Context) *exec.Cmd {
	if e.Name == ExecutorPowershell || e.Name == ExecutorPowershellOnLinux {
		// @lint-ignore G204
		return exec.CommandContext(ctx, e.Name, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "-")
	}
	if e.Name == ExecutorBash {
		return exec.CommandContext(ctx, ExecutorBash, "-o", "errexit")
	}

	// @lint-ignore G204
	return exec.CommandContext(ctx, e.Name)
}

// Execute runs the command
func (e *DefaultExecutor) Execute(ctx context.Context, execCtx TTPExecutionContext) (*ActResult, error) {
	// expand variables in command
	expandedInlines, err := execCtx.ExpandVariables([]string{e.Inline})
	if err != nil {
		return nil, err
	}

	body := expandedInlines[0]
	if e.Name == ExecutorPowershellOnLinux || e.Name == ExecutorPowershell {
		// Write the TTP step to executor stdin
		body = fmt.Sprintf("&{%s}\n\n", body)
	}

	// expand variables in environment
	envAsList := append(FetchEnv(e.Environment), os.Environ()...)
	expandedEnvAsList, err := execCtx.ExpandVariables(envAsList)
	if err != nil {
		return nil, err
	}

	cmd := e.buildCommand(ctx)
	cmd.Env = expandedEnvAsList
	cmd.Dir = execCtx.WorkDir
	cmd.Stdin = strings.NewReader(body)

	return streamAndCapture(*cmd, execCtx.Cfg.Stdout, execCtx.Cfg.Stderr)
}
