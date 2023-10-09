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

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	testCases := []struct {
		name                  string
		content               string
		wantUnmarshalError    bool
		wantExecuteError      bool
		expectedExecuteStdout string
		wantCleanupError      bool
		expectedCleanupStdout string
	}{
		{
			name: "Run inline command (no error)",
			content: `name: inline_step
description: runs a valid inline command
inline: echo inline_step_test`,
			expectedExecuteStdout: "inline_step_test\n",
		},
		{
			name: "Run Cleanup (inline - no error)",
			content: `name: inline_step
description: runs an invalid inline command
inline: echo executing
cleanup:
  inline: echo cleanup`,
			expectedExecuteStdout: "executing\n",
			expectedCleanupStdout: "cleanup\n",
		},
		{
			name: "Run inline command (execution error)",
			content: `name: inline_step
description: runs an invalid inline command
inline: this will error`,
			wantExecuteError: true,
		},
		{
			name:               "Step With Empty Name",
			content:            `inline: echo should_error_before_execution`,
			wantUnmarshalError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s blocks.Step
			var execCtx blocks.TTPExecutionContext

			// parse the step
			err := yaml.Unmarshal([]byte(tc.content), &s)
			if tc.wantUnmarshalError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// validate the step
			err = s.Validate(execCtx)
			require.NoError(t, err)

			// execute the step and check output
			result, err := s.Execute(execCtx)
			if tc.wantExecuteError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedExecuteStdout, result.Stdout)

			// run cleanup and check output
			cleanupResult, err := s.Cleanup(execCtx)
			if tc.wantCleanupError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCleanupStdout, cleanupResult.Stdout)
		})
	}
}
