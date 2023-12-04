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

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sourceDirectory = "/TTPForgeSource"
const destinationDirectory = "/TTPForgeDestination"

func TestCopyPathExecute(t *testing.T) {
	testCases := []struct {
		name                string
		description         string
		relativeSource      string
		relativeDestination string
		overwrite           bool
		recursive           bool
		expectExecuteError  bool
	}{
		{
			name:                "Attempt to copy non-existent file",
			description:         "Expected to fail due to non-existent source file",
			relativeSource:      "/thisfiledoesnotexistok",
			relativeDestination: "/thisdoesntmatter",
			expectExecuteError:  true,
		},
		{
			name:                "Copy existing file to new path",
			description:         "This should succeed as source file exists and destination does not",
			relativeSource:      filepath.Join(sourceDirectory, "/ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "/ttpforge_test_copy.txt"),
		},
		{
			name:                "Copy to preexisting desitnation (no overwrite)",
			description:         "This should fail since the destination file exists and we are not specifying to overwrite it",
			relativeSource:      filepath.Join(sourceDirectory, "/ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "/ttpforge_test.txt"),
			expectExecuteError:  true,
		},
		{
			name:                "Copy to preexisting desitnation (overwrite true)",
			description:         "This should pass since when destination file exists since we are specifying overwrite true",
			relativeSource:      filepath.Join(sourceDirectory, "/ttpforge_test.txt"),
			relativeDestination: filepath.Join(destinationDirectory, "/ttpforge_test.txt"),
			overwrite:           true,
		},
		{
			name:                "Copy a directory to a destination that already exists (no overwrite)",
			description:         "This should fail since the destination directory exists and we are not specifying overwrite true",
			relativeSource:      sourceDirectory,
			relativeDestination: destinationDirectory,
			recursive:           true,
			expectExecuteError:  true,
		},
		{
			name:                "Copy a directory to a destination that already exists (with overwrite)",
			description:         "This should fail since the destination directory exists and we are not specifying overwrite true",
			relativeSource:      sourceDirectory,
			relativeDestination: destinationDirectory,
			recursive:           true,
			overwrite:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// prep filesystem
			tempDir, err := os.MkdirTemp("", "ttpforge")
			if err != nil {
				return
			}
			defer os.RemoveAll(tempDir)
			PrepTestFs(tempDir)
			// create copy step
			copyTestPathStep := CopyPathStep{
				Source:      filepath.Join(tempDir, tc.relativeSource),
				Destination: filepath.Join(tempDir, tc.relativeDestination),
				Overwrite:   tc.overwrite,
				Recursive:   tc.recursive,
			}

			// execute and check error
			var execCtx TTPExecutionContext
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
				dirsEqual, err := AreDirsEqual(copyTestPathStep.Source, copyTestPathStep.Destination)
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

// PrepTestFs prepares the file system for testing, modify as needed if required when adding additional tests.
func PrepTestFs(tempDir string) error {

	filesMap := map[string][]byte{
		filepath.Join(tempDir, "/TTPForgeSource/ttpforge_test.txt"):        []byte("This is a TTPForge test file."),
		filepath.Join(tempDir, "/TTPForgeSource/subdir1/subdir1_test.txt"): []byte("This is a TTPForge test file OK."),
		filepath.Join(tempDir, "/TTPForgeSource/subdir2/subdir2_test.txt"): []byte("This is a TTPForge test file OKAAAAAAY."),
		filepath.Join(tempDir, "/TTPForgeDestination/ttpforge_test.txt"):   []byte("This is a TTPForge test file but I'm already here!."),
	}

	for path, contents := range filesMap {
		dirPath := filepath.Dir(path)
		err := os.MkdirAll(dirPath, 0700)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, contents, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// AreDirsEqual recursively compares two directories for equality
// NOTE: filepath.Wale guarantees lexical order traversal (elements already ordered) thus allowing us to use a slice
// as opposed to a map.
func AreDirsEqual(source string, dest string) (bool, error) {
	var files1, files2 []string
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files1 = append(files1, string(content))
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files2 = append(files2, string(content))
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return cmp.Equal(files1, files2), nil
}
