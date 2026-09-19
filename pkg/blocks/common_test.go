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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			result, err := FetchAbs(tc.inputPath, tc.inputWorkdir)

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
	tempDir, err := os.MkdirTemp("", "temp_test_directory")
	tempDirName := filepath.Base(tempDir)
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	workDir := filepath.Dir(tempDir)

	testFileName := "test_file.txt"
	tempFile := filepath.Join(tempDir, testFileName)
	f, err := os.Create(tempFile)
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(tempFile)

	// Create a tilde file for testing
	uniqueFileName := fmt.Sprintf("tilde_test_file%d.txt", os.Getpid())
	tildeFile := filepath.Join(os.Getenv("HOME"), uniqueFileName)
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
			inputPath:    filepath.Join(tempDirName, testFileName),
			inputWorkdir: workDir,
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
			inputPath:    filepath.Join("~", uniqueFileName),
			inputWorkdir: "",
			fsStat:       nil,
			expectError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FindFilePath(tc.inputPath, tc.inputWorkdir, tc.fsStat)

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
			expected: nil,
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
			result := FetchEnv(tt.environ)
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestForwardedEnv(t *testing.T) {
	t.Setenv(ForwardEnvVar, "")
	t.Setenv("TTPFORGE_TEST_ABSENT", "")
	assert.NoError(t, os.Unsetenv("TTPFORGE_TEST_ABSENT"))

	tests := []struct {
		name     string
		set      map[string]string
		names    []string
		expected []string
	}{
		{
			name:     "No names requested",
			names:    nil,
			expected: nil,
		},
		{
			name:     "Requested name is unset locally",
			names:    []string{"TTPFORGE_TEST_ABSENT"},
			expected: nil,
		},
		{
			name:     "Single forwarded variable",
			set:      map[string]string{"TTPFORGE_TEST_SESSION": "session-abc123"},
			names:    []string{"TTPFORGE_TEST_SESSION"},
			expected: []string{"TTPFORGE_TEST_SESSION=session-abc123"},
		},
		{
			name:     "Empty value is still forwarded",
			set:      map[string]string{"TTPFORGE_TEST_EMPTY": ""},
			names:    []string{"TTPFORGE_TEST_EMPTY"},
			expected: []string{"TTPFORGE_TEST_EMPTY="},
		},
		{
			name: "Portable variable names are accepted",
			set: map[string]string{
				"_":                   "underscore",
				"ttpforge_test_lower": "lowercase",
				"TTPFORGE_TEST_2":     "digit-after-first",
			},
			names: []string{"_", "ttpforge_test_lower", "TTPFORGE_TEST_2"},
			expected: []string{
				"_=underscore",
				"ttpforge_test_lower=lowercase",
				"TTPFORGE_TEST_2=digit-after-first",
			},
		},
		{
			name:     "Empty names are ignored in both sources",
			set:      map[string]string{ForwardEnvVar: ", ,\t"},
			names:    []string{"", " \t "},
			expected: nil,
		},
		{
			name: "Set and unset names mixed, order preserved",
			set: map[string]string{
				"TTPFORGE_TEST_ONE": "1",
				"TTPFORGE_TEST_TWO": "2",
			},
			names:    []string{"TTPFORGE_TEST_ONE", "TTPFORGE_TEST_ABSENT", "TTPFORGE_TEST_TWO"},
			expected: []string{"TTPFORGE_TEST_ONE=1", "TTPFORGE_TEST_TWO=2"},
		},
		{
			name: "Names come from TTPFORGE_FORWARD_ENV alone",
			set: map[string]string{
				"TTPFORGE_TEST_ONE": "1",
				ForwardEnvVar:       "TTPFORGE_TEST_ONE",
			},
			expected: []string{"TTPFORGE_TEST_ONE=1"},
		},
		{
			name: "TTPFORGE_FORWARD_ENV list is split and trimmed",
			set: map[string]string{
				"TTPFORGE_TEST_ONE": "1",
				"TTPFORGE_TEST_TWO": "2",
				ForwardEnvVar:       " TTPFORGE_TEST_ONE , TTPFORGE_TEST_TWO ",
			},
			expected: []string{"TTPFORGE_TEST_ONE=1", "TTPFORGE_TEST_TWO=2"},
		},
		{
			name: "A name given both ways is emitted once",
			set: map[string]string{
				"TTPFORGE_TEST_ONE": "1",
				ForwardEnvVar:       "TTPFORGE_TEST_ONE",
			},
			names:    []string{"TTPFORGE_TEST_ONE"},
			expected: []string{"TTPFORGE_TEST_ONE=1"},
		},
		{
			name:     "Empty TTPFORGE_FORWARD_ENV forwards nothing",
			set:      map[string]string{ForwardEnvVar: ""},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.set {
				t.Setenv(k, v)
			}

			result, err := ForwardedEnv(tt.names)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestForwardedEnvRejectsInvalidNames(t *testing.T) {
	t.Setenv(ForwardEnvVar, "")
	t.Setenv("TTPFORGE_TEST_SAFE", "safe")

	for _, tc := range []struct {
		name    string
		invalid string
	}{
		{name: "POSIX command separator", invalid: "TTPFORGE_TEST_BAD;printf injected"},
		{name: "PowerShell subexpression", invalid: "TTPFORGE_TEST_BAD$(Write-Output injected)"},
		{name: "cmd command separator", invalid: "TTPFORGE_TEST_BAD&echo injected&"},
		{name: "Digit first", invalid: "1TTPFORGE_TEST_BAD"},
		{name: "Interior whitespace", invalid: "TTPFORGE TEST_BAD"},
		{name: "Non-ASCII letter", invalid: "TTPFORGE_TEST_\u00e9"},
	} {
		for _, source := range []string{"explicit", "environment"} {
			for _, set := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/%s/set=%t", tc.name, source, set), func(t *testing.T) {
					t.Setenv(tc.invalid, "must-not-forward")
					if !set {
						require.NoError(t, os.Unsetenv(tc.invalid))
					}
					names := []string{"TTPFORGE_TEST_SAFE"}
					if source == "environment" {
						t.Setenv(ForwardEnvVar, tc.invalid)
					} else {
						names = append(names, tc.invalid)
					}

					result, err := ForwardedEnv(names)
					require.ErrorContains(t, err, "invalid forwarded environment variable name")
					assert.Nil(t, result)
				})
			}
		}
	}
}
