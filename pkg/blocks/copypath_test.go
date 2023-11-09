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

package blocks

import (
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyPathExecute(t *testing.T) {
	testCases := []struct {
		name               string
		description        string
		step               *CopyPathStep
		fsysContents       map[string][]byte
		expectExecuteError bool
	}{
		{
			name:        "Attempt to copy non-existent file",
			description: "Expected to fail due to non-existent source file",
			step: &CopyPathStep{
				Source:      "/etc/passwd",
				Destination: "/tmp/passwd",
			},
			expectExecuteError: true,
		},
		{
			name:        "Copy existing file to new path",
			description: "This should succeed as source file exists and destination does not",
			step: &CopyPathStep{
				Source:      "/etc/passwd",
				Destination: "/tmp/passwd",
			},
			fsysContents: map[string][]byte{
				"/etc/passwd": []byte("whoops"),
			},
		},
		{
			name:        "Copy to preexisting desitnation (no overwrite)",
			description: "This should fail since the destination file exists and we are not specifying to overwrite it",
			step: &CopyPathStep{
				Source:      "/etc/passwd",
				Destination: "/tmp/passwd",
			},
			fsysContents: map[string][]byte{
				"/etc/passwd": []byte("whoops"),
				"/tmp/passwd": []byte("fail"),
			},
			expectExecuteError: true,
		},
		{
			name:        "Copy to preexisting desitnation (overwrite true)",
			description: "This should pass since when destination file exists since we are specifying overwrite true",
			step: &CopyPathStep{
				Source:      "/etc/passwd",
				Destination: "/tmp/passwd",
				Overwrite:   true,
			},
			fsysContents: map[string][]byte{
				"/etc/passwd": []byte("whoops"),
				"/tmp/passwd": []byte("pass"),
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
			var execCtx TTPExecutionContext
			_, err := tc.step.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// get contents of source file
			srcContentBytes, err := afero.ReadFile(tc.step.FileSystem, tc.step.Source)
			require.NoError(t, err)

			destContentBytes, err := afero.ReadFile(tc.step.FileSystem, tc.step.Destination)
			require.NoError(t, err)

			assert.Equal(t, destContentBytes, srcContentBytes)

			// check permissions
			if tc.step.Mode != 0 {
				info, err := tc.step.FileSystem.Stat(tc.step.Destination)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(tc.step.Mode), info.Mode())
			}
		})
	}
}
