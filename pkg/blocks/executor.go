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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/logging"
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

// ScriptExecutor executes TTP steps by passing script via stdin
type ScriptExecutor struct {
	Name        string
	Inline      string
	Environment map[string]string
}

// FileExecutor executes TTP steps by calling a script file or binary with arguments
type FileExecutor struct {
	Name        string
	FilePath    string
	Args        []string
	Environment map[string]string
}

// NewExecutor creates a new ScriptExecutor or FileExecutor based on the executorName
func NewExecutor(executorName string, inline string, filePath string, args []string, environment map[string]string) Executor {
	if filePath != "" {
		return &FileExecutor{Name: executorName, FilePath: filePath, Args: args, Environment: environment}
	}
	return &ScriptExecutor{Name: executorName, Inline: inline, Environment: environment}
}

func (e *ScriptExecutor) buildCommand(ctx context.Context) *exec.Cmd {
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
func (e *ScriptExecutor) Execute(ctx context.Context, execCtx TTPExecutionContext) (*ActResult, error) {
	// expand variables in command
	expandedInlines, err := execCtx.ExpandVariables([]string{e.Inline})
	if err != nil {
		return nil, err
	}

	body := expandedInlines[0]
	if e.Name == ExecutorPowershellOnLinux || e.Name == ExecutorPowershell {
		// Wrap the PowerShell command in a script block
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
	cmd.Dir = execCtx.Vars.WorkDir
	cmd.Stdin = strings.NewReader(body)

	return streamAndCapture(*cmd, execCtx.Cfg.Stdout, execCtx.Cfg.Stderr)
}

// Execute runs the binary with arguments
func (e *FileExecutor) Execute(ctx context.Context, execCtx TTPExecutionContext) (*ActResult, error) {
	// expand variables in command line arguments
	expandedArgs, err := execCtx.ExpandVariables(e.Args)
	if err != nil {
		return nil, err
	}

	// expand variables in environment
	envAsList := append(FetchEnv(e.Environment), os.Environ()...)
	expandedEnvAsList, err := execCtx.ExpandVariables(envAsList)
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if e.Name == ExecutorBinary {
		cmd = exec.CommandContext(ctx, e.FilePath, expandedArgs...)
	} else {
		args := append([]string{e.FilePath}, expandedArgs...)
		cmd = exec.CommandContext(ctx, e.Name, args...)
	}

	cmd.Env = expandedEnvAsList
	cmd.Dir = execCtx.Vars.WorkDir
	return streamAndCapture(*cmd, execCtx.Cfg.Stdout, execCtx.Cfg.Stderr)
}

// InferExecutor infers the executor based on the file extension and
// returns it as a string.
func InferExecutor(filePath string) string {
	ext := filepath.Ext(filePath)
	logging.L().Debugw("file extension inferred", "filepath", filePath, "ext", ext)
	switch ext {
	case ".sh":
		return ExecutorSh
	case ".py":
		return ExecutorPython
	case ".rb":
		return ExecutorRuby
	case ".pwsh", ".ps1":
		if runtime.GOOS == "windows" {
			return ExecutorPowershell
		}
		return ExecutorPowershellOnLinux
	case ".bat":
		return ExecutorCmd
	case "":
		return ExecutorBinary
	default:
		if runtime.GOOS == "windows" {
			return ExecutorCmd
		}
		return ExecutorSh
	}
}
