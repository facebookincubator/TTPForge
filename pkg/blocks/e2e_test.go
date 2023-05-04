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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ScenarioResult struct {
	StepOutputs    []string
	CleanupOutputs []string
}

// this is broken out from runE2ETest so that a given
// test case can tweak the TTP after loading as needed
func loadEndToEndTestTTP(t *testing.T, cfg TTPConfig) *TTP {
	cfg.RelativePath = filepath.Join("e2e-test-ttps", cfg.RelativePath)
	ttp, err := NewTTPFromConfig(cfg)
	require.Nil(t, err)
	return ttp
}

func runE2ETest(t *testing.T, ttp *TTP, expectedResult ScenarioResult) {
	err := ttp.RunSteps()
	assert.Nil(t, err)
	assert.NotNil(t, ttp)

	expectedStepOutputs := expectedResult.StepOutputs
	assert.Equal(t, len(expectedStepOutputs), len(ttp.Steps), "step outputs should have correct length")

	// check step outputs
	var cleanupOutputs []string
	for stepIdx, step := range ttp.Steps {
		output := step.GetOutput()
		b, err := json.Marshal(output)
		assert.Nil(t, err)
		assert.Equal(t, expectedStepOutputs[stepIdx], string(b), "step output is incorrect")
		cleanups := step.GetCleanup()
		for _, cleanup := range cleanups {
			// cleanups that weren't run (bcs NoCleanup) have nil output
			cleanupOutput := cleanup.GetOutput()
			if cleanupOutput == nil {
				continue
			}
			cb, err := json.Marshal(cleanupOutput)
			assert.Nil(t, err)
			// put them in reverse order
			cleanupOutputs = append([]string{string(cb)}, cleanupOutputs...)
		}
	}

	// check cleanup outputs
	expectedCleanupOutputs := expectedResult.CleanupOutputs
	assert.Equal(t, len(expectedCleanupOutputs), len(cleanupOutputs), "Number of cleanup steps does not match expectation")
	for cleanupIdx, cleanupOutput := range cleanupOutputs {
		assert.Equal(t, expectedCleanupOutputs[cleanupIdx], cleanupOutput)
	}
}

func TestVariableExpansion(t *testing.T) {
	// something dynamic to keep things interesting
	dirname, err := os.UserHomeDir()
	assert.Nil(t, err)

	ttp := loadEndToEndTestTTP(t, TTPConfig{
		RelativePath: filepath.Join("variable-expansion", "ttp.yaml"),
	})
	runE2ETest(t, ttp, ScenarioResult{
		StepOutputs: []string{
			fmt.Sprintf("{\"output\":\"%v\"}", dirname),
			fmt.Sprintf("{\"another_key\":\"wut\",\"test_key\":\"%v\"}", dirname),
			"{\"output\":\"you said: wut\"}",
		},
		CleanupOutputs: []string{
			"{\"output\":\"cleaning up now\"}",
		},
	})
}

func TestRelativePaths(t *testing.T) {
	ttp := loadEndToEndTestTTP(t, TTPConfig{
		RelativePath: filepath.Join("relative-paths", "nested.yaml"),
	})
	runE2ETest(t, ttp, ScenarioResult{
		StepOutputs: []string{
			"{\"output\":\"A\"}",
			"{\"output\":\"B\"}",
			"{\"output\":\"D\"}",
		},
		CleanupOutputs: []string{
			"{\"output\":\"E\"}",
			"{\"output\":\"C\"}",
		},
	})
}

func TestNoCleanup(t *testing.T) {

	// need an explicit workdir
	// bcs we will clean it up manually
	wd := "TestNoCleanup-WorkDir"
	err := os.Mkdir(wd, 0700)
	assert.Nil(t, err)
	defer os.RemoveAll(wd)

	ttp := loadEndToEndTestTTP(t, TTPConfig{
		RelativePath: filepath.Join("relative-paths", "nested.yaml"),
		NoCleanup:    true,
	})

	runE2ETest(t, ttp, ScenarioResult{
		StepOutputs: []string{
			"{\"output\":\"A\"}",
			"{\"output\":\"B\"}",
			"{\"output\":\"D\"}",
		},
	})
}
