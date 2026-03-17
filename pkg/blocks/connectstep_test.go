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

package blocks

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/backends"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// mockBackend is a minimal ExecutionBackend that records which commands
// were run against it, so tests can verify which backend was selected.
type mockBackend struct {
	mu       sync.Mutex
	name     string
	commands []string
	fs       afero.Fs
}

func newMockBackend(name string) *mockBackend {
	return &mockBackend{
		name: name,
		fs:   afero.NewMemMapFs(),
	}
}

func (m *mockBackend) RunCommand(_ context.Context, name string, _ string, args []string, _ []string, _ string, _ io.Writer, _ io.Writer) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cmd := name + " " + joinArgs(args)
	m.commands = append(m.commands, cmd)
	return "", "", nil
}

func joinArgs(args []string) string {
	result := ""
	for i, a := range args {
		if i > 0 {
			result += " "
		}
		result += a
	}
	return result
}

func (m *mockBackend) GetFs() (afero.Fs, error)                    { return m.fs, nil }
func (m *mockBackend) KillProcess(_ int) error                     { return nil }
func (m *mockBackend) FindProcessesByName(_ string) ([]int, error) { return nil, nil }
func (m *mockBackend) ProcessExists(_ int) (bool, error)           { return false, nil }
func (m *mockBackend) Close() error                                { return nil }

func (m *mockBackend) getCommands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.commands))
	copy(out, m.commands)
	return out
}

func (m *mockBackend) clearCommands() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = nil
}

// newExecCtxWithMockRemote creates a TTPExecutionContext with a
// ConnectionPool that has the named mock backend registered.
func newExecCtxWithMockRemote(name string, mock *mockBackend) TTPExecutionContext {
	pool := backends.NewConnectionPool()
	cfg := &backends.RemoteConfig{Host: name + ".example.com", Protocol: "ssh"}
	_ = pool.RegisterWithBackend(name, cfg, mock)
	ctx := NewTTPExecutionContext()
	ctx.ConnPool = pool
	return ctx
}

func TestConnectStepUnmarshal(t *testing.T) {
	content := `name: setup
connect:
  host: example.com
  user: root
  auth: key
  key_file: /tmp/key
  connection_name: myconn`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)

	cs, ok := s.action.(*ConnectStep)
	require.True(t, ok, "action should be a ConnectStep")
	assert.Equal(t, "example.com", cs.Host)
	assert.Equal(t, "root", cs.User)
	assert.Equal(t, "key", cs.Auth)
	assert.Equal(t, "/tmp/key", cs.KeyFile)
	assert.Equal(t, "myconn", cs.ConnectionName)
}

func TestConnectStepIsNil(t *testing.T) {
	s := &ConnectStep{}
	assert.True(t, s.IsNil(), "empty ConnectStep should be nil")

	s.ConnectionName = "test"
	assert.False(t, s.IsNil(), "ConnectStep with connection_name should not be nil")
}

func TestConnectStepValidate(t *testing.T) {
	testCases := []struct {
		name      string
		step      ConnectStep
		wantError bool
	}{
		{
			name: "valid",
			step: ConnectStep{
				Host:           "example.com",
				ConnectionName: "myconn",
			},
			wantError: false,
		},
		{
			name: "missing host",
			step: ConnectStep{
				ConnectionName: "myconn",
			},
			wantError: true,
		},
		{
			name: "missing connection_name",
			step: ConnectStep{
				Host: "example.com",
			},
			wantError: true,
		},
	}

	execCtx := NewTTPExecutionContext()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.step.Validate(execCtx)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRemoteStringUnmarshal(t *testing.T) {
	content := `name: run_remote
remote: myconn
inline: echo hello`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.Equal(t, "myconn", s.Remote)
}

func TestRemoteUndefinedConnectionErrors(t *testing.T) {
	// Create a step with remote: pointing to a name that doesn't exist
	var s Step
	content := `name: run_remote
remote: nonexistent
inline: echo hello`

	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)

	execCtx := NewTTPExecutionContext()
	err = s.Validate(execCtx)
	require.NoError(t, err)
	err = s.Template(execCtx)
	require.NoError(t, err)

	// Execute should fail because the connection pool doesn't have this name
	_, err = s.Execute(execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection pool is not initialized")
}

func TestCleanupDefaultInheritsStepRemote(t *testing.T) {
	content := `name: drop-payload
remote: target
create_file: /tmp/payload.sh
contents: echo hello
cleanup: default`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.Equal(t, "target", s.Remote)
	assert.True(t, s.isDefaultCleanup, "cleanup: default should set isDefaultCleanup")
	assert.Equal(t, "", s.cleanupRemote, "cleanup: default should not set cleanupRemote")
}

func TestCleanupDefaultExecutesOnRemoteBackend(t *testing.T) {
	// cleanup: default on a remote step should run the cleanup action
	// against the remote backend (inheriting the step's remote:).
	content := `name: drop-payload
remote: target
create_file: /tmp/cleanup_default_test_file
contents: hello
cleanup: default`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)

	mock := newMockBackend("target")
	execCtx := newExecCtxWithMockRemote("target", mock)

	// Execute: creates the file on the remote mock backend's filesystem
	_, err = s.Execute(execCtx)
	require.NoError(t, err)

	// The default cleanup for create_file is remove_path — it should
	// run on the same remote backend. Verify the backend's FS has the
	// file created, then run cleanup to verify it gets removed.
	exists, _ := afero.Exists(mock.fs, "/tmp/cleanup_default_test_file")
	assert.True(t, exists, "file should exist on remote backend after Execute")

	_, err = s.Cleanup(execCtx)
	require.NoError(t, err)

	exists, _ = afero.Exists(mock.fs, "/tmp/cleanup_default_test_file")
	assert.False(t, exists, "file should be removed from remote backend after Cleanup")
}

