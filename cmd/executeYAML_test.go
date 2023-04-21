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
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runE2ETest(t *testing.T, testFile string, stepOutputs []string) {
	ttp, err := ExecuteYAML("e2e-tests/" + testFile)
	assert.Nil(t, err)

	assert.Equal(t, len(stepOutputs), len(ttp.Steps), "step outputs should have correct length")

	for stepIdx, step := range ttp.Steps {
		output := step.GetOutput()
		b, err := json.Marshal(output)
		assert.Nil(t, err)
		assert.Equal(t, stepOutputs[stepIdx], string(b), "step output is incorrect")

	}

	assert.NotNil(t, ttp)
}

func TestE2E(t *testing.T) {

	dirname, err := os.UserHomeDir()
	assert.Nil(t, err)

	scenarios := map[string][]string{
		"test_variable_expansion.yaml": {
			fmt.Sprintf("{\"output\":\"%v\"}", dirname),
			fmt.Sprintf("{\"another_key\":\"wut\",\"test_key\":\"%v\"}", dirname),
			"{\"output\":\"you said: wut\"}",
		},
	}
	for scenarioFile, stepOutputs := range scenarios {
		runE2ETest(t, scenarioFile, stepOutputs)
	}
}
