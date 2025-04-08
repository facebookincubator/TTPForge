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

func TestRemovePathExecute(t *testing.T) {
	// need this for some test cases
	homedir, err := os.UserHomeDir()
	require.NoError(t, err)

	testCases := []struct {
		name                string
		description         string
		step                *RemovePathAction
		stepVars            map[string]string
		fsysContents        map[string][]byte
		expectValidateError bool
		expectTemplateError bool
		expectExecuteError  bool
	}{
		{
			name:        "Remove Valid File",
			description: "Remove a single unremarkable file",
			step: &RemovePathAction{
				Path: "valid-file.txt",
			},
			fsysContents: map[string][]byte{
				"valid-file.txt": []byte("whoops"),
			},
		},
		{
			name:        "Remove Non-Existent File",
			description: "Remove a non-existent file - should error",
			step: &RemovePathAction{
				Path: "does-not-exist.txt",
			},
			fsysContents: map[string][]byte{
				"valid-file.txt": []byte("whoops"),
			},
			expectExecuteError: true,
		},
		{
			name:        "Remove Directory - Success",
			description: "Set Recursive to make directory removal succeed",
			step: &RemovePathAction{
				Path:      "valid-directory",
				Recursive: true,
			},
			fsysContents: map[string][]byte{
				"valid-directory/valid-file.txt": []byte("whoops"),
			},
		},
		{
			name:        "Remove Directory - Failure",
			description: "Refuse to remove directory because `recursive: true` was not specified",
			step: &RemovePathAction{
				Path: "valid-directory",
			},
			fsysContents: map[string][]byte{
				"valid-directory/valid-file.txt": []byte("whoops"),
			},
			expectExecuteError: true,
		},
		{
			name:        "Expand Tilde Into Home Directory",
			description: "Ensure that ~ is expanded into home directory appropriately",
			step: &RemovePathAction{
				Path: "~/this-should-work",
			},

			fsysContents: map[string][]byte{
				homedir + "/this-should-work": []byte("hopefully this file gets deleted correctly"),
			},
		},
		{
			name:        "Remove Valid File with templating",
			description: "Remove a single unremarkable file with templating",
			step: &RemovePathAction{
				Path: "{[{.StepVars.filename}]}.txt",
			},
			fsysContents: map[string][]byte{
				"valid-file.txt": []byte("whoops"),
			},
			stepVars: map[string]string{
				"filename": "valid-file",
			},
		},
		{
			name:        "Error on missing templating variable",
			description: "Remove a single unremarkable file with templating",
			step: &RemovePathAction{
				Path: "{[{.StepVars.filename}]}.txt",
			},
			fsysContents: map[string][]byte{
				"valid-file.txt": []byte("whoops"),
			},
			expectTemplateError: true,
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

			// prep execution context
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

			// validate
			err := tc.step.Validate(execCtx)
			if tc.expectValidateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// template
			err = tc.step.Template(execCtx)
			if tc.expectTemplateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// execute
			_, err = tc.step.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify that the file is gone
			exists, err := afero.Exists(tc.step.FileSystem, tc.step.Path)
			require.NoError(t, err)
			assert.False(t, exists)
		})
	}
}