func TestCustomCleanupWithoutRemoteRunsLocally(t *testing.T) {
	// Custom cleanup without remote: should run locally, NOT on the
	// step's remote backend.
	content := `name: exfil
remote: target
inline: cat /etc/passwd
cleanup:
  inline: echo cleaned`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.False(t, s.isDefaultCleanup)
	assert.Equal(t, "", s.cleanupRemote)

	mock := newMockBackend("target")
	execCtx := newExecCtxWithMockRemote("target", mock)

	// Execute runs on the remote mock
	_, err = s.Execute(execCtx)
	require.NoError(t, err)
	assert.NotEmpty(t, mock.getCommands(), "Execute should have run on the remote backend")

	mock.clearCommands()

	// Cleanup should run locally (no remote:), so mock should NOT record
	// any new commands. The local inline execution may fail (no shell
	// configured in test env) but the key assertion is that it did NOT
	// dispatch to the remote backend.
	_, _ = s.Cleanup(execCtx)
	assert.Empty(t, mock.getCommands(), "custom cleanup without remote: should NOT run on the remote backend")
}

func TestCustomCleanupWithRemoteExecutesOnRemoteBackend(t *testing.T) {
	// Custom cleanup with remote: target should run the cleanup on the
	// named remote backend.
	content := `name: run-payload
remote: target
inline: /tmp/payload.sh
cleanup:
  remote: target
  inline: rm -f /tmp/payload.sh`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.Equal(t, "target", s.cleanupRemote)

	mock := newMockBackend("target")
	execCtx := newExecCtxWithMockRemote("target", mock)

	_, err = s.Execute(execCtx)
	require.NoError(t, err)

	mock.clearCommands()

	_, err = s.Cleanup(execCtx)
	require.NoError(t, err)
	assert.NotEmpty(t, mock.getCommands(), "custom cleanup with remote: should run on the remote backend")
}

func TestCustomCleanupWithDifferentRemote(t *testing.T) {
	// Step runs on target-a, but cleanup runs on target-b.
	content := `name: mixed-step
remote: target-a
inline: echo hello
cleanup:
  remote: target-b
  inline: echo cleanup`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.Equal(t, "target-a", s.Remote)
	assert.Equal(t, "target-b", s.cleanupRemote)

	mockA := newMockBackend("target-a")
	mockB := newMockBackend("target-b")

	pool := backends.NewConnectionPool()
	_ = pool.RegisterWithBackend("target-a", &backends.RemoteConfig{Host: "a.example.com", Protocol: "ssh"}, mockA)
	_ = pool.RegisterWithBackend("target-b", &backends.RemoteConfig{Host: "b.example.com", Protocol: "ssh"}, mockB)

	execCtx := NewTTPExecutionContext()
	execCtx.ConnPool = pool

	// Execute runs on target-a
	_, err = s.Execute(execCtx)
	require.NoError(t, err)
	assert.NotEmpty(t, mockA.getCommands(), "Execute should run on target-a")
	assert.Empty(t, mockB.getCommands(), "Execute should NOT run on target-b")

	// Cleanup runs on target-b
	_, err = s.Cleanup(execCtx)
	require.NoError(t, err)
	assert.NotEmpty(t, mockB.getCommands(), "Cleanup should run on target-b")
}

func TestCleanupRemoteUndefinedConnectionErrors(t *testing.T) {
	content := `name: run-payload
inline: echo hello
cleanup:
  remote: nonexistent
  inline: rm -f /tmp/payload.sh`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	assert.Equal(t, "nonexistent", s.cleanupRemote)

	// Cleanup should fail because the connection pool is not initialized
	execCtx := NewTTPExecutionContext()
	_, err = s.Cleanup(execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection pool is not initialized")
}

func TestCleanupRemoteUnregisteredConnectionErrors(t *testing.T) {
	content := `name: run-payload
inline: echo hello
cleanup:
  remote: nonexistent
  inline: rm -f /tmp/payload.sh`

	var s Step
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)

	// Pool exists but "nonexistent" is not registered
	mock := newMockBackend("other")
	execCtx := newExecCtxWithMockRemote("other", mock)

	_, err = s.Cleanup(execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no connection registered")
}

func TestConnectStepTemplate(t *testing.T) {
	s := &ConnectStep{
		Host:           "{[{.StepVars.myhost}]}",
		User:           "{[{.StepVars.myuser}]}",
		ConnectionName: "conn1",
	}

	execCtx := NewTTPExecutionContext()
	execCtx.Vars.StepVars = map[string]string{
		"myhost": "resolved.example.com",
		"myuser": "admin",
	}

	err := s.Template(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "resolved.example.com", s.Host)
	assert.Equal(t, "admin", s.User)
}
