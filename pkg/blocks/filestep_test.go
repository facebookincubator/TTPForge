package blocks_test

import (
	"runtime"
	"testing"

	"github.com/facebookincubator/TTP-Runner/pkg/blocks"
	"github.com/facebookincubator/TTP-Runner/pkg/logging"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}
func TestUnmarshalSimpleFile(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
description: this is a test
steps:
  - name: test_file
    file: test_file
  `

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Errorf("failed to unmarshal file step %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)
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
			if executor != testCase.expectedExec {
				t.Errorf("Expected executor %q for file path %q, but got %q", testCase.expectedExec, testCase.filePath, executor)
			}
		})
	}
}

func getDefaultExecutor() string {
	if runtime.GOOS == "windows" {
		return blocks.ExecutorCmd
	}
	return blocks.ExecutorSh
}
