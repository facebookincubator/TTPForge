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
	"strings"
	"testing"
)

const (
	workingDir        = "ttpforge-move-test"
	primaryRepoName   = "test-repo"
	secondaryRepoName = "new-repo"
	searchPath        = "ttps"
)

type moveCmdTestCase struct {
	name              string
	description       string
	sourceArg         string
	destArg           string
	wantError         bool
	errorContains     string
	expectedNewPath   string
	expectedOldPath   string
	setupFiles        []string
	referencingFiles  map[string]string // path -> content that references the source file
	expectUpdatedRefs map[string]string // path -> expected updated content
}

func setupMoveTestFiles(t *testing.T, fsys afero.Fs, baseDir string, files []string) {
	for _, file := range files {
		fullPath := filepath.Join(baseDir, file)
		dir := filepath.Dir(fullPath)
		err := fsys.MkdirAll(dir, 0755)
		require.NoError(t, err)

		content := `---
name: Test TTP
description: A test TTP for move command testing
steps:
  - name: test_step
    inline: echo "test output"
`
		err = afero.WriteFile(fsys, fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func setupReferencingFiles(t *testing.T, fsys afero.Fs, baseDir string, files map[string]string) {
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		dir := filepath.Dir(fullPath)
		err := fsys.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = afero.WriteFile(fsys, fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func checkMoveTestCase(t *testing.T, tc moveCmdTestCase) {
	// Create a temporary directory for this test
	fsys := afero.NewOsFs()
	tempDir, err := afero.TempDir(fsys, "", workingDir)
	require.NoError(t, err)
	defer fsys.RemoveAll(tempDir)

	// Create test repository structure
	primaryRepoDir := filepath.Join(tempDir, primaryRepoName)
	secondaryRepoDir := filepath.Join(tempDir, secondaryRepoName)
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Create repo config
	repoConfigContent := `---
ttp_search_paths:
  - ` + searchPath + `
`
	err = fsys.MkdirAll(filepath.Join(primaryRepoDir, "ttps"), 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fsys, filepath.Join(primaryRepoDir, "ttpforge-repo-config.yaml"), []byte(repoConfigContent), 0644)
	require.NoError(t, err)

	err = fsys.MkdirAll(filepath.Join(secondaryRepoDir, "ttps"), 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fsys, filepath.Join(secondaryRepoDir, "ttpforge-repo-config.yaml"), []byte(repoConfigContent), 0644)
	require.NoError(t, err)

	// Create main config
	mainConfigContent := `---
repos:
  - name: ` + primaryRepoName + `
    path: ` + primaryRepoDir + `
  - name: ` + secondaryRepoName + `
    path: ` + secondaryRepoDir + `
`
	err = afero.WriteFile(fsys, configPath, []byte(mainConfigContent), 0644)
	require.NoError(t, err)

	// Setup test files
	if tc.setupFiles != nil {
		setupMoveTestFiles(t, fsys, tempDir, tc.setupFiles)
	}

	// Setup referencing files
	if tc.referencingFiles != nil {
		setupReferencingFiles(t, fsys, tempDir, tc.referencingFiles)
	}

	// Convert reference paths to absolute paths for test repo when needed
	sourceArg := tc.sourceArg
	destArg := tc.destArg

	// Handle test cases that use actual file paths instead of repo references
	// This converts paths like primaryRepoName + searchPath + "basic.yaml" to absolute paths in our test repo
	if !strings.Contains(tc.sourceArg, "//") && !filepath.IsAbs(tc.sourceArg) && tc.sourceArg != "" {
		sourceArg = filepath.Join(tempDir, tc.sourceArg)
	}
	if !strings.Contains(tc.destArg, "//") && !filepath.IsAbs(tc.destArg) && tc.destArg != "" {
		destArg = filepath.Join(tempDir, tc.destArg)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	rc := BuildRootCommand(&TestConfig{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})

	args := []string{"move", "-c", configPath, sourceArg, destArg}
	rc.SetArgs(args)

	err = rc.Execute()

	if tc.wantError {
		require.Error(t, err)
		if tc.errorContains != "" {
			assert.Contains(t, err.Error(), tc.errorContains)
		}
		return
	}

	require.NoError(t, err, "Move command should succeed")

	// Check that the file was moved to the expected location
	if tc.expectedNewPath != "" {
		expectedPath := filepath.Join(tempDir, tc.expectedNewPath)
		exists, err := afero.Exists(fsys, expectedPath)
		require.NoError(t, err)
		assert.True(t, exists, "File should exist at new location: %s", expectedPath)
	}

	// Check that the file was removed from the old location
	if tc.expectedOldPath != "" {
		oldPath := filepath.Join(tempDir, tc.expectedOldPath)
		exists, err := afero.Exists(fsys, oldPath)
		require.NoError(t, err)
		assert.False(t, exists, "File should not exist at old location: %s", oldPath)
	}

	// Check that referencing files were updated correctly
	if tc.expectUpdatedRefs != nil {
		for path, expectedContent := range tc.expectUpdatedRefs {
			fullPath := filepath.Join(tempDir, path)
			content, err := afero.ReadFile(fsys, fullPath)
			require.NoError(t, err)
			assert.Equal(t, expectedContent, string(content), "File %s should have updated references", path)
		}
	}
}

func TestMoveCommand(t *testing.T) {
	testCases := []moveCmdTestCase{
		{
			name:        "reference-to-reference",
			description: "Move TTP from reference path to reference path",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "reference-to-absolute",
			description: "Move TTP from reference path to absolute path",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "absolute-to-reference",
			description: "Move TTP from absolute path to reference path",
			sourceArg:   filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			destArg:     primaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "absolute-to-absolute",
			description: "Move TTP from absolute path to absolute path",
			sourceArg:   filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			destArg:     filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "nested-directory-move",
			description: "Move TTP to nested directory structure",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//subdir/nested/moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "subdir/nested/moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "move-with-references",
			description: "Move TTP and update references in other TTPs",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP
steps:
  - name: run_basic
    ttp: ` + primaryRepoName + `//basic.yaml
`,
			},
			expectUpdatedRefs: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP
steps:
  - name: run_basic
    ttp: //moved.yaml
`,
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "move-with-scoped-references",
			description: "Move TTP and update scoped references (without repo prefix)",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP with scoped reference
steps:
  - name: run_basic
    ttp: //basic.yaml
`,
			},
			expectUpdatedRefs: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP with scoped reference
steps:
  - name: run_basic
    ttp: //moved.yaml
`,
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "move-with-multiple-references",
			description: "Move TTP and update multiple references across different files",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
				filepath.Join(primaryRepoName, searchPath, "ref1.yaml"),
				filepath.Join(primaryRepoName, searchPath, "ref2.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "ref1.yaml"): `---
name: First Reference
steps:
  - name: call_basic
    ttp: ` + primaryRepoName + `//basic.yaml
`,
				filepath.Join(primaryRepoName, searchPath, "ref2.yaml"): `---
name: Second Reference
steps:
  - name: also_call_basic
    ttp: //basic.yaml
  - name: call_something_else
    ttp: ` + primaryRepoName + `//other.yaml
`,
			},
			expectUpdatedRefs: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "ref1.yaml"): `---
name: First Reference
steps:
  - name: call_basic
    ttp: //moved.yaml
`,
				filepath.Join(primaryRepoName, searchPath, "ref2.yaml"): `---
name: Second Reference
steps:
  - name: also_call_basic
    ttp: //moved.yaml
  - name: call_something_else
    ttp: ` + primaryRepoName + `//other.yaml
`,
			},
			wantError:       false,
			expectedNewPath: filepath.Join(primaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
		{
			name:        "cross-repository-move",
			description: "Move TTP into a different repository and update references",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     secondaryRepoName + "//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"),
			},
			referencingFiles: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP with scoped reference
steps:
  - name: run_basic
    ttp: //basic.yaml
`,
			},
			expectUpdatedRefs: map[string]string{
				filepath.Join(primaryRepoName, searchPath, "referencing.yaml"): `---
name: Referencing TTP
description: A TTP that references another TTP with scoped reference
steps:
  - name: run_basic
    ttp: ` + secondaryRepoName + `//moved.yaml
`,
			},
			wantError:       false,
			expectedNewPath: filepath.Join(secondaryRepoName, searchPath, "moved.yaml"),
			expectedOldPath: filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkMoveTestCase(t, tc)
		})
	}
}

func TestMoveCommandEdgeCases(t *testing.T) {
	edgeCases := []moveCmdTestCase{
		{
			name:          "missing-arguments",
			description:   "Move command requires exactly 2 arguments",
			sourceArg:     "", // Will be ignored since we're not passing any args
			destArg:       "",
			wantError:     true,
			errorContains: "accepts 2 arg(s), received 0",
		},
		{
			name:        "same-source-and-destination",
			description: "Moving a file to itself should fail",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//basic.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:     true,
			errorContains: "already exists",
		},
		{
			name:          "invalid-source-format",
			description:   "Invalid source reference format should fail",
			sourceArg:     "invalid//format//too//many//separators.yaml",
			destArg:       primaryRepoName + "//moved.yaml",
			wantError:     true,
			errorContains: "too many occurrences",
		},
		{
			name:        "invalid-dest-format",
			description: "Invalid destination reference format should fail",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     "invalid//format//too//many//separators.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:     true,
			errorContains: "too many occurrences",
		},
		{
			name:        "invalid-destination-repo",
			description: "Moving to a non-existent repository should fail",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     "nonexistent-repo//moved.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
			},
			wantError:     true,
			errorContains: "repository 'nonexistent-repo' not found",
		},
		{
			name:          "nonexistent-source",
			description:   "Attempting to move a nonexistent TTP should fail",
			sourceArg:     primaryRepoName + "//nonexistent.yaml",
			destArg:       primaryRepoName + "//moved.yaml",
			setupFiles:    []string{}, // Don't create the source file
			wantError:     true,
			errorContains: "failed to resolve source TTP reference",
		},
		{
			name:        "destination-exists",
			description: "Attempting to move to an existing destination should fail",
			sourceArg:   primaryRepoName + "//basic.yaml",
			destArg:     primaryRepoName + "//existing.yaml",
			setupFiles: []string{
				filepath.Join(primaryRepoName, searchPath, "basic.yaml"),
				filepath.Join(primaryRepoName, searchPath, "existing.yaml"),
			},
			wantError:     true,
			errorContains: "already exists",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "missing-arguments" {
				// Special case for missing arguments test
				var stdoutBuf, stderrBuf bytes.Buffer
				rc := BuildRootCommand(&TestConfig{
					Stdout: &stdoutBuf,
					Stderr: &stderrBuf,
				})
				rc.SetArgs([]string{"move"}) // No arguments
				err := rc.Execute()
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				return
			}
			checkMoveTestCase(t, tc)
		})
	}
}
