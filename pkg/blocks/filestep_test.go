package blocks_test

import (
	"runtime"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

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
