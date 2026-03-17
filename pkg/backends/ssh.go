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
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// remoteShell abstracts shell-specific command building for different
// remote operating systems.
type remoteShell interface {
	setEnv(key, value string) string
	changeDir(path string) string
	quoteArg(arg string) string
	chainCommands(parts []string) string
}

// posixShell builds commands for POSIX-compatible shells (bash, sh, zsh).
type posixShell struct{}

func (s *posixShell) setEnv(key, value string) string {
	return fmt.Sprintf("export %s='%s'", key, strings.ReplaceAll(value, "'", "'\\''"))
}

func (s *posixShell) changeDir(path string) string {
	return fmt.Sprintf("cd '%s'", strings.ReplaceAll(path, "'", "'\\''"))
}

func (s *posixShell) quoteArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

func (s *posixShell) chainCommands(parts []string) string {
	return strings.Join(parts, " && ")
}

// powershellShell builds commands for PowerShell on remote Windows hosts.
type powershellShell struct{}

func (s *powershellShell) setEnv(key, value string) string {
	escaped := strings.ReplaceAll(value, "`", "``")
	escaped = strings.ReplaceAll(escaped, "\"", "`\"")
	escaped = strings.ReplaceAll(escaped, "$", "`$")
	return fmt.Sprintf("$env:%s = \"%s\"", key, escaped)
}

func (s *powershellShell) changeDir(path string) string {
	escaped := strings.ReplaceAll(path, "`", "``")
	escaped = strings.ReplaceAll(escaped, "\"", "`\"")
	return fmt.Sprintf("Set-Location \"%s\"", escaped)
}

func (s *powershellShell) quoteArg(arg string) string {
	escaped := strings.ReplaceAll(arg, "`", "``")
	escaped = strings.ReplaceAll(escaped, "\"", "`\"")
	escaped = strings.ReplaceAll(escaped, "$", "`$")
	return "\"" + escaped + "\""
}

func (s *powershellShell) chainCommands(parts []string) string {
	return strings.Join(parts, "; ")
}

// cmdShell builds commands for cmd.exe on remote Windows hosts.
type cmdShell struct{}

func (s *cmdShell) setEnv(key, value string) string {
	return fmt.Sprintf("set %s=%s", key, value)
}

func (s *cmdShell) changeDir(path string) string {
	escaped := strings.ReplaceAll(path, "\"", "\\\"")
	return fmt.Sprintf("cd /d \"%s\"", escaped)
}

func (s *cmdShell) quoteArg(arg string) string {
	escaped := strings.ReplaceAll(arg, "\"", "\\\"")
	return "\"" + escaped + "\""
}

func (s *cmdShell) chainCommands(parts []string) string {
	return strings.Join(parts, " & ")
}

// shellForType returns the remoteShell implementation for the given shell type.
func shellForType(shellType string) remoteShell {
	switch shellType {
	case "powershell":
		return &powershellShell{}
	case "cmd":
		return &cmdShell{}
	default:
		return &posixShell{}
	}
}

// SSHBackend implements ExecutionBackend over an SSH connection.
type SSHBackend struct {
	client        *ssh.Client
	bastionClient *ssh.Client // non-nil when connected via jump host
	agentConn     net.Conn    // non-nil when using agent auth
	sftpClient    *sftp.Client
	fs            afero.Fs
	fsOnce        sync.Once
	fsErr         error
	shell         remoteShell
	shellType     string
}

