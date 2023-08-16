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
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/l50/goutils/v2/str"
	"github.com/stretchr/testify/assert"
)

func TestFetchAbs(t *testing.T) {
	testCases := []struct {
		name         string
		inputPath    string
		inputWorkdir string
		expectError  bool
	}{
		{
			name:         "Absolute path",
			inputPath:    "/tmp",
			inputWorkdir: "",
			expectError:  false,
		},
		{
			name:         "Home directory",
			inputPath:    "~/",
			inputWorkdir: "",
			expectError:  false,
		},
		{
			name:         "Relative path",
			inputPath:    "test_directory",
			inputWorkdir: ".",
			expectError:  false,
		},
		{
			name:         "Invalid path",
			inputPath:    "",
			inputWorkdir: "",
			expectError:  true,
		},
		{
			name:         "Path with dot prefix",
			inputPath:    "./test_directory",
			inputWorkdir: "/tmp",
			expectError:  false,
		},
		{
			name:         "Common prefix path",
			inputPath:    "./ttps/privilege-escalation/credential-theft/hello-world/hello-world.sh",
			inputWorkdir: "/Users/test/ttpforge/ttps/privilege-escalation/credential-theft/hello-world",
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := blocks.FetchAbs(tc.inputPath, tc.inputWorkdir)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				switch {
				case tc.inputPath == "~/":
					homeDir, _ := os.UserHomeDir()
					assert.Equal(t, homeDir, filepath.Clean(result))
				case filepath.IsAbs(tc.inputPath):
					assert.Equal(t, tc.inputPath, result)
				default:
					expected, _ := filepath.Abs(filepath.Join(tc.inputWorkdir, tc.inputPath))
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}

func TestFindFilePath(t *testing.T) {
	workdir, err := os.Getwd()
	assert.NoError(t, err)

	tempDir := filepath.Join(workdir, "temp_test_directory")
	err = os.Mkdir(tempDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test_file.txt")
	f, err := os.Create(tempFile)
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(tempFile)

	// Create a tilde file for testing
	tildeFile := filepath.Join(os.Getenv("HOME"), "tilde_test_file.txt")
	f, err = os.Create(tildeFile)
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(tildeFile)

	testCases := []struct {
		name         string
		inputPath    string
		inputWorkdir string
		fsStat       fs.StatFS
		expectError  bool
	}{
		{
			name:         "Absolute path",
			inputPath:    tempFile,
			inputWorkdir: "",
			fsStat:       nil,
			expectError:  false,
		},
		{
			name:         "Relative path",
			inputPath:    "temp_test_directory/test_file.txt",
			inputWorkdir: workdir,
			fsStat:       nil,
			expectError:  false,
		},
		{
			name:         "Non-existent path",
			inputPath:    "non_existent_file.txt",
			inputWorkdir: "",
			fsStat:       nil,
			expectError:  true,
		},
		{
			name:         "Tilde path",
			inputPath:    "~/tilde_test_file.txt",
			inputWorkdir: "",
			fsStat:       nil,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := blocks.FindFilePath(tc.inputPath, tc.inputWorkdir, tc.fsStat)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				switch {
				case filepath.IsAbs(tc.inputPath):
					assert.Equal(t, tc.inputPath, result)
				case strings.HasPrefix(tc.inputPath, "~"):
					expandedPath := strings.Replace(tc.inputPath, "~", os.Getenv("HOME"), 1)
					assert.Equal(t, expandedPath, result)
				default:
					expected, _ := filepath.Abs(filepath.Join(tc.inputWorkdir, tc.inputPath))
					assert.Equal(t, expected, result)
				}
			}
		})
	}
}

func TestFetchEnv(t *testing.T) {
	tests := []struct {
		name     string
		environ  map[string]string
		expected []string
	}{
		{
			name:     "Empty environment map",
			environ:  map[string]string{},
			expected: []string{},
		},
		{
			name: "Single environment variable",
			environ: map[string]string{
				"TEST_ENV_VAR": "test_value",
			},
			expected: []string{"TEST_ENV_VAR=test_value"},
		},
		{
			name: "Multiple environment variables",
			environ: map[string]string{
				"TEST_ENV_VAR_1": "test_value_1",
				"TEST_ENV_VAR_2": "test_value_2",
			},
			expected: []string{"TEST_ENV_VAR_1=test_value_1", "TEST_ENV_VAR_2=test_value_2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := blocks.FetchEnv(tt.environ)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !str.SlicesEqual(tt.expected, result) {
				t.Errorf("mismatch in environment variable slice. expected length: %d, got length: %d, expected: %v, got: %v", len(tt.expected), len(result), tt.expected, result)
			} else {
				t.Logf("passed: expected length: %d, got length: %d, expected: %v, got: %v", len(tt.expected), len(result), tt.expected, result)
			}
		})
	}
}
