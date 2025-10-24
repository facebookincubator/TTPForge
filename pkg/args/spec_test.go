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

package args

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validateTestCase struct {
	name           string
	specs          []Spec
	argKvStrs      []string
	expectedResult map[string]any
	wantError      bool
}

func checkValidateTestCase(t *testing.T, tc validateTestCase) {
	// For tests, use empty strings for base directories (use current directory)
	args, err := ParseAndValidate(tc.specs, tc.argKvStrs, "", "")
	if tc.wantError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, tc.expectedResult, args)
}

func TestValidateArgs(t *testing.T) {

	testCases := []validateTestCase{
		{
			name: "Parse String and Integer Arguments",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
					Type: "int",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=3",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  3,
			},
			wantError: false,
		},
		{
			name: "Parse String and Integer Argument (Default Value)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name:    "beta",
					Type:    "int",
					Default: "1337",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  1337,
			},
			wantError: false,
		},
		{
			name: "Handle Extra Equals",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=bar=baz",
			},
			expectedResult: map[string]any{
				"alpha": "foo",
				"beta":  "bar=baz",
			},
			wantError: false,
		},
		{
			name: "Invalid Inputs (no '=')",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"wut",
			},
			wantError: true,
		},
		{
			name: "Invalid Inputs (Missing Required Argument)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			wantError: true,
		},
		{
			name: "Argument Name Not In Specs",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"gamma=bar",
			},
			wantError: true,
		},
		{
			name: "Duplicate Name in Specs",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "alpha",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
			},
			wantError: true,
		},
		{
			name: "Wrong Type (string instead of int)",
			specs: []Spec{
				{
					Name: "alpha",
				},
				{
					Name: "beta",
					Type: "int",
				},
			},
			argKvStrs: []string{
				"alpha=foo",
				"beta=bar",
			},
			wantError: true,
		},
		{
			name: "Default Value Wrong Type (string instead of int)",
			specs: []Spec{
				{
					Name:    "alpha",
					Type:    "int",
					Default: "wut",
				},
			},
			argKvStrs: []string{
				"alpha=1337",
			},
			wantError: true,
		},
		{
			name: "Format with valid value",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "[A-Z_]+",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT",
			},
			expectedResult: map[string]any{
				"alpha": "CECI_NEST_PAS_UNE_INT",
			},
			wantError: false,
		},
		{
			name: "Format (Flexible Match; No Error)",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "ab",
				},
			},
			argKvStrs: []string{
				"alpha=xabyabz",
			},
			expectedResult: map[string]any{
				"alpha": "xabyabz",
			},
		},
		{
			name: "Format (Strict Match; No Error)",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^ab$",
				},
			},
			argKvStrs: []string{
				"alpha=ab",
			},
			expectedResult: map[string]any{
				"alpha": "ab",
			},
		},
		{
			name: "Format (Strict Match; Error)",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^ab$",
				},
			},
			argKvStrs: []string{
				"alpha=xaby",
			},
			wantError: true,
		},
		{
			name: "Format with proper end and beginning tags",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^[A-Z_-]+$",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT-",
			},
			expectedResult: map[string]any{
				"alpha": "CECI_NEST_PAS_UNE_INT-",
			},
			wantError: false,
		},
		{
			name: "Format with improper regex",
			specs: []Spec{
				{
					Name:   "alpha",
					Type:   "string",
					Format: "^[A-Z_-]+[$$",
				},
			},
			argKvStrs: []string{
				"alpha=CECI_NEST_PAS_UNE_INT-",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkValidateTestCase(t, tc)
		})
	}

}

// TestPathArgHandling tests various edge cases for path-type arguments
func TestPathArgHandling(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	cliDir := filepath.Join(tmpDir, "cli")
	yamlDir := filepath.Join(tmpDir, "yaml")

	err := os.MkdirAll(cliDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(yamlDir, 0755)
	require.NoError(t, err)

	// Create test files
	cliFile := filepath.Join(cliDir, "cli-file.txt")
	yamlFile := filepath.Join(yamlDir, "yaml-file.txt")
	err = os.WriteFile(cliFile, []byte("cli"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(yamlFile, []byte("yaml"), 0644)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		specs          []Spec
		argKvStrs      []string
		cliBaseDir     string
		defaultBaseDir string
		expectedResult map[string]any
		wantError      bool
	}{
		{
			name: "Relative path in default resolves to YAML directory",
			specs: []Spec{
				{
					Name:    "file",
					Type:    "path",
					Default: "yaml-file.txt",
				},
			},
			argKvStrs:      []string{},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": yamlFile,
			},
			wantError: false,
		},
		{
			name: "Relative path in CLI arg resolves to CLI directory",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=cli-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": cliFile,
			},
			wantError: false,
		},
		{
			name: "Absolute path in default is preserved",
			specs: []Spec{
				{
					Name:    "file",
					Type:    "path",
					Default: cliFile,
				},
			},
			argKvStrs:      []string{},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": cliFile,
			},
			wantError: false,
		},
		{
			name: "Absolute path in CLI arg is preserved",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=" + yamlFile,
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": yamlFile,
			},
			wantError: false,
		},
		{
			name: "Path with ./ prefix in default",
			specs: []Spec{
				{
					Name:    "file",
					Type:    "path",
					Default: "./yaml-file.txt",
				},
			},
			argKvStrs:      []string{},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": yamlFile,
			},
			wantError: false,
		},
		{
			name: "Path with ./ prefix in CLI arg",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=./cli-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": cliFile,
			},
			wantError: false,
		},
		{
			name: "Path with .. (parent directory) in default",
			specs: []Spec{
				{
					Name:    "file",
					Type:    "path",
					Default: "../cli/cli-file.txt",
				},
			},
			argKvStrs:      []string{},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": cliFile,
			},
			wantError: false,
		},
		{
			name: "Path with .. (parent directory) in CLI arg",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=../yaml/yaml-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": yamlFile,
			},
			wantError: false,
		},
		{
			name: "CLI arg overrides default with different base dirs",
			specs: []Spec{
				{
					Name:    "file",
					Type:    "path",
					Default: "yaml-file.txt",
				},
			},
			argKvStrs: []string{
				"file=cli-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": cliFile,
			},
			wantError: false,
		},
		{
			name: "Multiple path arguments with mixed bases",
			specs: []Spec{
				{
					Name:    "default_file",
					Type:    "path",
					Default: "yaml-file.txt",
				},
				{
					Name: "cli_file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"cli_file=cli-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"default_file": yamlFile,
				"cli_file":     cliFile,
			},
			wantError: false,
		},
		{
			name: "Shell variable in path (should be expanded by AbsPath)",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=$HOME/test.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": filepath.Join(os.Getenv("HOME"), "test.txt"),
			},
			wantError: false,
		},
		{
			name: "Non-existent path (should still resolve)",
			specs: []Spec{
				{
					Name: "file",
					Type: "path",
				},
			},
			argKvStrs: []string{
				"file=non-existent-file.txt",
			},
			cliBaseDir:     cliDir,
			defaultBaseDir: yamlDir,
			expectedResult: map[string]any{
				"file": filepath.Join(cliDir, "non-existent-file.txt"),
			},
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseAndValidate(tc.specs, tc.argKvStrs, tc.cliBaseDir, tc.defaultBaseDir)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