// NewSSHBackend creates a new SSH backend from the given remote config.
func NewSSHBackend(cfg *RemoteConfig) (*SSHBackend, error) {
	authMethods, agentConn, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH auth methods: %w", err)
	}
	// Ensure agent connection is closed if we fail before returning
	// the SSHBackend that would take ownership of it.
	closeAgentOnErr := true
	defer func() {
		if closeAgentOnErr && agentConn != nil {
			agentConn.Close()
		}
	}()

	hostKeyCallback, err := buildHostKeyCallback(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build host key callback: %w", err)
	}

	port := cfg.Port
	if port == 0 {
		port = 22
	}

	user := cfg.User
	if user == "" {
		user = os.Getenv("USER")
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}

	target := net.JoinHostPort(cfg.Host, strconv.Itoa(port))

	var client *ssh.Client
	var bastionClient *ssh.Client
	if cfg.JumpHost != "" {
		client, bastionClient, err = dialViaJumpHost(cfg.JumpHost, target, sshConfig)
	} else {
		client, err = ssh.Dial("tcp", target, sshConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	logging.L().Infof("SSH connection established to %s", target)

	shellType := cfg.Shell
	if shellType == "" {
		shellType = "posix"
	}

	closeAgentOnErr = false
	return &SSHBackend{
		client:        client,
		bastionClient: bastionClient,
		agentConn:     agentConn,
		shell:         shellForType(shellType),
		shellType:     shellType,
	}, nil
}

// RunCommand executes a command on the remote host over SSH.
// If stdoutW or stderrW are non-nil, output is tee'd to the writer and
// a capture buffer simultaneously for real-time streaming.
func (b *SSHBackend) RunCommand(ctx context.Context, name string, stdin string, args []string, env []string, workDir string, stdoutW io.Writer, stderrW io.Writer) (string, string, error) {
	session, err := b.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Build shell-specific command parts using the configured shell builder.
	var cmdParts []string

	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			cmdParts = append(cmdParts, b.shell.setEnv(parts[0], parts[1]))
		}
	}

	if workDir != "" {
		cmdParts = append(cmdParts, b.shell.changeDir(workDir))
	}

	// Build the actual command
	var command string
	if len(args) > 0 {
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			quotedArgs[i] = b.shell.quoteArg(arg)
		}
		command = name + " " + strings.Join(quotedArgs, " ")
	} else {
		command = name
	}
	cmdParts = append(cmdParts, command)

	fullCmd := b.shell.chainCommands(cmdParts)

	if stdin != "" {
		session.Stdin = strings.NewReader(stdin)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	if stdoutW != nil {
		session.Stdout = io.MultiWriter(stdoutW, &stdoutBuf)
	} else {
		session.Stdout = &stdoutBuf
	}
	if stderrW != nil {
		session.Stderr = io.MultiWriter(stderrW, &stderrBuf)
	} else {
		session.Stderr = &stderrBuf
	}

	// Use a goroutine + context for cancellation
	done := make(chan error, 1)
	go func() {
		done <- session.Run(fullCmd)
	}()

	select {
	case err := <-done:
		return stdoutBuf.String(), stderrBuf.String(), err
	case <-ctx.Done():
		// Signal the remote process
		_ = session.Signal(ssh.SIGKILL)
		return stdoutBuf.String(), stderrBuf.String(), ctx.Err()
	}
}

// ShellType returns the configured shell type for this backend.
func (b *SSHBackend) ShellType() string {
	return b.shellType
}

// GetFs returns an SFTP-backed filesystem. The SFTP client is created
// lazily on first call so that SSH connections to hosts without SFTP
// support (e.g. network devices) still work for command-only steps.
// Thread-safe via sync.Once.
func (b *SSHBackend) GetFs() (afero.Fs, error) {
	b.fsOnce.Do(func() {
		sftpClient, err := sftp.NewClient(b.client)
		if err != nil {
			b.fsErr = fmt.Errorf("failed to create SFTP client: %w", err)
			return
		}
		b.sftpClient = sftpClient
		b.fs = NewSFTPFs(sftpClient)
	})
	return b.fs, b.fsErr
}

// KillProcess kills a process on the remote host.
func (b *SSHBackend) KillProcess(pid int) error {
	ctx := context.Background()
	var cmdName string
	var args []string

	switch b.shellType {
	case "powershell":
		cmdName = "Stop-Process"
		args = []string{"-Id", strconv.Itoa(pid), "-Force"}
	case "cmd":
		cmdName = "taskkill"
		args = []string{"/F", "/PID", strconv.Itoa(pid)}
	default:
		cmdName = "kill"
		args = []string{"-9", strconv.Itoa(pid)}
	}

	_, stderr, err := b.RunCommand(ctx, cmdName, "", args, nil, "", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to kill remote process %d: %s: %w", pid, stderr, err)
	}
	return nil
}

// FindProcessesByName returns PIDs of processes matching the name on the remote host.
func (b *SSHBackend) FindProcessesByName(name string) ([]int, error) {
	ctx := context.Background()
	var cmdName string
	var args []string

	switch b.shellType {
	case "powershell":
		cmdName = "powershell"
		args = []string{"-Command", fmt.Sprintf("(Get-Process -Name '%s' -ErrorAction SilentlyContinue).Id", name)}
	case "cmd":
		cmdName = "tasklist"
		args = []string{"/FI", fmt.Sprintf("IMAGENAME eq %s", name), "/NH", "/FO", "CSV"}
	default:
		cmdName = "pgrep"
		args = []string{"-x", name}
	}

	stdout, _, err := b.RunCommand(ctx, cmdName, "", args, nil, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("no process found with name: %s", name)
	}

	var pids []int
	for line := range strings.SplitSeq(strings.TrimSpace(stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// For cmd shell, tasklist with CSV format outputs lines like:
		//   "name.exe","1234","Console","1","12,345 K"
		// Extract the PID from the second CSV field.
		if b.shellType == "cmd" {
			fields := strings.Split(line, ",")
			if len(fields) < 2 {
				continue
			}
			line = strings.Trim(fields[1], "\"")
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// ProcessExists checks whether a process exists on the remote host.
func (b *SSHBackend) ProcessExists(pid int) (bool, error) {
	ctx := context.Background()
	var cmdName string
	var args []string

	switch b.shellType {
	case "powershell":
		cmdName = "powershell"
		args = []string{"-Command", fmt.Sprintf("Get-Process -Id %d -ErrorAction SilentlyContinue", pid)}
	case "cmd":
		cmdName = "tasklist"
		args = []string{"/FI", fmt.Sprintf("PID eq %d", pid), "/NH"}
	default:
		cmdName = "kill"
		args = []string{"-0", strconv.Itoa(pid)}
	}

	_, _, err := b.RunCommand(ctx, cmdName, "", args, nil, "", nil, nil)
	return err == nil, nil
}

// Close closes the SFTP, SSH, bastion, and agent connections.
func (b *SSHBackend) Close() error {
	var errs []string
	if b.sftpClient != nil {
		if err := b.sftpClient.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("sftp close: %v", err))
		}
	}
	if b.client != nil {
		if err := b.client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("ssh close: %v", err))
		}
	}
	if b.bastionClient != nil {
		if err := b.bastionClient.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("bastion close: %v", err))
		}
	}
	if b.agentConn != nil {
		if err := b.agentConn.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("agent conn close: %v", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing SSH backend: %s", strings.Join(errs, "; "))
	}
	return nil
}

