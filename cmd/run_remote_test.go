/*
Copyright (c) 2026-present, Meta Platforms, Inc. and affiliates
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

package cmd

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestRunRemoteForwardEnvPrecedence(t *testing.T) {
	t.Setenv("TTPFORGE_FORWARD_ENV", "")
	t.Setenv("TTPFORGE_TEST_PRECEDENCE", "runner")
	t.Setenv("TTPFORGE_TEST_REMOTE", "local")

	for _, executor := range []string{"inline", "file"} {
		for _, tc := range []struct {
			name      string
			forwarded string
			ttpEnv    string
			stepEnv   string
			want      string
		}{
			{name: "forwarded-only", want: "runner"},
			{name: "ttp-overrides-forwarded", ttpEnv: "ttp", want: "ttp"},
			{name: "step-overrides-both", ttpEnv: "ttp", stepEnv: "step", want: "step"},
			{
				name:      "forwarded-unknown-reference-stays-literal",
				forwarded: "literal-$forge.missing.value",
				want:      "literal-$forge.missing.value",
			},
			{
				name:      "forwarded-valid-reference-stays-literal",
				forwarded: "$forge.steps.connect_loopback.stdout",
				want:      "$forge.steps.connect_loopback.stdout",
			},
			{
				name:   "ttp-reference-expands",
				ttpEnv: "$forge.steps.connect_loopback.stdout",
				want:   "connected",
			},
			{
				name:    "step-reference-expands",
				ttpEnv:  "ttp",
				stepEnv: "$forge.steps.connect_loopback.stdout",
				want:    "connected",
			},
		} {
			t.Run(executor+"/"+tc.name, func(t *testing.T) {
				if tc.forwarded != "" {
					t.Setenv("TTPFORGE_TEST_PRECEDENCE", tc.forwarded)
				}
				dir := t.TempDir()
				repoConfig := filepath.Join(dir, "ttpforge-repo-config.yaml")
				require.NoError(t, os.WriteFile(repoConfig, []byte("ttp_search_paths: ['.']\n"), 0600))
				port, keyFile, knownHosts := startEnvTestSSHServer(t, dir)
				script := `printf '%s:%s\n' "$TTPFORGE_TEST_REMOTE" "$TTPFORGE_TEST_PRECEDENCE"`
				step := "    inline: |\n      " + script + "\n"
				if executor == "file" {
					scriptPath := filepath.Join(dir, "print-env.sh")
					require.NoError(t, os.WriteFile(scriptPath, []byte(script+"\n"), 0600))
					step = fmt.Sprintf("    file: %q\n", scriptPath)
				}
				if tc.stepEnv != "" {
					step += "    env:\n      TTPFORGE_TEST_PRECEDENCE: " + tc.stepEnv + "\n"
				}
				preamble := "name: remote-env-precedence\n"
				if tc.ttpEnv != "" {
					preamble += "env:\n  TTPFORGE_TEST_PRECEDENCE: " + tc.ttpEnv + "\n"
				}
				ttp := preamble + fmt.Sprintf(`steps:
  - name: connect_loopback
    connect:
      host: 127.0.0.1
      port: %d
      user: test
      auth: key
      key_file: %q
      known_hosts: %q
      connection_name: loopback
  - name: print-env
    remote: loopback
    executor: sh
    step_timeout: 5s
`, port, keyFile, knownHosts) + step
				ttpPath := filepath.Join(dir, "precedence.yaml")
				require.NoError(t, os.WriteFile(ttpPath, []byte(ttp), 0600))

				checkRunCmdTestCase(t, runCmdTestCase{
					args:           []string{"--forward-env", "TTPFORGE_TEST_PRECEDENCE", ttpPath},
					expectedStdout: "loopback:" + tc.want + "\n",
				})
			})
		}
	}
}

func startEnvTestSSHServer(t *testing.T, dir string) (int, string, string) {
	t.Helper()
	_, clientKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	require.NoError(t, err)
	keyBytes, err := x509.MarshalPKCS8PrivateKey(clientKey)
	require.NoError(t, err)
	keyFile := filepath.Join(dir, "client-key")
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	require.NoError(t, os.WriteFile(keyFile, keyPEM, 0600))

	_, hostKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	hostSigner, err := ssh.NewSignerFromKey(hostKey)
	require.NoError(t, err)
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if !bytes.Equal(key.Marshal(), clientSigner.PublicKey().Marshal()) {
				return nil, errors.New("unexpected client key")
			}
			return &ssh.Permissions{}, nil
		},
	}
	config.AddHostKey(hostSigner)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	knownHosts := filepath.Join(dir, "known-hosts")
	hostLine := fmt.Sprintf("[127.0.0.1]:%d %s", port, ssh.MarshalAuthorizedKey(hostSigner.PublicKey()))
	require.NoError(t, os.WriteFile(knownHosts, []byte(hostLine), 0600))

	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		stop := context.AfterFunc(t.Context(), func() { conn.Close() })
		defer stop()
		if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			done <- err
			return
		}
		done <- serveEnvTestSSHConnection(t.Context(), conn, config)
	}()
	t.Cleanup(func() {
		listener.Close()
		err := <-done
		if !errors.Is(err, net.ErrClosed) {
			assert.NoError(t, err)
		}
	})
	return port, keyFile, knownHosts
}

func serveEnvTestSSHConnection(ctx context.Context, conn net.Conn, config *ssh.ServerConfig) error {
	server, channels, requests, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return err
	}
	defer server.Close()
	go ssh.DiscardRequests(requests)
	for newChannel := range channels {
		if newChannel.ChannelType() != "session" {
			if err := newChannel.Reject(ssh.UnknownChannelType, "only sessions are supported"); err != nil {
				return err
			}
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return err
		}
		if err := executeEnvTestSSHSession(ctx, channel, requests); err != nil {
			return err
		}
	}
	return nil
}

func executeEnvTestSSHSession(ctx context.Context, channel ssh.Channel, requests <-chan *ssh.Request) error {
	defer channel.Close()
	for request := range requests {
		if request.Type != "exec" {
			if err := request.Reply(false, nil); err != nil {
				return err
			}
			continue
		}
		var payload struct{ Command string }
		if err := ssh.Unmarshal(request.Payload, &payload); err != nil {
			return err
		}
		if err := request.Reply(true, nil); err != nil {
			return err
		}
		// Execute the command supplied by the authenticated loopback test client.
		/* #nosec G204 */
		cmd := exec.CommandContext(ctx, "sh", "-c", payload.Command)
		// The marker proves the command ran on the SSH side, without inheriting runner variables.
		cmd.Env = []string{"PATH=/usr/bin:/bin", "TTPFORGE_TEST_REMOTE=loopback"}
		cmd.Stdin, cmd.Stdout, cmd.Stderr = channel, channel, channel.Stderr()
		cmd.WaitDelay = time.Second
		status := uint32(0)
		if err := cmd.Run(); err != nil {
			status = 1
		}
		_, err := channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
		return err
	}
	return errors.New("SSH session ended without a command")
}
