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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalFile(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple file",
			content: `name: test
description: this is a test
steps:
  - name: test_file
    file: test_file
  `,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInferExecutor(t *testing.T) {
	testCases := []struct {
		filePath     string
		expectedExec string
	}{
		{filePath: "script.sh", expectedExec: blocks.ExecutorSh},
		{filePath: "script.py", expectedExec: blocks.ExecutorPython},
		{filePath: "script.rb", expectedExec: blocks.ExecutorRuby},
		{filePath: "script.pwsh", expectedExec: blocks.ExecutorPowershell},
		{filePath: "script.ps1", expectedExec: blocks.ExecutorPowershell},
		{filePath: "script.bat", expectedExec: blocks.ExecutorCmd},
		{filePath: "binary", expectedExec: blocks.ExecutorBinary},
		{filePath: "unknown.xyz", expectedExec: getDefaultExecutor()},
		{filePath: "", expectedExec: blocks.ExecutorBinary},
	}

	for _, testCase := range testCases {
		t.Run(testCase.filePath, func(t *testing.T) {
			executor := blocks.InferExecutor(testCase.filePath)
			assert.Equal(t, testCase.expectedExec, executor, "Expected executor %q for file path %q, but got %q", testCase.expectedExec, testCase.filePath, executor)
		})
	}
}

func getDefaultExecutor() string {
	if runtime.GOOS == "windows" {
		return blocks.ExecutorCmd
	}
	return blocks.ExecutorSh
}

func TestFileStepUnmarshalIgnoreErrors(t *testing.T) {
	data := `
name: testFileStep
file: expect-fail-example.sh
ignore_errors: true
`
	step := &blocks.FileStep{}
	err := yaml.Unmarshal([]byte(data), step)
	assert.NoError(t, err)
	assert.True(t, step.IgnoreErrors)
}

func createInvalidScript() (string, error) {
	tmpDir := os.TempDir()

	// Create the script that is designed to fail
	scriptPath := filepath.Join(tmpDir, "expect-fail-example.sh")
	content := `#!/bin/bash
set -e

exit 1
`
	err := os.WriteFile(scriptPath, []byte(content), 0755)
	return scriptPath, err
}

func TestFileStepIgnoreErrors(t *testing.T) {
	scriptPath, err := createInvalidScript()
	if err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer os.Remove(scriptPath) // Cleanup after test

	data := `
steps:
  - name: file-step-expected-to-fail
    ignore_errors: true
    file: ` + scriptPath + `
  - name: file-step-post-error
    inline: |
      echo -e "We still reach this step despite there being an error"
      echo -e "in the previous file step."
`
	var ttps blocks.TTP
	err = yaml.Unmarshal([]byte(data), &ttps)
	assert.NoError(t, err)

	step := ttps.Steps[0].(*blocks.FileStep) // Cast to FileStep

	// Execute the FileStep and expect no error because of the `ignore_errors` flag
	ctx := blocks.TTPExecutionContext{}
	_, err = step.Execute(ctx)
	assert.NoError(t, err) // Since IgnoreErrors is true, we shouldn't get an error.
}

func TestExecuteFileWithoutIgnoreErrors(t *testing.T) {
	step := &blocks.FileStep{
		Act: &blocks.Act{
			Type: blocks.StepFile,
			Name: "errorFileStep",
		},
		Executor:     "bash",
		FilePath:     "/non/existent/path/error.sh", // This path will surely cause an error.
		IgnoreErrors: false,
	}

	ctx := blocks.TTPExecutionContext{}

	_, err := step.Execute(ctx)
	assert.Error(t, err) // Since IgnoreErrors is false, we should get an error.
}

func createValidScript() (string, error) {
	tmpDir := os.TempDir()

	// Create a valid script
	scriptPath := filepath.Join(tmpDir, "valid-file.sh")
	content := `#!/bin/bash
echo "This is valid."
`
	err := os.WriteFile(scriptPath, []byte(content), 0755) // Make it executable
	return scriptPath, err
}

func TestValidExecuteFileWithIgnoreErrors(t *testing.T) {
	validFilePath, err := createValidScript()
	if err != nil {
		t.Fatalf("failed to create valid script: %v", err)
	}
	defer os.Remove(validFilePath) // Cleanup after test

	step := &blocks.FileStep{
		Act: &blocks.Act{
			Type: blocks.StepFile,
			Name: "validFileStep",
		},
		Executor:     "bash",
		FilePath:     validFilePath,
		IgnoreErrors: true,
	}

	ctx := blocks.TTPExecutionContext{}

	result, err := step.Execute(ctx)
	assert.NoError(t, err) // Even if IgnoreErrors is true, we shouldn't get an error since the step is valid.
	assert.NotNil(t, result)
}
