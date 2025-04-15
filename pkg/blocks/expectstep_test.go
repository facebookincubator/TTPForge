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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

func createTestScript(t *testing.T, scriptContent string) (string, string) {
	tempDir, err := os.MkdirTemp("", "python-script-test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	scriptPath := filepath.Join(tempDir, "interactive.py")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err)
	return scriptPath, tempDir
}

func NewTestTTPExecutionContext(workDir string) TTPExecutionContext {
	return TTPExecutionContext{
		Vars: &TTPExecutionVars{
			WorkDir: workDir,
		},
	}
}

func TestYAMLUnmarshal(t *testing.T) {
	testCases := []struct {
		name              string
		yamlConfig        string
		expectedErr       bool
		expectedLen       int
		expectNil         bool
		wantValidateError bool
		expectedErrTxt    string
	}{
		{
			name: "Valid YAML with ExpectStep",
			yamlConfig: `
steps:
  - name: run_expect_script
    expect:
      inline: |
        echo "Hello"
      responses:
        - prompt: "Enter your name: "
          response: "John"
`,
			expectedErr: false,
			expectedLen: 1,
			expectNil:   false,
		},
		{
			name: "YAML without ExpectStep",
			yamlConfig: `
steps:
  - name: run_expect_script
`,
			expectedErr: false,
			expectedLen: 1,
			expectNil:   true,
		},
		{
			name:        "Empty YAML",
			yamlConfig:  ``,
			expectedErr: false,
			expectedLen: 0,
			expectNil:   true,
		},
		{
			name: "Test Unmarshal Expect Valid",
			yamlConfig: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      inline: |
        python3 interactive.py
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
			expectedErr: false,
			expectedLen: 1,
			expectNil:   false,
		},
		{
			name: "Test Unmarshal Expect No Inline",
			yamlConfig: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
			expectedErr:       false,
			expectedLen:       1,
			expectNil:         false,
			wantValidateError: true,
			expectedErrTxt:    "expectStep is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var config struct {
				Steps []struct {
					Name   string      `yaml:"name"`
					Expect *ExpectStep `yaml:"expect"`
				} `yaml:"steps"`
			}

			err := yaml.Unmarshal([]byte(tc.yamlConfig), &config)
			if (err != nil) != tc.expectedErr {
				t.Fatalf("Unmarshal() error = %v, expectedErr %v", err, tc.expectedErr)
			}
			require.Len(t, config.Steps, tc.expectedLen)
			if !tc.expectedErr && tc.expectedLen > 0 {
				if tc.expectNil {
					require.Nil(t, config.Steps[0].Expect)
				} else {
					require.NotNil(t, config.Steps[0].Expect)
				}
			}

			if tc.wantValidateError {
				validateErr := config.Steps[0].Expect.Validate(TTPExecutionContext{})
				if validateErr == nil || validateErr.Error() != tc.expectedErrTxt {
					t.Fatalf("Validate() error = %v, expectedErrTxt %v", validateErr, tc.expectedErrTxt)
				}
			}

			t.Logf("Unmarshaled config: %+v", config)
		})
	}
}

// Mock for the tcPExecutionContext and ExpectStep to avoid using the real system
type MocktcPExecutionContext struct {
	mock.Mock
}

// MockExpectStep is a mock implementation of an expect step.
type MockExpectStep struct {
	mock.Mock
}

// SSHServer represents a simple SSH server for testing purposes.
type SSHServer struct {
	Addr     string
	Config   *ssh.ServerConfig
	Listener net.Listener
}

// handleConn handles an incoming connection.
func (s *SSHServer) handleConn(conn net.Conn) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.Config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return
		}
		defer channel.Close()

		go func(in <-chan *ssh.Request) {
			for req := range in {
				switch req.Type {
				case "exec":
					req.Reply(true, nil)
					channel.Write([]byte("command output\n"))
					channel.Close()
				}
			}
		}(requests)
	}
}

// NewSSHServer creates a new SSH server with an in-memory key pair.
func NewSSHServer() (*SSHServer, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	signer, err := ssh.ParsePrivateKey(privatePEM)
	if err != nil {
		return nil, err
	}

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	server := &SSHServer{
		Addr:     listener.Addr().String(),
		Config:   config,
		Listener: listener,
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.handleConn(conn)
		}
	}()

	return server, nil
}

