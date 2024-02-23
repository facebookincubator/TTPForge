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

package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeTempTestDir(t *testing.T) {
	testCases := []struct {
		name        string
		description string
		filesMap    map[string][]byte
		expectError bool
	}{
		{
			name:        "Create Directory Tree",
			description: "Create multiple files and directories without errors",
			filesMap: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
		},
		{
			name:        "Pass in Absolute Paths",
			description: "Should fail because this function should reject absolute paths",
			filesMap: map[string][]byte{
				"/a/b/foo.txt": []byte("hello world"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prep directories
			tempDir, err := MakeTempTestDir(tc.filesMap)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// check all files
			for path, content := range tc.filesMap {
				fullPath := filepath.Join(tempDir, path)
				actualContent, err := os.ReadFile(fullPath)
				require.NoError(t, err)
				assert.Equal(t, content, actualContent)
			}
		})
	}
}

func TestAreDirsEqual(t *testing.T) {
	testCases := []struct {
		name           string
		description    string
		filesMapOne    map[string][]byte
		filesMapTwo    map[string][]byte
		expectedResult bool
		expectError    bool
	}{
		{
			name:        "Equal Directories",
			description: "Expected to return true",
			filesMapOne: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
			filesMapTwo: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
			expectedResult: true,
		},
		{
			name:        "Unequal Directories (different file contents)",
			description: "Expected to return false",
			filesMapOne: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
			filesMapTwo: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("this is different"),
			},
			expectedResult: false,
		},
		{
			name:        "Unequal Directories (extra file in dir two)",
			description: "Expected to return false",
			filesMapOne: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
				"c/moar.txt":  []byte("should not be here"),
			},
			filesMapTwo: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
			expectedResult: false,
		},
		{
			name:        "Unequal Directories (extra file in dir two)",
			description: "Expected to return false",
			filesMapOne: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
			},
			filesMapTwo: map[string][]byte{
				"a/b/foo.txt": []byte("hello world"),
				"a/b/bar.txt": []byte("hey there"),
				"c/baz.txt":   []byte("victory"),
				"c/moar.txt":  []byte("should not be here"),
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// prep directories
			dirOne, err := MakeTempTestDir(tc.filesMapOne)
			require.NoError(t, err)
			defer os.RemoveAll(dirOne)
			dirTwo, err := MakeTempTestDir(tc.filesMapTwo)
			require.NoError(t, err)
			defer os.RemoveAll(dirTwo)

			// compare
			result, err := AreDirsEqual(dirOne, dirTwo)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// verify result
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
