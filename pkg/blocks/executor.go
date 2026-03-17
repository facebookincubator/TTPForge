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
	"io"
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

// resolveStreamWriters returns the stdout and stderr writers to pass to
// RunCommand for remote execution. If the execution context provides explicit
// writers they are used; otherwise default streaming writers are created.
// The returned cleanup function must be called after RunCommand to flush any
// trailing partial lines.
func resolveStreamWriters(execCtx TTPExecutionContext) (stdoutW, stderrW io.Writer, cleanup func()) {
	stdoutW = execCtx.Cfg.Stdout
	stderrW = execCtx.Cfg.Stderr

	var toClose []*bufferedWriter
	if stdoutW == nil {
		bw := defaultStreamWriter("[STDOUT] ")
		stdoutW = bw
		toClose = append(toClose, bw)
	}
	if stderrW == nil {
		bw := defaultStreamWriter("[STDERR] ")
		stderrW = bw
		toClose = append(toClose, bw)
	}
	cleanup = func() {
		for _, bw := range toClose {
			bw.Close()
		}
	}
	return stdoutW, stderrW, cleanup
}

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
		body = fmt.Sprintf("$ErrorActionPreference = 'Stop' ; &{%s}\n\n", body)
	}

	// Remote backend path: delegate to backend.RunCommand
	if execCtx.Backend != nil {
		// For remote execution, only pass explicitly declared env vars
		expandedEnvAsList, err := execCtx.ExpandVariables(FetchEnv(e.Environment))
		if err != nil {
			return nil, err
		}

		stdoutW, stderrW, flushWriters := resolveStreamWriters(execCtx)
		stdout, stderr, err := execCtx.Backend.RunCommand(ctx, e.Name, body, nil, expandedEnvAsList, execCtx.Vars.WorkDir, stdoutW, stderrW)
		flushWriters()
		if err != nil {
			return nil, err
		}
		return &ActResult{Stdout: stdout, Stderr: stderr}, nil
	}

	// Local path: validate executor exists at runtime
	if e.Name != ExecutorBinary {
		if _, err := exec.LookPath(e.Name); err != nil {
			return nil, fmt.Errorf("executor %q not found in PATH: %w", e.Name, err)
		}
		logging.L().Debugw("executor found in path", "executor", e.Name)
	}

	// expand variables in environment (include inherited env for local execution)
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

	// Remote backend path: delegate to backend.RunCommand
	if execCtx.Backend != nil {
		expandedEnvAsList, err := execCtx.ExpandVariables(FetchEnv(e.Environment))
		if err != nil {
			return nil, err
		}

		var name string
		var args []string
		if e.Name == ExecutorBinary {
			name = e.FilePath
			args = expandedArgs
		} else {
			name = e.Name
			args = append([]string{e.FilePath}, expandedArgs...)
		}

		stdoutW, stderrW, flushWriters := resolveStreamWriters(execCtx)
		stdout, stderr, err := execCtx.Backend.RunCommand(ctx, name, "", args, expandedEnvAsList, execCtx.Vars.WorkDir, stdoutW, stderrW)
		flushWriters()
		if err != nil {
			return nil, err
		}
		return &ActResult{Stdout: stdout, Stderr: stderr}, nil
	}

	// Local path: validate executor exists at runtime
	if e.Name != ExecutorBinary {
		if _, err := exec.LookPath(e.Name); err != nil {
			return nil, fmt.Errorf("executor %q not found in PATH: %w", e.Name, err)
		}
		logging.L().Debugw("executor found in path", "executor", e.Name)
	}

	// expand variables in environment (include inherited env for local execution)
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