func expectNoError(t *testing.T) expect.ConsoleOpt {
	return expect.WithExpectObserver(
		func(matchers []expect.Matcher, buf string, err error) {
			if err == nil {
				return
			}
			if len(matchers) == 0 {
				t.Fatalf("Error occurred while matching %q: %s\n%s", buf, err, string(debug.Stack()))
			} else {
				var criteria []string
				for _, matcher := range matchers {
					if crit, ok := matcher.Criteria().([]string); ok {
						criteria = append(criteria, crit...)
					} else {
						criteria = append(criteria, "unknown criteria")
					}
				}
				t.Fatalf("Failed to find [%s] in %q: %s\n%s", strings.Join(criteria, ", "), buf, err, string(debug.Stack()))
			}
		},
	)
}

func sendNoError(t *testing.T) expect.ConsoleOpt {
	return expect.WithSendObserver(
		func(msg string, n int, err error) {
			if err != nil {
				t.Fatalf("Failed to send %q: %s\n%s", msg, err, string(debug.Stack()))
			}
			if len(msg) != n {
				t.Fatalf("Only sent %d of %d bytes for %q\n%s", n, len(msg), msg, string(debug.Stack()))
			}
		},
	)
}

func TestExpectStep(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name               string
		script             string
		content            string
		wantUnmarshalError bool
		wantValidateError  bool
		wantTemplateError  bool
		wantExecuteError   bool
		expectedErrTxt     string
		stepVars           map[string]string
	}{
		{
			name: "Test ExpectStep Execute With Output",
			script: `
print("Enter your name:")
name = input()
print(f"Hello, {name}!")
print("Enter your age:")
age = input()
print(f"You are {age} years old.")
`,
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      inline: |
        python3 interactive.py
      responses:
        - prompt: "Enter your name:"
          response: "John"
        - prompt: "Enter your age:"
          response: "30"
`,
		},
		{
			name: "Test ExpectStep with Chdir",
			script: `
import os
print("Current directory:", os.getcwd())
print("Enter a number:")
number = input()
print(f"You input {number}.")
`,
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      chdir: "/tmp"
      inline: |
        python3 interactive.py
      responses:
        - prompt: "Enter a number:"
          response: "30"
`,
		},
		{
			name: "Test templating",
			script: `
print("Enter your name:")
name = input()
print(f"Hello, {name}!")
print("Enter your age:")
age = input()
print(f"You are {age} years old.")
`,
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      chdir: "{[{.StepVars.dir}]}"
      inline: |
        python3 {[{.StepVars.script}]}
      responses:
        - prompt: "{[{.StepVars.prompt}]}"
          response: "{[{.StepVars.number}]}"
`,
			stepVars: map[string]string{
				"dir":    "/tmp",
				"script": "interactive.py",
				"prompt": "Enter a number:",
				"number": "30",
			},
		},
		{
			name: "Errors on missing template",
			content: `
steps:
  - name: run_expect_script
    description: "Run an expect script to interact with the command."
    expect:
      chdir: "{[{.StepVars.dir}]}"
      inline: |
        python3 {[{.StepVars.script}]}
      responses:
        - prompt: "{[{.StepVars.prompt}]}"
          response: "{[{.StepVars.number}]}"
`,
			wantTemplateError: true,
			expectedErrTxt:    "template: BasicStep:1:20: executing \"BasicStep\" at <.StepVars.script>: map has no entry for key \"script\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scriptPath, tempDir := createTestScript(t, tc.script)

			var steps struct {
				Steps []ExpectStep `yaml:"steps"`
			}

			// prep execution context
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.WorkDir = tempDir
			execCtx.Vars.StepVars = tc.stepVars

			// Unmarshal
			err := yaml.Unmarshal([]byte(tc.content), &steps)
			if err != nil {
				assert.Fail(t, "Failed to unmarshal test case content: %v", err)
				return
			}
			require.NoError(t, err)

			if len(steps.Steps) == 0 || steps.Steps[0].Expect == nil {
				assert.Fail(t, "Failed to unmarshal test case content")
				return
			}

			expectStep := &steps.Steps[0]
			require.NotNil(t, expectStep, "expectStep is nil")

			// validate and check error
			err = expectStep.Validate(execCtx)
			if tc.wantValidateError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			// template and check error
			err = expectStep.Template(execCtx)
			if tc.wantTemplateError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			if tc.name == "Test ExpectStep Execute With Output" {
				// Mock the command execution
				execCtx := NewTestTTPExecutionContext(tempDir)
				console, err := expect.NewConsole(expectNoError(t), sendNoError(t), expect.WithStdout(os.Stdout), expect.WithStdin(os.Stdin))
				require.NoError(t, err)
				defer console.Close()

				cmd := exec.Command("sh", "-c", "python3 "+scriptPath)
				cmd.Stdin = console.Tty()
				cmd.Stdout = console.Tty()
				cmd.Stderr = console.Tty()

				err = cmd.Start()
				require.NoError(t, err)

				done := make(chan struct{})

				// simulate console input
				go func() {
					defer close(done)
					time.Sleep(1 * time.Second)
					console.SendLine("John")
					time.Sleep(1 * time.Second)
					console.SendLine("30")
					time.Sleep(1 * time.Second)
					console.Tty().Close() // Close the tcY to signal EOF
				}()

				_, err = expectStep.Execute(execCtx)
				require.NoError(t, err)
				<-done

				output, err := console.ExpectEOF()
				require.NoError(t, err)

				// Check the output of the command execution
				normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")
				expectedSubstring1 := "Hello, John!\n"
				expectedSubstring2 := "You are 30 years old.\n"
				assert.Contains(t, normalizedOutput, expectedSubstring1)
				assert.Contains(t, normalizedOutput, expectedSubstring2)
			}

			if tc.name == "Test ExpectStep with Chdir" {
				// Mock the command execution
				execCtx := NewTestTTPExecutionContext(tempDir)
				console, err := expect.NewConsole(expectNoError(t), sendNoError(t), expect.WithStdout(os.Stdout), expect.WithStdin(os.Stdin))
				require.NoError(t, err)
				defer console.Close()

				cmd := exec.Command("sh", "-c", "python3 "+scriptPath)
				cmd.Stdin = console.Tty()
				cmd.Stdout = console.Tty()
				cmd.Stderr = console.Tty()

				if expectStep.Chdir != "" {
					cmd.Dir = expectStep.Chdir
				}

				err = cmd.Start()
				require.NoError(t, err)

				done := make(chan struct{})

				// simulate console input
				go func() {
					defer close(done)
					time.Sleep(1 * time.Second)
					console.SendLine("30")
					time.Sleep(1 * time.Second)
					console.Tty().Close() // Close the tcY to signal EOF
				}()

				_, err = expectStep.Execute(execCtx)
				require.NoError(t, err)
				<-done

				output, err := console.ExpectEOF()
				require.NoError(t, err)

				// Check the output of the command execution
				normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")
				// Get the actual current directory from the output
				lines := strings.Split(normalizedOutput, "\n")
				var actualDir string
				for _, line := range lines {
					if strings.HasPrefix(line, "Current directory: ") {
						actualDir = line
						break
					}
				}
				require.NotEmpty(t, actualDir, "Failed to get the actual current directory")
				expectedSubstring2 := "You input 30.\n"
				assert.Contains(t, normalizedOutput, actualDir)
				assert.Contains(t, normalizedOutput, expectedSubstring2)
			}

			if tc.name == "Test Templating" {
				// Mock the command execution
				execCtx := NewTestTTPExecutionContext(tempDir)
				console, err := expect.NewConsole(expectNoError(t), sendNoError(t), expect.WithStdout(os.Stdout), expect.WithStdin(os.Stdin))
				require.NoError(t, err)
				defer console.Close()

				cmd := exec.Command("sh", "-c", "python3 "+scriptPath)
				cmd.Stdin = console.Tty()
				cmd.Stdout = console.Tty()
				cmd.Stderr = console.Tty()

				if expectStep.Chdir != "" {
					cmd.Dir = expectStep.Chdir
				}

				err = cmd.Start()
				require.NoError(t, err)

				done := make(chan struct{})

				// simulate console input
				go func() {
					defer close(done)
					time.Sleep(1 * time.Second)
					console.SendLine("30")
					time.Sleep(1 * time.Second)
					console.Tty().Close() // Close the tcY to signal EOF
				}()

				_, err = expectStep.Execute(execCtx)
				require.NoError(t, err)
				<-done

				output, err := console.ExpectEOF()
				require.NoError(t, err)

				// Check the output of the command execution
				normalizedOutput := strings.ReplaceAll(output, "\r\n", "\n")
				// Get the actual current directory from the output
				lines := strings.Split(normalizedOutput, "\n")
				var actualDir string
				for _, line := range lines {
					if strings.HasPrefix(line, "Current directory: ") {
						actualDir = line
						break
					}
				}
				require.NotEmpty(t, actualDir, "Failed to get the actual current directory")
				expectedSubstring2 := "You input 30.\n"
				assert.Contains(t, normalizedOutput, actualDir)
				assert.Contains(t, normalizedOutput, expectedSubstring2)
			}
		})
	}
}

// Execute is a mock implementation of the Execute method.
func (m *MockExpectStep) Execute(ctx TTPExecutionContext) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestExpectSSH(t *testing.T) {
	testCases := []struct {
		name    string
		step    *ExpectStep
		wantErr bool
		mockRet string
		mockErr error
	}{
		{
			name: "Test_Unmarshal_Expect_Valid",
			step: &ExpectStep{
				Executor: "bash",
				Expect: &ExpectSpec{
					Inline: "echo 'hello world'",
					Responses: []Response{
						{Prompt: "hello", Response: "world"},
					},
				},
			},
			wantErr: false,
			mockRet: "success",
			mockErr: nil,
		},
		{
			name: "Test_Unmarshal_Expect_No_Inline",
			step: &ExpectStep{
				Executor: "bash",
				Expect: &ExpectSpec{
					Responses: []Response{
						{Prompt: "hello", Response: "world"},
					},
				},
			},
			wantErr: true,
			mockRet: "",
			mockErr: errors.New("no inline script"),
		},
		{
			name: "Test_ExpectStep_Execute_With_Output",
			step: &ExpectStep{
				Executor: "bash",
				Expect: &ExpectSpec{
					Inline: `
					sshpass -p Password123! ssh bobbo@k8s1`,
					Responses: []Response{
						{Prompt: "Welcome to Ubuntu", Response: "whoami"},
						{Prompt: "bobbo", Response: "exit"},
					},
				},
			},
			wantErr: false,
			mockRet: "command output",
			mockErr: nil,
		},
		{
			name: "Test_ExpectStep_Execute_Network_Device_With_Output",
			step: &ExpectStep{
				Executor: "bash",
				Expect: &ExpectSpec{
					Inline: `
					ssh dr01.lab6`,
					Responses: []Response{
						{Prompt: "(user@system.school01) password:", Response: "${SSH_PASSWORD}"},
						{Prompt: ">", Response: "?"},
						{Prompt: ">", Response: "start shell"},
					},
				},
			},
			wantErr: false,
			mockRet: "cron job executed",
			mockErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStep := new(MockExpectStep)
			mockStep.On("Execute", mock.Anything).Return(tc.mockRet, tc.mockErr)

			execCtx := TTPExecutionContext{
				Vars: &TTPExecutionVars{
					WorkDir: ".",
				},
			}
			fmt.Println("Executing command:", tc.step.Expect.Inline)
			_, err := mockStep.Execute(execCtx)
			if (err != nil) != tc.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tc.wantErr)
			}
			fmt.Println("Command execution complete")
		})
	}
}

// Mocking execCommand for testing purposes
var execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

type MockCommandExecutor struct {
	runFunc func() error
}

func (m MockCommandExecutor) Run() error {
	return m.runFunc()
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	output := os.Getenv("MOCK_RUN_OUTPUT")
	if output != "" {
		fmt.Fprint(os.Stdout, output)
	}

	errorOutput := os.Getenv("MOCK_RUN_ERROR")
	if errorOutput != "nil" {
		fmt.Fprint(os.Stderr, errorOutput)
		os.Exit(1)
	}

	os.Exit(0)
}

func TestCanBeUsedInCompositeAction(t *testing.T) {
	s := &ExpectStep{}
	if !s.CanBeUsedInCompositeAction() {
		t.Errorf("Expected CanBeUsedInCompositeAction to return true, got false")
	}
}
