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

func TestRun(t *testing.T) {
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
			name:        "file-step",
			description: "check that a regular file step works",
			args: []string{
				"-c",
				testConfigFilePath,
				testRepoName + "//steps/file-step-demo.yaml",
			},
			expectedStdout: "Hello World\n",
		},
		{
			name:        "file-step-no-config",
			description: "verify that execution works with no config file specified",
			args: []string{
				testResourcesDir + "/repos/" + testRepoName + "/ttps/steps/file-step-demo.yaml",
			},
			expectedStdout: "Hello World\n",
		},
		{
			name:        "second-repo",
			description: "verify that execution of a TTP in a second repo succeeds",
			args: []string{
				"-c",
				testConfigFilePath,
				"another-repo//simple-inline.yaml",
			},
			expectedStdout: "simple inline was executed\ncleaning up simple inline\n",
		},
		{
			name:        "subttp-cleanup",
			description: "verify that execution of a subTTP with cleanup succeeds",
			args: []string{
				"-c",
				testConfigFilePath,
				"another-repo//sub-ttp-example/ttp.yaml",
			},
			expectedStdout: "subttp1_step_1\nsubttp1_step_2\nsubttp2_step_1\nsubttp2_step_1_cleanup\nsubttp1_step_2_cleanup\nsubttp1_step_1_cleanup\n",
			wantError:      true,
		},
		{
			name:        "dry-run-success",
			description: "validating a TTP with `--dry-run` should work for a syntactically valid TTP",
			args: []string{
				"-c",
				testConfigFilePath,
				"--dry-run",
				testRepoName + "//dry-run/dry-run-success.yaml",
			},
			expectedStdout: "",
		},
		{
			name:        "dry-run-fail",
			description: "validating a TTP with `--dry-run` should fail for a syntactically invalid TTP",
			args: []string{
				"-c",
				testConfigFilePath,
				"--dry-run",
				testRepoName + "//dry-run/dry-run-fail.yaml",
			},
			wantError: true,
		},
		{
			name:        "no-cleanup",
			description: "Using the no-cleanup flag should prevent cleanup",
			args: []string{
				"-c",
				testConfigFilePath,
				"--no-cleanup",
				"another-repo//simple-inline.yaml",
			},
			expectedStdout: "simple inline was executed\n",
		},
		{
			name:        "cleanup-stress-test",
			description: "run many different execute+cleanup combinations",
			args: []string{
				"-c",
				testConfigFilePath,
				"another-repo//cleanup-tests/stress-tests.yaml",
			},
			expectedStdout: "execute_step_1\nexecute_step_2\nexecute_step_3\nexecute_step_4\ncleanup_step_4\ncleanup_step_3\ncleanup_step_2\ncleanup_step_1\n",
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
