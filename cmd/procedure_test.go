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
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/cmd"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunProcedureCommand(t *testing.T) {
	testDir, _ := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	embeddedFS, err := fs.Sub(os.DirFS(testDir), ".generated_ttps")
	require.NoError(t, err, "failed to create a sub FS")

	runProcCmd := cmd.NewRunProcCmd(&embeddedFS)

	testCases := []struct {
		name            string
		setFlags        func()
		expectError     bool
		path            string
		expectedSubDirs []string
		expectedFiles   []string
	}{
		{
			name:            "root directory",
			path:            ".",
			expectedSubDirs: []string{"dir1", "dir2"},
			expectedFiles:   []string{"file1.yaml", "file2.yaml"},
		},
		{
			name:            "subdirectory",
			path:            "dir1",
			expectedSubDirs: []string{"subdir1"},
			expectedFiles:   []string{"file3.yaml"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setFlags()

			// Capture stdout
			old := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err = runProcCmd.Execute()

			w.Close()
			out, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			os.Stdout = old

			output := string(out)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				for _, expectedFile := range tc.expectedFiles {
					assert.Contains(t, output, fmt.Sprintf("/%s", expectedFile))
				}
			}

			// Reset flags
			runProcCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				_ = runProcCmd.Flags().Set(flag.Name, "")
			})
		})
	}
}
