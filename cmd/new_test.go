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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/facebookincubator/ttpforge/cmd"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

var absConfigPath string

func init() {
	// Find the absolute path to the config.yaml file
	_, currentFilePath, _, _ := runtime.Caller(0)
	absConfigPath = filepath.Join(path.Dir(currentFilePath), "..", "config.yaml")

	// Change into the same directory as the config (repo root).
	if err := os.Chdir(filepath.Dir(absConfigPath)); err != nil {
		panic(err)
	}
}

func TestCreateAndRunTTP(t *testing.T) {
	newTTPBuilderCmd := cmd.NewTTPBuilderCmd()

	basicTestPath := filepath.Join("ttps", "test", "testBasicTTP.yaml")
	fileTestPath := filepath.Join("ttps", "test", "testFileTTP.yaml")

	testCases := []struct {
		name             string
		setFlags         func()
		input            cmd.TTPInput
		expected         string
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "All required flags set",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", absConfigPath)
				_ = newTTPBuilderCmd.Flags().Set("path", basicTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "bash")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "file")
				_ = newTTPBuilderCmd.Flags().Set("args", "arg1,arg2,arg3")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "true")
				_ = newTTPBuilderCmd.Flags().Set("env", "EXAMPLE_ENV_VAR=example_value")
			},
			expectError: false,
		},
		{
			name: "Create basic bash TTP",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", absConfigPath)
				_ = newTTPBuilderCmd.Flags().Set("path", basicTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "bash")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "basic")
				_ = newTTPBuilderCmd.Flags().Set("args", "arg1,arg2,arg3")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "false")
				_ = newTTPBuilderCmd.Flags().Set("env", "EXAMPLE_ENV_VAR=example_value")
			},
			expected: basicTestPath,
		},
		{
			name: "Create file-based bash TTP",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", absConfigPath)
				_ = newTTPBuilderCmd.Flags().Set("path", basicTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "bash")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "file")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "true")
			},
			expected: fileTestPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			newTTPBuilderCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				_ = newTTPBuilderCmd.Flags().Set(flag.Name, "")
			})

			// Set flags for the test case
			tc.setFlags()

			// Call ExecuteContext with the custom context
			err := newTTPBuilderCmd.Execute()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMsg)
			} else {
				require.NoError(t, err)
				// Ensure we are able to read ttps from the TTP directory
				_, err = os.Stat(basicTestPath)
				assert.NoError(t, err, "the test directory should exist")
			}

			// Check if the bash script file was created (for file TTP type)
			if tc.input.TTPType == "file" {
				bashTTPPath := filepath.Join(filepath.Dir(tc.expected), "bashTTP.sh")
				_, err = os.Stat(bashTTPPath)
				assert.False(t, os.IsNotExist(err), "bashTTP.sh file not found: %s", bashTTPPath)
			}

			// Check if the README was created
			readmePath := filepath.Join(filepath.Dir(tc.expected), "README.md")
			_, err = os.Stat(readmePath)
			assert.False(t, os.IsNotExist(err), "README.md file not found: %s", readmePath)

			// Run the created TTP
			runCmd := cmd.RunTTPCmd()
			runCmd.SetArgs([]string{tc.expected}) // Change from basicTestPath to tc.expected
			runOutput := new(bytes.Buffer)
			runCmd.SetOut(runOutput)

			err = runCmd.Execute()
			require.NoError(t, err, fmt.Sprintf("failed to run TTP: %v", err))

			// Cleanup
			os.RemoveAll(filepath.Dir(tc.expected))
		})
	}
}
