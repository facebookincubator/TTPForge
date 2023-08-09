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

package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/cmd"
	"github.com/otiai10/copy"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func getTemplatesDir(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get the current working directory: %v", err)
	}

	return filepath.Join(wd, "..", "templates")
}

func createTestInventory(t *testing.T, dir string) {
	t.Helper()

	lateralMovementDir := filepath.Join(dir, "ttps", "lateral-movement", "ssh")
	if err := os.MkdirAll(lateralMovementDir, 0755); err != nil {
		t.Fatalf("failed to create lateral movement dir: %v", err)
	}

	privEscalationDir := filepath.Join(dir, "ttps", "privilege-escalation", "credential-theft", "hello-world")
	if err := os.MkdirAll(privEscalationDir, 0755); err != nil {
		t.Fatalf("failed to create privilege escalation dir: %v", err)
	}

	invalidTTPDir := filepath.Join(dir, "ttps", "invalid", "ttp")
	if err := os.MkdirAll(invalidTTPDir, 0755); err != nil {
		t.Fatalf("failed to create invalid TTP dir: %v", err)
	}

	testFiles := []struct {
		path     string
		contents string
	}{
		{
			path:     filepath.Join(lateralMovementDir, "rogue-ssh-key.yaml"),
			contents: fmt.Sprintln("---\nname: test-rogue-ssh-key-contents"),
		},
		{
			path:     filepath.Join(privEscalationDir, "priv-esc.yaml"),
			contents: fmt.Sprintln("---\nname: test-priv-esc-contents"),
		},
		{
			path:     filepath.Join(invalidTTPDir, "invalidTTP.yaml"),
			contents: fmt.Sprintln("THIS SHOULD NOT WORK"),
		},
	}

	for _, file := range testFiles {
		f, err := os.Create(file.path)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		if _, err := io.WriteString(f, file.contents); err != nil {
			t.Fatalf("failed to write to test file: %v", err)
		}
		f.Close()
	}
}

func setupTestEnvironment(t *testing.T) (string, string) {
	// Create a temporary directory
	testDir, err := os.MkdirTemp("", "cmd-new-test")
	assert.NoError(t, err, "failed to create temporary directory")

	createTestInventory(t, testDir)

	templatesDir := getTemplatesDir(t)

	if err := copy.Copy(templatesDir, filepath.Join(testDir, "templates")); err != nil {
		t.Fatalf("failed to copy templates dir: %v", err)
	}

	// Create ttp dir
	ttpDir := filepath.Join(testDir, "ttps")
	if err := os.MkdirAll(ttpDir, 0755); err != nil {
		t.Fatalf("failed to create ttps directory: %v", err)
	}

	// config for the test
	testConfigYAML := `---
inventory:
  - ` + ttpDir + `
logfile: ""
nocolor: false
verbose: false
`

	// Write the config to a temporary file
	testConfigYAMLPath := filepath.Join(testDir, "config.yaml")
	err = os.WriteFile(testConfigYAMLPath, []byte(testConfigYAML), 0644)
	assert.NoError(t, err, "failed to write the temporary YAML file")

	return testDir, testConfigYAMLPath
}

func captureStdout(t *testing.T, f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	assert.NoError(t, err, "failed to copy to buffer")
	return buf.String()
}