// buildAuthMethods constructs SSH auth methods from the RemoteConfig.
// Returns the auth methods and, when using agent auth, the agent socket
// connection (which the caller must close when done).
func buildAuthMethods(cfg *RemoteConfig) ([]ssh.AuthMethod, net.Conn, error) {
	auth := cfg.Auth
	if auth == "" {
		auth = "agent"
	}

	switch auth {
	case "agent":
		sock := os.Getenv("SSH_AUTH_SOCK")
		if sock == "" {
			return nil, nil, fmt.Errorf("SSH_AUTH_SOCK is not set; cannot use agent auth")
		}
		conn, err := net.Dial("unix", sock) // #nosec G704 -- connecting to user's own SSH agent socket
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to SSH agent: %w", err)
		}
		agentClient := agent.NewClient(conn)
		return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, conn, nil

	case "key":
		if cfg.KeyFile == "" {
			return nil, nil, fmt.Errorf("key_file must be set when auth is 'key'")
		}
		keyData, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read key file %s: %w", cfg.KeyFile, err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		// Auto-detect certificate file (OpenSSH convention: key_file-cert.pub)
		certFile := cfg.KeyFile + "-cert.pub"
		if certData, err := os.ReadFile(certFile); err == nil {
			pubKey, _, _, _, err := ssh.ParseAuthorizedKey(certData)
			if err == nil {
				if cert, ok := pubKey.(*ssh.Certificate); ok {
					certSigner, err := ssh.NewCertSigner(cert, signer)
					if err == nil {
						logging.L().Infof("Using certificate auth with %s", certFile)
						return []ssh.AuthMethod{ssh.PublicKeys(certSigner)}, nil, nil
					}
				}
			}
			logging.L().Warnf("Found certificate file %s but could not parse it, falling back to key-only auth", certFile)
		}

		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil, nil

	case "password":
		if cfg.Password == "" {
			return nil, nil, fmt.Errorf("password must be set when auth is 'password'")
		}
		return []ssh.AuthMethod{ssh.Password(cfg.Password)}, nil, nil

	case "password_env":
		if cfg.PasswordEnv == "" {
			return nil, nil, fmt.Errorf("password_env must be set when auth is 'password_env'")
		}
		password := os.Getenv(cfg.PasswordEnv)
		if password == "" {
			return nil, nil, fmt.Errorf("environment variable %s is empty", cfg.PasswordEnv)
		}
		return []ssh.AuthMethod{ssh.Password(password)}, nil, nil

	default:
		return nil, nil, fmt.Errorf("unsupported auth method: %s", auth)
	}
}

// buildHostKeyCallback constructs the host key callback from config.
func buildHostKeyCallback(cfg *RemoteConfig) (ssh.HostKeyCallback, error) {
	if cfg.KnownHosts == "" {
		// #nosec G106 -- intentional for red team tooling
		// nosemgrep: go.lang.security.audit.crypto.insecure_ssh.avoid-ssh-insecure-ignore-host-key
		return ssh.InsecureIgnoreHostKey(), nil
	}
	callback, err := knownhosts.New(cfg.KnownHosts)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts from %s: %w", cfg.KnownHosts, err)
	}
	return callback, nil
}

// dialViaJumpHost connects to the target through a bastion/jump host.
// Returns both the target client and the bastion client so the caller
// can close the bastion when done.
func dialViaJumpHost(jumpHost, target string, config *ssh.ClientConfig) (targetClient *ssh.Client, bastionClient *ssh.Client, err error) {
	// Connect to the jump host first
	bastion, err := ssh.Dial("tcp", jumpHost, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to jump host %s: %w", jumpHost, err)
	}

	// Dial the target through the bastion
	conn, err := bastion.Dial("tcp", target)
	if err != nil {
		bastion.Close()
		return nil, nil, fmt.Errorf("failed to dial target %s via jump host: %w", target, err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		conn.Close()
		bastion.Close()
		return nil, nil, fmt.Errorf("failed to create client connection to %s: %w", target, err)
	}

	return ssh.NewClient(ncc, chans, reqs), bastion, nil
}
