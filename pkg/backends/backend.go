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
	"context"
	"io"

	"github.com/spf13/afero"
)

// RemoteConfig holds the configuration for a remote execution target.
// It is parsed from the `remote:` field on a TTPForge step.
type RemoteConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port,omitempty"`
	Protocol    string `yaml:"protocol,omitempty"`
	User        string `yaml:"user,omitempty"`
	Auth        string `yaml:"auth,omitempty"`
	KeyFile     string `yaml:"key_file,omitempty"`
	Password    string `yaml:"password,omitempty"`
	PasswordEnv string `yaml:"password_env,omitempty"`
	KnownHosts  string `yaml:"known_hosts,omitempty"`
	JumpHost    string `yaml:"jump_host,omitempty"`
	Shell       string `yaml:"shell,omitempty"` // "posix" (default), "powershell", or "cmd"
}

// ExecutionBackend abstracts the execution environment so that
// actions can run locally or on a remote host transparently.
type ExecutionBackend interface {
	// RunCommand executes a command with optional stdin, environment variables,
	// working directory, and streaming writers. If stdoutW or stderrW are non-nil,
	// output is tee'd to both the writer and the capture buffer. Returns captured
	// stdout, stderr, and any error.
	RunCommand(ctx context.Context, name string, stdin string, args []string, env []string, workDir string, stdoutW io.Writer, stderrW io.Writer) (stdout string, stderr string, err error)

	// GetFs returns a filesystem abstraction for file operations.
	// For local execution this is the OS filesystem; for SSH it is SFTP.
	// Returns an error if the filesystem is not available (e.g. SFTP
	// not supported on the remote host).
	GetFs() (afero.Fs, error)

	// KillProcess sends a kill signal to the process with the given PID.
	KillProcess(pid int) error

	// FindProcessesByName returns PIDs of processes matching the given name.
	FindProcessesByName(name string) ([]int, error)

	// ProcessExists checks whether a process with the given PID exists.
	ProcessExists(pid int) (bool, error)

	// Close releases any resources held by the backend (e.g., SSH connections).
	Close() error
}
