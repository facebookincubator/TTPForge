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

package backends

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdShellSetEnvOnWindows(t *testing.T) {
	shell := &cmdShell{}
	for _, delayedExpansion := range []string{"off", "on"} {
		for _, tc := range cmdEnvLiteralCases {
			t.Run(delayedExpansion+"/"+tc.name, func(t *testing.T) {
				assignment, err := shell.setEnv(cmdEnvTestName, tc.value)
				require.NoError(t, err)
				dir := t.TempDir()
				scriptPath := filepath.Join(dir, "env.cmd")
				script := shell.chainCommands([]string{assignment, "set " + cmdEnvTestName}) + "\r\n"
				require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0600))
				// The script contains only the test values above and runs in a
				// temporary directory if a quoting regression exposes operators.
				/* #nosec G204 */
				cmd := exec.CommandContext(t.Context(), "cmd.exe", "/d", "/q", "/v:"+delayedExpansion, "/c", scriptPath)
				cmd.Dir = dir
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", output)
				assert.Equal(t, cmdEnvTestName+"="+tc.value+"\r\n", string(output))
			})
		}
	}
}
