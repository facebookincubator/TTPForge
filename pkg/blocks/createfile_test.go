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

package blocks_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFileExecute(t *testing.T) {
	testCases := []struct {
		name               string
		description        string
		step               *blocks.CreateFileStep
		fsysContents       map[string][]byte
		expectExecuteError bool
	}{
		{
			name:        "Create Valid File",
			description: "Create a single unremarkable file",
			step: &blocks.CreateFileStep{
				Path:     "valid-file.txt",
				Contents: "hello world",
			},
		},
		{
			name:        "Nested Directories",
			description: "Afero should handle this under the hood",
			step: &blocks.CreateFileStep{
				Path:     "/directory/does/not/exist",
				Contents: "should still work",
			},
		},
		{
			name:        "Already Exists (No Overwrite)",
			description: "Should fail because file already exists",
			step: &blocks.CreateFileStep{
				Path:     "already-exists.txt",
				Contents: "will fail",
			},
			fsysContents: map[string][]byte{
				"already-exists.txt": []byte("whoops"),
			},
			expectExecuteError: true,
		},
		{
			name:        "Already Exists (With Overwrite)",
			description: "Should succeed and overwrite existing file",
			step: &blocks.CreateFileStep{
				Path:      "already-exists.txt",
				Contents:  "will succeed",
				Overwrite: true,
			},
			fsysContents: map[string][]byte{
				"already-exists.txt": []byte("whoops"),
			},
		},
		{
			name:        "Set Permissions Manually",
			description: "Make the file read-only",
			step: &blocks.CreateFileStep{
				Path:     "make-read-only",
				Contents: "very-read-only",
				Perm:     0600,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prep filesystem
			if tc.fsysContents != nil {
				fsys, err := testutils.MakeAferoTestFs(tc.fsysContents)
				require.NoError(t, err)
				tc.step.FileSystem = fsys
			} else {
				tc.step.FileSystem = afero.NewMemMapFs()
			}

			// execute and check error
			var execCtx blocks.TTPExecutionContext
			_, err := tc.step.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check contents
			contentBytes, err := afero.ReadFile(tc.step.FileSystem, tc.step.Path)
			require.NoError(t, err)
			assert.Equal(t, tc.step.Contents, string(contentBytes))

			// check permissions
			if tc.step.Perm != 0 {
				info, err := tc.step.FileSystem.Stat(tc.step.Path)
				require.NoError(t, err)
				assert.Equal(t, tc.step.Perm, info.Mode())
			}
		})
	}
}