func TestCreateAndValidateTTP(t *testing.T) {
	testDir, testConfigYAMLPath := setupTestEnvironment(t)
	defer os.RemoveAll(testDir)

	basicTestPath := filepath.Join("ttps", "basicTest", "testBasicTTP.yaml")
	fileTestPath := filepath.Join("ttps", "fileTest", "testFileTTP.yaml")
	invalidTTPPath := filepath.Join("ttps", "invalid", "ttp", "invalidTTP.yaml")

	newTTPBuilderCmd := cmd.NewTTPBuilderCmd()
	newRunTTPCmd := cmd.RunTTPCmd()
	testCases := []struct {
		name             string
		setFlags         func()
		input            cmd.NewTTPInput
		expected         string
		expectError      bool
		expectedErrorMsg []string
	}{
		{
			name: "Create basic bash TTP",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", testConfigYAMLPath)
				_ = newTTPBuilderCmd.Flags().Set("path", basicTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "bash")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "basic")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "false")
				_ = newTTPBuilderCmd.Flags().Set("env", "EXAMPLE_ENV_VAR=example_value")
			},
			input: cmd.NewTTPInput{
				Path: basicTestPath,
			},
			expected: "Hello, World\n",
		},
		{
			name: "Create file-based bash TTP",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", testConfigYAMLPath)
				_ = newTTPBuilderCmd.Flags().Set("path", fileTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "bash")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "file")
				_ = newTTPBuilderCmd.Flags().Set("args", "arg1,arg2,arg3")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "true")
				_ = newTTPBuilderCmd.Flags().Set("env", "EXAMPLE_ENV_VAR=example_value")
			},
			input: cmd.NewTTPInput{
				Path: fileTestPath,
			},
			expected: "Now running testFileTTP.yaml, please wait\n\n",
		},
		{
			name: "Fail to create TTP with missing fields",
			setFlags: func() {
				_ = newTTPBuilderCmd.Flags().Set("config", testConfigYAMLPath)
				_ = newTTPBuilderCmd.Flags().Set("path", fileTestPath)
				_ = newTTPBuilderCmd.Flags().Set("template", "asdf")
				_ = newTTPBuilderCmd.Flags().Set("ttp-type", "file")
				_ = newTTPBuilderCmd.Flags().Set("args", "arg1,arg2,arg3")
				_ = newTTPBuilderCmd.Flags().Set("cleanup", "true")
				_ = newTTPBuilderCmd.Flags().Set("env", "EXAMPLE_ENV_VAR=example_value")
			},
			input: cmd.NewTTPInput{
				Path: fileTestPath,
			},
			expectError: true,
		},
		{
			name: "Detect invalid input TTP",
			setFlags: func() {
				_ = newRunTTPCmd.Flags().Set("config", testConfigYAMLPath)
				_ = newRunTTPCmd.Flags().Set("template", "bash")
				_ = newRunTTPCmd.Flags().Set("ttp-type", "basic")
				_ = newRunTTPCmd.Flags().Set("cleanup", "false")
			},
			input: cmd.NewTTPInput{
				Path: invalidTTPPath,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set flags for the test case
			tc.setFlags()

			// Update tc.input.TTPType value
			ttpTypeFlag, err := newTTPBuilderCmd.Flags().GetString("ttp-type")
			if err != nil {
				t.Fatalf("failed to get ttp-type flag: %v", err)
			}
			tc.input.TTPType = ttpTypeFlag

			testRoot := filepath.Dir(testConfigYAMLPath)

			if err := os.Chdir(testRoot); err != nil {
				t.Fatalf("failed to change into test directory: %v", err)
			}
			// Set filepath for current test TTP
			ttpPath := basicTestPath
			if tc.input.TTPType == "file" {
				ttpPath = fileTestPath
			}

			// Create the test TTP directory if it doesn't already exist
			if err := os.MkdirAll(filepath.Dir(ttpPath), 0755); err != nil {
				t.Fatalf("failed to create ttps directory: %v", err)
			}

			// Create the test TTP file
			ttpFile, err := os.Create(ttpPath)
			if err != nil {
				t.Fatalf("failed to create test ttp: %v", err)
			}
			if _, err := io.WriteString(ttpFile, fmt.Sprintln("---\nname: test-ttp-contents")); err != nil {
				t.Fatalf("failed to write to test ttp: %v", err)
			}
			defer ttpFile.Close()

			// Reset flags
			newTTPBuilderCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				_ = newTTPBuilderCmd.Flags().Set(flag.Name, "")
			})

			// Set flags for the test case
			tc.setFlags()

			err = newTTPBuilderCmd.Execute()
			if tc.expectError {
				if tc.input.Template != "" {
					require.Error(t, err)
				}
			} else {
				require.NoError(t, err)
				switch tc.input.TTPType {
				case "basic":
					_, err = os.Stat(basicTestPath)
					assert.NoErrorf(t, err, "failed to create path %s: %v", basicTestPath, err)
				case "file":
					_, err = os.Stat(fileTestPath)
					assert.NoErrorf(t, err, "failed to create path %s: %v", fileTestPath, err)
				default:
					t.Fatal("invalid TTPType provided")
				}
			}

			// Check if the bash script file was created (for file TTP type)
			if tc.input.TTPType == "file" {
				bashTTPPath := filepath.Join(filepath.Dir(tc.expected), "bashTTP.bash")
				_, err = os.Stat(bashTTPPath)
				if tc.input.Template != "" {
					assert.False(t, os.IsNotExist(err), "bashTTP.bash file not found: %s", bashTTPPath)
				}
			}

			// Check if the README was created
			readmePath := filepath.Join(filepath.Dir(tc.expected), "README.md")
			_, err = os.Stat(readmePath)
			if tc.input.Template != "" {
				assert.False(t, os.IsNotExist(err), "README.md file not found: %s", readmePath)
			}

			// Run the created TTP
			output := captureStdout(t, func() {
				runCmd := cmd.RunTTPCmd()
				runCmd.SetArgs([]string{tc.input.Path})
				err := runCmd.Execute()
				// assert.Equal(t, tc.expected, output, "The output of the executed TTP script does not match the expected output")
				require.NoError(t, err, fmt.Sprintf("failed to run TTP: %v", err))
			})

			assert.Equal(t, tc.expected, output, "The output of the executed TTP script does not match the expected output")

			// Cleanup
			os.RemoveAll(filepath.Dir(tc.expected))
		})
	}
}
