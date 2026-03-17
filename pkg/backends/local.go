/*
Copyright © 2025-present, Meta Platforms, Inc. and affiliates
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

package backends

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/processutils"
	"github.com/spf13/afero"
)

// LocalBackend implements ExecutionBackend by running commands
// and accessing files on the local machine.
type LocalBackend struct {
	fs afero.Fs
}

// NewLocalBackend creates a new LocalBackend.
func NewLocalBackend() *LocalBackend {
	return &LocalBackend{
		fs: afero.NewOsFs(),
	}
}

// RunCommand executes a command locally using os/exec.
// If stdoutW or stderrW are non-nil, output is tee'd to the writer and
// a capture buffer simultaneously.
func (b *LocalBackend) RunCommand(ctx context.Context, name string, stdin string, args []string, env []string, workDir string, stdoutW io.Writer, stderrW io.Writer) (string, string, error) {
	// @lint-ignore G204
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = env
	}
	if workDir != "" {
		cmd.Dir = workDir
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	if stdoutW != nil {
		cmd.Stdout = io.MultiWriter(stdoutW, &stdoutBuf)
	} else {
		cmd.Stdout = &stdoutBuf
	}
	if stderrW != nil {
		cmd.Stderr = io.MultiWriter(stderrW, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
}

// GetFs returns the local OS filesystem.
func (b *LocalBackend) GetFs() (afero.Fs, error) {
	return b.fs, nil
}

// KillProcess kills a local process by PID.
func (b *LocalBackend) KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	return proc.Kill()
}

// FindProcessesByName returns PIDs of local processes matching the name.
func (b *LocalBackend) FindProcessesByName(name string) ([]int, error) {
	pids32, err := processutils.GetPIDsByName(name)
	if err != nil {
		return nil, err
	}
	pids := make([]int, len(pids32))
	for i, p := range pids32 {
		pids[i] = int(p)
	}
	return pids, nil
}

// ProcessExists checks if a local process exists.
func (b *LocalBackend) ProcessExists(pid int) (bool, error) {
	if err := processutils.VerifyPIDExists(pid); err != nil {
		return false, nil //nolint:nilerr // error means process doesn't exist, not a failure
	}
	return true, nil
}

// Close is a no-op for the local backend.
func (b *LocalBackend) Close() error {
	return nil
}
