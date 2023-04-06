package blocks_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/facebookincubator/TTP-Runner/blocks"
	goutils "github.com/l50/goutils"
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
			inputWorkdir: "/Users/test/TTP-Runner/ttps/privilege-escalation/credential-theft/hello-world",
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

			if !goutils.StringSlicesEqual(tt.expected, result) {
				t.Errorf("mismatch in environment variable slice. expected length: %d, got length: %d, expected: %v, got: %v", len(tt.expected), len(result), tt.expected, result)
			} else {
				t.Logf("passed: expected length: %d, got length: %d, expected: %v, got: %v", len(tt.expected), len(result), tt.expected, result)
			}
		})
	}
}

func TestJSONString(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	tests := []struct {
		name     string
		input    any
		expected string
		wantErr  bool
	}{
		{
			name: "Valid input",
			input: TestStruct{
				Field1: "test",
				Field2: 123,
			},
			expected: "'{\"field1\":\"test\",\"field2\":123}'",
			wantErr:  false,
		},
		{
			name:     "Invalid input",
			input:    make(chan int),
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := blocks.JSONString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("error running JSONString(): %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expected {
				t.Errorf("unexpected output from JSONString() got = %v, expected %v", got, tt.expected)
			}
		})
	}
}
