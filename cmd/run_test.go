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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/facebookincubator/ttpforge/cmd"
	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	testConfigFilePath := filepath.Join("test-resources", "test-config.yaml")
	testCases := []struct {
		name      string
		ttpRef    string
		wantError bool
	}{
		{
			name:   "basic-file",
			ttpRef: "test-repo//basic/basic-file.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc := cmd.BuildRootCommand()
			rc.SetArgs([]string{"run", "-c", testConfigFilePath, tc.ttpRef})
			err := rc.Execute()
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestRunWithoutConfig verifies
// that running TTPs still works without a top-level config file
func TestRunWithoutConfig(t *testing.T) {
	testCases := []struct {
		name      string
		ttpRef    string
		wantError bool
	}{
		{
			name:   "basic-file",
			ttpRef: "test-resources/repos/test-repo/ttps/basic/basic-file.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rc := cmd.BuildRootCommand()
			rc.SetArgs([]string{"run", tc.ttpRef})
			err := rc.Execute()
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func directoryExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestNoCleanupFlag(t *testing.T) {
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
			tempDir, err := os.MkdirTemp("", "testCleanup")
			if err != nil {
				t.Fatalf("failed to create temp directory: %v", err)
				return
			}
			defer os.RemoveAll(tempDir) // cleanup temp directory

			// Update content to work within the temp directory
			tc.content = strings.ReplaceAll(tc.content, "mkdir ", "mkdir "+tempDir+"/")
			tc.content = strings.ReplaceAll(tc.content, "rm -rf ", "rm -rf "+tempDir+"/")

			// Render the templated TTP first
			ttp, err := blocks.RenderTemplatedTTP(tc.content, &tc.execConfig)
			if err != nil {
				t.Fatalf("failed to render and unmarshal templated TTP: %v", err)
				return
			}

			_, err = ttp.RunSteps(tc.execConfig)
			if tc.wantError && err == nil {
				t.Error("expected an error from step execution but got none")
				return
			}
			if !tc.wantError && err != nil {
				t.Errorf("didn't expect an error from step execution but got: %s", err)
				return
			}

			// Determine which directory to check based on the test case content
			dirName := tempDir + "/testDir"
			if strings.Contains(tc.content, "testDir2") {
				dirName = tempDir + "/testDir2"
			}

			// Check if the directory exists
			if tc.expectedDirExist && !directoryExists(dirName) {
				t.Errorf("expected the directory '%s' to exist but it doesn't", dirName)
				return
			}
			if !tc.expectedDirExist && directoryExists(dirName) {
				t.Errorf("didn't expect the directory '%s' to exist but it does", dirName)
				return
			}
		})
	}
}

func TestCleanupDelayFlag(t *testing.T) {
	testCases := []struct {
		name             string
		content          string
		execConfig       blocks.TTPExecutionConfig
		expectedDirExist bool
		cleanupDelay     time.Duration
		wantError        bool
	}{
		{
			name: "Test Cleanup Delay Behavior - Directory Creation",
			content: `
---
name: test-cleanup-delay
steps:
  - name: step_one
    inline: mkdir -p TEMP_DIR_PLACEHOLDER/testDirDelay
    cleanup:
      inline: rm -rf TEMP_DIR_PLACEHOLDER/testDirDelay
`,
			execConfig: blocks.TTPExecutionConfig{
				CleanupDelaySeconds: 5, // 5 second delay
			},
			expectedDirExist: true,
			cleanupDelay:     5 * time.Second,
			wantError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temp directory to work within
			tempDir, err := os.MkdirTemp("", "testCleanupDelay")
			if err != nil {
				t.Fatalf("failed to create temp directory: %v", err)
				return
			}
			defer os.RemoveAll(tempDir) // cleanup temp directory

			// Prepare the directory name where operations are expected
			dirName := filepath.Join(tempDir, "testDirDelay")

			// Update content to work within the temp directory using filepath.Join to ensure platform compatibility
			tc.content = strings.ReplaceAll(tc.content, "TEMP_DIR_PLACEHOLDER", tempDir)

			// Render the templated TTP first
			ttp, err := blocks.RenderTemplatedTTP(tc.content, &tc.execConfig)
			if err != nil {
				t.Fatalf("failed to render and unmarshal templated TTP: %v", err)
				return
			}

			// Capture the start time
			startTime := time.Now()

			_, err = ttp.RunSteps(tc.execConfig)
			if tc.wantError && err == nil {
				t.Error("expected an error from step execution but got none")
				return
			}
			if !tc.wantError && err != nil {
				t.Errorf("didn't expect an error from step execution but got: %s", err)
				return
			}

			// Wait for the cleanup delay
			time.Sleep(tc.cleanupDelay)

			// Check if the directory exists after the cleanup delay
			if directoryExists(dirName) {
				t.Errorf("didn't expect the directory '%s' to exist after the cleanup delay but it does", dirName)
				return
			}

			// Calculate the total duration
			duration := time.Since(startTime)

			// Ensure that the total time taken is reasonably close to the cleanup delay
			if duration < tc.cleanupDelay {
				t.Errorf("total time taken %s is less than expected cleanup delay %s", duration, tc.cleanupDelay)
			}
		})
	}
}
