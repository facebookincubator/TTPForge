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

func TestPrintStrExecute(t *testing.T) {
	testCases := []struct {
		name               string
		description        string
		action             *blocks.PrintStrAction
		expectExecuteError bool
		expectedStdout     string
	}{
		{
			name:        "Simple Print",
			description: "Just Print a String",
			action: &blocks.PrintStrAction{
				Message: "hello",
			},
			expectedStdout: "hello\n",
		},
		{
			name:        "Print Step Output",
			description: "Should be Expanded",
			action: &blocks.PrintStrAction{
				Message: "value is $forge.steps.first_step.stdout",
			},
			expectedStdout: "value is first-step-output\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// for use testing output variables
			execCtx := blocks.TTPExecutionContext{
				StepResults: &blocks.StepResultsRecord{
					ByName: map[string]*blocks.ExecutionResult{
						"first_step": {
							ActResult: blocks.ActResult{
								Stdout: "first-step-output",
							},
						},
					},
				},
			}

			// execute and check error
			result, err := tc.action.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check stdout
			assert.Equal(t, tc.expectedStdout, result.Stdout)
		})
	}
}
