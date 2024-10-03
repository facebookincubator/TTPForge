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
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeDirectoryExecute(t *testing.T) {

	testCases := []struct {
		name          string
		description   string
		step          *ChangeDirectoryStep
		fsysContents  map[string][]byte
		expectedError bool
		startingDir   string
	}{
		{
			name:        "Change directory to valid directory",
			description: "Change directory and expect successful change of workdir",
			step: &ChangeDirectoryStep{
				Cd: "/tmp",
			},
			fsysContents: map[string][]byte{
				"/home/testuser/test": []byte("test"),
				"/tmp/test":           []byte("test"),
			},
			expectedError: false,
			startingDir:   "/home/testuser/",
		},
		{
			name:        "Change directory to invalid directory",
			description: "Try to change directory to invalid directory and expect error",
			step: &ChangeDirectoryStep{
				Cd: "/doesntexist",
			},
			fsysContents: map[string][]byte{
				"/home/testuser/test": []byte("test"),
				"/tmp/test":           []byte("test"),
			},
			expectedError: true,
			startingDir:   "/home/testuser/",
		},
		{
			name:        "Change directory with no given directory",
			description: "Try to change directory to no directory and expect error",
			step: &ChangeDirectoryStep{
				Cd: "",
			},
			fsysContents: map[string][]byte{
				"/home/testuser/test": []byte("test"),
				"/tmp/test":           []byte("test"),
			},
			expectedError: true,
			startingDir:   "/home/testuser/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prep filesystem
			fsys, err := testutils.MakeAferoTestFs(tc.fsysContents)
			require.NoError(t, err)
			tc.step.FileSystem = fsys

			// validate and check error
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.WorkDir = tc.startingDir
			err = tc.step.Validate(execCtx)

			if tc.expectedError && err != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// execute and check error
			_, err = tc.step.Execute(execCtx)

			if tc.expectedError && err != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check current working directory
			assert.Equal(t, tc.step.Cd, execCtx.Vars.WorkDir)

			// cleanup and check error
			err = tc.step.GetDefaultCleanupAction().Validate(execCtx)
			require.NoError(t, err)
			_, err = tc.step.GetDefaultCleanupAction().Execute(execCtx)
			require.NoError(t, err)

			// expect working directory to be rolled back to starting directory
			assert.Equal(t, tc.startingDir, execCtx.Vars.WorkDir)
		})
	}
}
