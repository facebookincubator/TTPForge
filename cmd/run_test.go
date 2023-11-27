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
	"os"
	"path/filepath"

	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testResourcesDir = "test-resources"
	testRepoName     = "test-repo"
)

type runCmdTestCase struct {
	name           string
	description    string
	args           []string
	expectedStdout string
	wantError      bool
}

func checkRunCmdTestCase(t *testing.T, tc runCmdTestCase) {
	var stdoutBuf, stderrBuf bytes.Buffer
	rc := BuildRootCommand(&TestConfig{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})
	rc.SetArgs(append([]string{"run"}, tc.args...))
	err := rc.Execute()
	if tc.wantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, tc.expectedStdout, stdoutBuf.String())

}

func TestRun(t *testing.T) {
	testConfigFilePath := filepath.Join(testResourcesDir, "test-config.yaml")

	testCases := []runCmdTestCase{
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
			checkRunCmdTestCase(t, tc)
		})
	}
}

// TestRunPathArguments checks that
// referencing relative paths in `--arg` values
// when executing `ttpforge run` works as expected.
// One typically needs to specify `type: path` in
// the argument specification in order to get desired
// behavior.
func TestRunPathArguments(t *testing.T) {
	// in this test, we initially execute every test case from a
	// temporary directory, so we need the absolute path to the config
	testConfigFilePath, err := filepath.Abs(filepath.Join(testResourcesDir, "test-config.yaml"))
	require.NoError(t, err)
	// the file we will read from in the test cases
	targetFileName := "path-test.txt"

	// setup dummy file(s) for our test cases
	fsys := afero.NewOsFs()
	targetDir, err := afero.TempDir(fsys, "", "ttpforge-run-test")
	targetAbsPath := filepath.Join(targetDir, targetFileName)
	require.NoError(t, err)
	f, err := fsys.Create(targetAbsPath)
	require.NoError(t, err)
	defer fsys.Remove(targetFileName)
	f.Write([]byte("It worked!\n"))

	testCases := []runCmdTestCase{
		{
			name:        "Argument with `type: path` - Should Fail",
			description: "This should fail because the TTP's argument spec does not have `type: path`",
			args: []string{
				"-c",
				testConfigFilePath,
				testRepoName + "//args/path/without-path.yaml",
				"--arg",
				"target_path=" + targetFileName,
			},
			wantError: true,
		},
		{
			name:        "Argument with `type: path` - Should Succeed",
			description: "This should succeed because the TTP's argument spec has `type: path`",
			args: []string{
				"-c",
				testConfigFilePath,
				testRepoName + "//args/path/with-path.yaml",
				"--arg",
				"target_path=" + targetFileName,
			},
			expectedStdout: "It worked!\n",
		},
		{
			name:        "Argument with `type: path` - Absolute Path - Should Succeed",
			description: "Extra test case to check that `type: path` works with absolute paths",
			args: []string{
				"-c",
				testConfigFilePath,
				testRepoName + "//args/path/with-path.yaml",
				"--arg",
				"target_path=" + targetAbsPath,
			},
			expectedStdout: "It worked!\n",
		},
	}

	// change directory to the target dir so we can use relative paths
	// to refer to the target file
	wd, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(targetDir)
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(wd); err != nil {
			panic(err)
		}
	}()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkRunCmdTestCase(t, tc)
		})
	}
}
