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

package cmd_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/facebookincubator/ttpforge/cmd"
	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/spf13/afero"
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
				testRepoName + "//basic/basic-file.yaml",
			},
			expectedStdout: "Hello World\n",
		},
		{
			name:        "file-step-no-config",
			description: "verify that execution works with no config file specified",
			args: []string{
				testResourcesDir + "/repos/" + testRepoName + "/ttps/basic/basic-file.yaml",
			},
			expectedStdout: "Hello World\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdoutBuf, stderrBuf bytes.Buffer
			rc := cmd.BuildRootCommand(&cmd.Config{
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

func TestNoCleanupFlag(t *testing.T) {
	afs := afero.NewOsFs()
	testCases := []struct {
		name             string
		content          string
		execConfig       blocks.TTPExecutionConfig
		expectedDirExist bool
		wantError        bool
	}{
		{
			name: "Test No Cleanup Behavior - Directory Creation",
			content: `
---
name: test-cleanup
steps:
  - name: step_one
    inline: mkdir testDir
    cleanup:
      inline: rm -rf testDir`,
			execConfig: blocks.TTPExecutionConfig{
				NoCleanup: true,
			},
			expectedDirExist: true,
			wantError:        false,
		},
		{
			name: "Test Cleanup Behavior - Directory Deletion",
			content: `
---
name: test-cleanup-2
steps:
  - name: step_two
    inline: mkdir testDir2
    cleanup:
      inline: rm -rf testDir2`,
			execConfig: blocks.TTPExecutionConfig{
				NoCleanup: false,
			},
			expectedDirExist: false,
			wantError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temp directory to work within
			tempDir, err := afero.TempDir(afs, "", "testCleanup")
			require.NoError(t, err)

			// Update content to work within the temp directory
			tc.content = strings.ReplaceAll(tc.content, "mkdir ", "mkdir "+tempDir+"/")
			tc.content = strings.ReplaceAll(tc.content, "rm -rf ", "rm -rf "+tempDir+"/")

			// Render the templated TTP first
			ttp, err := blocks.RenderTemplatedTTP(tc.content, &tc.execConfig)
			require.NoError(t, err)

			// Handle potential error from RemoveAll within a deferred function
			defer func() {
				err := afs.RemoveAll(tempDir) // cleanup temp directory
				if err != nil {
					t.Errorf("failed to remove temp directory: %v", err)
				}
			}()

			_, err = ttp.RunSteps(tc.execConfig)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Determine which directory to check based on the test case content
			dirName := tempDir + "/testDir"
			if strings.Contains(tc.content, "testDir2") {
				dirName = tempDir + "/testDir2"
			}

			// Check if the directory exists
			dirExists, err := afero.DirExists(afs, dirName)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDirExist, dirExists)
		})
	}
}
