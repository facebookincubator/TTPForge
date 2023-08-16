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
