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
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandVariablesStepResults(t *testing.T) {
	// build the test fixture
	stepResults := blocks.NewStepResultsRecord()
	stepResults.ByName["first_step"] = &blocks.ExecutionResult{
		ActResult: blocks.ActResult{
			Stdout: "hello",
		},
	}
	stepResults.ByName["second_step"] = &blocks.ExecutionResult{
		ActResult: blocks.ActResult{
			Stdout: "world",
		},
	}
	stepResults.ByIndex = append(stepResults.ByIndex, stepResults.ByName["first_step"])
	stepResults.ByIndex = append(stepResults.ByIndex, stepResults.ByName["second_step"])
	execCtx := blocks.TTPExecutionContext{
		StepResults: stepResults,
	}

	// test variable expansion
	expandedStrs, err := execCtx.ExpandVariables([]string{"{{steps.first_step.stdout}} {{steps.second_step.stdout}}"})
	require.NoError(t, err)
	require.Equal(t, 1, len(expandedStrs))
	assert.Equal(t, "hello world", expandedStrs[0])
}
