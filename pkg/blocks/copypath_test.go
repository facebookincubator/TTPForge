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
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyPathExecute(t *testing.T) {
	const sourceDirectory = "TTPForgeSource"
	const destinationDirectory = "TTPForgeDestination"

	filesMap := map[string][]byte{
		filepath.Join(sourceDirectory, "ttpforge_test.txt"):           []byte("This is a TTPForge test file."),
		filepath.Join(sourceDirectory, "subdir1", "subdir1_test.txt"): []byte("This is a TTPForge test file OK."),
		filepath.Join(sourceDirectory, "subdir2", "subdir2_test.txt"): []byte("This is a TTPForge test file OKAAAAAAY."),
		filepath.Join(destinationDirectory, "ttpforge_test.txt"):      []byte("This is a TTPForge test file but I'm already here!."),
	}

	testCases := []struct {
		name                string
		description         string
		relativeSource      string
		relativeDestination string
		overwrite           bool
		recursive           bool
		stepVars            map[string]string
		expectTemplateError bool
		expectExecuteError  bool
	}{
		{
			name:                "Attempt to copy non-existent file",
			description:         "Expected to fail due to non-existent source file",
			relativeSource:      "thisfiledoesnotexistok",
			relativeDestination: "thisdoesntmatter",
			stepVars:            map[string]string{},
			expectTemplateError: false,
			expectExecuteError:  true,
		},
		{
			name:                "Copy existing file to new path",
			description:         "This should succeed as source file exists and destination does not",
			relativeSource:      filepath.Join(sourceDirectory, "ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "ttpforge_test_copy.txt"),
			stepVars:            map[string]string{},
			expectTemplateError: false,
		},
		{
			name:                "Copy to preexisting desitnation (no overwrite)",
			description:         "This should fail since the destination file exists and we are not specifying to overwrite it",
			relativeSource:      filepath.Join(sourceDirectory, "ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "ttpforge_test.txt"),
			stepVars:            map[string]string{},
			expectTemplateError: false,
			expectExecuteError:  true,
		},
		{
			name:                "Copy to preexisting destination (overwrite true)",
			description:         "This should pass since when destination file exists since we are specifying overwrite true",
			relativeSource:      filepath.Join(sourceDirectory, "ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "ttpforge_test.txt"),
			stepVars:            map[string]string{},
			expectTemplateError: false,
			overwrite:           true,
		},
		{
			name:                "Copy a directory to a destination that already exists (no overwrite)",
			description:         "This should fail since the destination directory exists and we are not specifying overwrite true",
			relativeSource:      sourceDirectory,
			relativeDestination: destinationDirectory,
			recursive:           true,
			stepVars:            map[string]string{},
			expectTemplateError: false,
			expectExecuteError:  true,
		},
		{
			name:                "Copy a directory to a destination that already exists (with overwrite)",
			description:         "This should pass when the destination directory exists since we specifying overwrite true",
			relativeSource:      sourceDirectory,
			relativeDestination: destinationDirectory,
			recursive:           true,
			stepVars:            map[string]string{},
			expectTemplateError: false,
			overwrite:           true,
		},
		{
			name:                "Successfully templates source and destination paths",
			description:         "This should pass when the source and destination paths are templated",
			relativeSource:      filepath.Join(sourceDirectory, "{[{.StepVars.foo}]}_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "{[{.StepVars.foo}]}_test_copy.txt"),
			stepVars: map[string]string{
				"foo": "ttpforge",
			},
			expectTemplateError: false,
		},
		{
			name:                "Errors out when variable is missing",
			description:         "This should fail when the foo variable is missing",
			relativeSource:      filepath.Join(sourceDirectory, "{[{.StepVars.foo}]}_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "{[{.StepVars.foo}]}_test_copy.txt"),
			stepVars:            map[string]string{},
			expectTemplateError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// prep filesystem
			tempDir, err := testutils.MakeTempTestDir(filesMap)
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// prep execution context
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

			// create copy step
			copyTestPathStep := CopyPathStep{
				Source:      filepath.Join(tempDir, tc.relativeSource),
				Destination: filepath.Join(tempDir, tc.relativeDestination),
				Overwrite:   tc.overwrite,
				Recursive:   tc.recursive,
			}

			// template and check error
			err = copyTestPathStep.Template(execCtx)
			if tc.expectTemplateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// execute and check error
			_, err = copyTestPathStep.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if !copyTestPathStep.Recursive {
				srcContentBytes, err := os.ReadFile(copyTestPathStep.Source)
				require.NoError(t, err)
				destContentBytes, err := os.ReadFile(copyTestPathStep.Destination)
				require.NoError(t, err)
				assert.Equal(t, destContentBytes, srcContentBytes)
			} else {
				dirsEqual, err := testutils.AreDirsEqual(copyTestPathStep.Source, copyTestPathStep.Destination)
				require.NoError(t, err)
				assert.True(t, dirsEqual)
			}

			// check permissions
			if copyTestPathStep.Mode != 0 {
				info, err := os.Stat(copyTestPathStep.Destination)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(copyTestPathStep.Mode), info.Mode())
			}

		})
	}
}
