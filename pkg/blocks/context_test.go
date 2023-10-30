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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandVariablesStepResults(t *testing.T) {
	// build the test fixture used across all cases
	stepResults := NewStepResultsRecord()
	stepResults.ByName["first_step"] = &ExecutionResult{
		ActResult: ActResult{
			Stdout: "hello",
		},
	}
	stepResults.ByName["second_step"] = &ExecutionResult{
		ActResult: ActResult{
			Stdout: "world",
		},
	}
	stepResults.ByName["third_step"] = &ExecutionResult{
		ActResult: ActResult{
			Stdout: `{"foo":{"bar":"baz"}}`,
			Outputs: map[string]string{
				"myresult": "baz",
			},
		},
	}
	stepResults.ByIndex = append(stepResults.ByIndex, stepResults.ByName["first_step"])
	stepResults.ByIndex = append(stepResults.ByIndex, stepResults.ByName["second_step"])
	stepResults.ByIndex = append(stepResults.ByIndex, stepResults.ByName["third_step"])
	execCtx := TTPExecutionContext{
		Cfg: TTPExecutionConfig{
			Args: map[string]any{
				"arg1": "myarg1",
				"arg2": "myarg2",
			},
		},
		StepResults: stepResults,
	}

	// individual test cases that use the above fixture
	testCases := []struct {
		name            string
		stringsToExpand []string
		expectedResult  []string
		wantError       bool
	}{
		{
			name: "Step Stdout Expansion",
			stringsToExpand: []string{
				"first: $forge.steps.first_step.stdout",
				"second: $forge.steps.second_step.stdout",
			},
			expectedResult: []string{
				"first: hello",
				"second: world",
			},
			wantError: false,
		},
		{
			name: "Step Output Expansion - JSON",
			stringsToExpand: []string{
				"third: $forge.steps.third_step.outputs.myresult",
			},
			expectedResult: []string{
				"third: baz",
			},
			wantError: false,
		},
		{
			name: "Escape forge magic string",
			stringsToExpand: []string{
				"this should be escaped: $$forge.foo.bar.baz",
				"and so should this: $$$forge.a.b.c",
			},
			expectedResult: []string{
				"this should be escaped: $forge.foo.bar.baz",
				"and so should this: $$forge.a.b.c",
			},
			wantError: false,
		},
		{
			name: "Empty Variable Specifier",
			stringsToExpand: []string{
				"this is empty: $forge.)",
			},
			wantError: true,
		},
		{
			name: "Trailing dot in variable expression",
			stringsToExpand: []string{
				"this is wrong: $forge.steps.wut.",
			},
			wantError: true,
		},
		{
			name: "Invalid Step Name",
			stringsToExpand: []string{
				"first: $forge.steps.first_step.stdout",
				"second: $forge.steps.fakestep.stdout",
			},
			wantError: true,
		},
		{
			name: "Invalid Variable Path Prefix",
			stringsToExpand: []string{
				"should fail: $forge.notreal.first_step",
			},
			wantError: true,
		},
		{
			name: "Invalid Output Key",
			stringsToExpand: []string{
				"should fail: $forge.steps.third_step.outputs.fail",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// test variable expansion
			expandedStrs, err := execCtx.ExpandVariables(tc.stringsToExpand)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tc.expectedResult), len(expandedStrs), "returned slice should have correct length")
			assert.Equal(t, tc.expectedResult, expandedStrs, "returned slice should match expected value")
		})
	}
}
