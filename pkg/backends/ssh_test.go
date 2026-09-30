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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cmdEnvTestName = "TTPFORGE_CMD_TEST"

var cmdEnvLiteralCases = []struct {
	name  string
	value string
	want  string
}{
	{name: "marker", value: "run-123:456", want: `set "TTPFORGE_CMD_TEST=run-123:456"`},
	{
		name:  "operators",
		value: `left&right|next<input>output^(group)`,
		want:  `set "TTPFORGE_CMD_TEST=left&right|next<input>output^(group)"`,
	},
	{
		name:  "leading-and-trailing-spaces",
		value: " leading and trailing  ",
		want:  `set "TTPFORGE_CMD_TEST= leading and trailing  "`,
	},
	{name: "equals", value: "left=right", want: `set "TTPFORGE_CMD_TEST=left=right"`},
	{
		name:  "literal-reference",
		value: "$forge.steps.test.stdout",
		want:  `set "TTPFORGE_CMD_TEST=$forge.steps.test.stdout"`,
	},
	{name: "path", value: `C:\Temp\script`, want: `set "TTPFORGE_CMD_TEST=C:\Temp\script"`},
}

func TestCmdShellSetEnv(t *testing.T) {
	shell := &cmdShell{}
	for _, tc := range cmdEnvLiteralCases {
		t.Run(tc.name, func(t *testing.T) {
			assignment, err := shell.setEnv(cmdEnvTestName, tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.want, assignment)
		})
	}
}

func TestCmdShellSetEnvRejectsUnsafeValues(t *testing.T) {
	type testCase struct {
		name  string
		value string
	}
	testCases := make([]testCase, 0, 37)
	testCases = append(testCases,
		testCase{name: "empty"},
		testCase{name: "quote", value: `secret"&echo injected`},
		testCase{name: "percent-expansion", value: "secret-%PATH%"},
		testCase{name: "delayed-expansion", value: "secret-!PATH!"},
		testCase{name: "delete", value: "secret-\x7f"},
	)
	for control := range 32 {
		testCases = append(testCases, testCase{
			name:  fmt.Sprintf("control-%02x", control),
			value: "secret-" + string(rune(control)),
		})
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assignment, err := (&cmdShell{}).setEnv(cmdEnvTestName, tc.value)
			require.ErrorContains(t, err, "cannot be passed literally")
			assert.ErrorContains(t, err, cmdEnvTestName)
			assert.NotContains(t, err.Error(), "secret")
			assert.Empty(t, assignment)
		})
	}
}

func TestSSHBackendRejectsInvalidEnvironmentNamesBeforeSession(t *testing.T) {
	for _, shellType := range []string{"posix", "powershell", "cmd"} {
		for _, name := range []string{
			"", "1INVALID", "BAD-NAME", "BAD&name", `BAD"name`, "BAD%name", "BAD!name", "BAD\nname",
			"X;id", "X$(id)", "X`id`",
		} {
			t.Run(shellType+"/"+fmt.Sprintf("%q", name), func(t *testing.T) {
				backend := &SSHBackend{shell: shellForType(shellType), shellType: shellType}
				stdout, stderr, err := backend.RunCommand(
					t.Context(), "unused", "", nil, []string{"SAFE=valid", name + "=secret"}, "", nil, nil,
				)
				require.ErrorContains(t, err, "invalid environment variable name")
				assert.NotContains(t, err.Error(), "secret")
				assert.Empty(t, stdout)
				assert.Empty(t, stderr)
			})
		}
	}
}

func TestSSHBackendRejectsInvalidCmdEnvironmentBeforeSession(t *testing.T) {
	for _, entry := range []string{
		"UNSAFE=secret-%PATH%",
		"EMPTY=",
		"BAD&name=secret",
	} {
		t.Run(entry, func(t *testing.T) {
			// A nil client would panic if RunCommand opened a session before
			// validating every entry, including one after a valid assignment.
			backend := &SSHBackend{shell: &cmdShell{}, shellType: "cmd"}
			stdout, stderr, err := backend.RunCommand(
				t.Context(), "cmd.exe", "echo executed", nil,
				[]string{"SAFE=valid", entry}, "", nil, nil,
			)
			require.ErrorContains(t, err, "environment variable")
			assert.NotContains(t, err.Error(), "secret")
			assert.Empty(t, stdout)
			assert.Empty(t, stderr)
		})
	}
}
