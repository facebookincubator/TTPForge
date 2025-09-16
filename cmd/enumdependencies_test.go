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

package cmd

import (
	"bytes"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

const (
	enumDepsWorkingDir  = "ttpforge-enumdeps-test"
	enumDepsPrimaryRepo = "test-repo"
	enumDepsSearchPath  = "ttps"
)

type enumDepsCmdTestCase struct {
	name             string
	description      string
	sourceArg        string
	wantError        bool
	errorContains    string
	setupFiles       []string
	referencingFiles map[string]string // path -> content that references the source file
	verbose          bool
}

func setupEnumDepsTestFiles(t *testing.T, fsys afero.Fs, baseDir string, files []string) {
	for _, file := range files {
		fullPath := filepath.Join(baseDir, file)
		dir := filepath.Dir(fullPath)
		err := fsys.MkdirAll(dir, 0755)
		require.NoError(t, err)

		content := `---
name: Test TTP
description: A test TTP for enum dependencies testing
steps:
  - name: test_step
    inline: echo "test output"
`
		err = afero.WriteFile(fsys, fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func setupEnumDepsReferencingFiles(t *testing.T, fsys afero.Fs, baseDir string, files map[string]string) {
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		dir := filepath.Dir(fullPath)
		err := fsys.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = afero.WriteFile(fsys, fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func checkEnumDepsTestCase(t *testing.T, tc enumDepsCmdTestCase) {
	// Create a temporary directory for this test
	fsys := afero.NewOsFs()
	tempDir, err := afero.TempDir(fsys, "", enumDepsWorkingDir)
	require.NoError(t, err)
	defer fsys.RemoveAll(tempDir)

	// Create test repository structure
	primaryRepoDir := filepath.Join(tempDir, enumDepsPrimaryRepo)
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create repo config
	repoConfigContent := `---
ttp_search_paths:
  - ` + enumDepsSearchPath + `
`
	err = fsys.MkdirAll(filepath.Join(primaryRepoDir, enumDepsSearchPath), 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fsys, filepath.Join(primaryRepoDir, "ttpforge-repo-config.yaml"), []byte(repoConfigContent), 0644)
	require.NoError(t, err)

	// Create main config
	mainConfigContent := `---
repos:
  - name: ` + enumDepsPrimaryRepo + `
    path: ` + primaryRepoDir + `
`
	err = afero.WriteFile(fsys, configPath, []byte(mainConfigContent), 0644)
	require.NoError(t, err)

	// Setup test files
	if tc.setupFiles != nil {
		setupEnumDepsTestFiles(t, fsys, tempDir, tc.setupFiles)
	}

	// Setup referencing files
	if tc.referencingFiles != nil {
		setupEnumDepsReferencingFiles(t, fsys, tempDir, tc.referencingFiles)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	rc := BuildRootCommand(&TestConfig{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})

	args := []string{"enum", "dependencies", "-c", configPath, tc.sourceArg}
	if tc.verbose {
		args = append(args, "--verbose")
	}
	rc.SetArgs(args)

	err = rc.Execute()

	if tc.wantError {
		require.Error(t, err)
		if tc.errorContains != "" {
			assert.Contains(t, err.Error(), tc.errorContains)
		}
		return
	}

	require.NoError(t, err, "Enum dependencies command should succeed")
}

func TestEnumDependenciesCommand(t *testing.T) {
	testCases := []enumDepsCmdTestCase{
		{
			name:        "dependency-count-1",
			description: "Find exactly 1 dependency with verbose output",
			sourceArg:   enumDepsPrimaryRepo + "//basic.yaml",
			setupFiles: []string{
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "basic.yaml"),
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "referencing.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references basic.yaml
steps:
  - name: call_basic
    ttp: ` + enumDepsPrimaryRepo + `//basic.yaml
`,
			},
			verbose:   true,
			wantError: false,
		},
		{
			name:        "dependency-count-0",
			description: "Find no dependencies with verbose output",
			sourceArg:   enumDepsPrimaryRepo + "//standalone.yaml",
			setupFiles: []string{
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "standalone.yaml"),
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "other.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(enumDepsPrimaryRepo, enumDepsSearchPath, "other.yaml"): `---
name: Other TTP
description: Does not reference standalone
steps:
  - name: test_step
    inline: echo "no reference"
`,
			},
			verbose:   true,
			wantError: false,
		},
		{
			name:          "nonexistent-ttp",
			description:   "Attempting to enum dependencies for nonexistent TTP should fail",
			sourceArg:     enumDepsPrimaryRepo + "//nonexistent.yaml",
			setupFiles:    []string{}, // Don't create the source file
			wantError:     true,
			errorContains: "failed to resolve source TTP reference",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkEnumDepsTestCase(t, tc)
		})
	}
}
