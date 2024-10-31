/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
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
	"context"
	"testing"
)

func TestBashExecutor(t *testing.T) {
	emptyEnvironment := map[string]string{}
	execCtx := TTPExecutionContext{Vars: &TTPExecutionVars{}}

	testCases := []struct {
		name           string
		executorName   string
		body           string
		expectedResult string
		expectedErrTxt string
	}{
		{
			name:           "bash ok",
			executorName:   "bash",
			body:           "echo success",
			expectedResult: "success\n",
			expectedErrTxt: "",
		},
		{
			name:           "bash fail fast",
			executorName:   "bash",
			body:           "false; echo success",
			expectedResult: "",
			expectedErrTxt: "exit status 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executor := NewExecutor(tc.executorName, tc.body, "", []string{}, emptyEnvironment)
			result, err := executor.Execute(context.Background(), execCtx)

			if tc.expectedErrTxt != "" {
				if err == nil {
					t.Fatalf("expected error, got nil")
				} else if err.Error() != tc.expectedErrTxt {
					t.Fatalf("expected %v error, got %v", tc.expectedErrTxt, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
			if tc.expectedResult != "" && result.Stdout != tc.expectedResult {
				t.Fatalf("expected output %#v, got %#v", tc.expectedResult, result.Stdout)
			}
		})
	}
}
