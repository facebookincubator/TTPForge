//go:build linux
// +build linux

/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
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
	"path/filepath"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunLinux(t *testing.T) {
	const testResourcesDir = "test-resources"
	const testRepoName = "test-repo"
	testConfigFilePath := filepath.Join(testResourcesDir, "test-config.yaml")

	testCases := []struct {
		name           string
		description    string
		args           []string
		expectedStdout string
		wantError      bool
	}{
		{
			name:        "Check that requirements feature works on Linux",
			description: "Should pass since this file will only be built if we are on linux",
			args: []string{
				"-c",
				testConfigFilePath,
				testRepoName + "//requirements/linux-only.yaml",
			},
			expectedStdout: "just a placeholder - we are testing `requirements:`\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdoutBuf, stderrBuf bytes.Buffer
			rc := BuildRootCommand(&TestConfig{
				Stdout: &stdoutBuf,
				Stderr: &stderrBuf,
			})
			rc.SetArgs(append([]string{"run"}, tc.args...))
			err := rc.Execute()
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedStdout, stdoutBuf.String())
		})
	}
}
