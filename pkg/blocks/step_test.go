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
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func TestStep(t *testing.T) {
	testCases := []struct {
		name                  string
		content               string
		wantUnmarshalError    bool
		wantValidateError     bool
		wantTemplateError     bool
		wantExecuteError      bool
		expectedExecuteStdout string
		wantCleanupError      bool
		expectedCleanupStdout string
		stepVars              map[string]string
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
		{
			name: "Basic bash executor doesn't tolerate non-zero exit codes in inline scripts",
			content: `name: inline_step
description: this is a test
inline: |
  false
  echo executing
cleanup:
  inline: echo cleanup
`,
			wantExecuteError:      true,
			expectedExecuteStdout: "",
			expectedCleanupStdout: "cleanup\n",
		},
		{
			name: "Basic bash supports setting error processing option to ignore errors",
			content: `name: inline_step
inline: |
  set +e
  false
  echo executing
cleanup:
  inline: echo cleanup
`,
			wantExecuteError:      false,
			expectedExecuteStdout: "executing\n",
			expectedCleanupStdout: "cleanup\n",
		},
		{
			name: "name: copypath test, copy from an existing file to nonexisting with overwrite.",
			content: `name: copypath_step
copy_path: /etc/passwd
to: /tmp/passwd
overwrite: true
cleanup: default`,
			wantExecuteError: false,
		},
		{
			name: "Templates both step and cleanup",
			content: `
name: template_step
inline: echo {[{.StepVars.run_message}]}
cleanup:
  inline: echo {[{.StepVars.cleanup_message}]}
`,
			stepVars: map[string]string{
				"run_message":     "this is a run",
				"cleanup_message": "this is a cleanup",
			},
			expectedExecuteStdout: "this is a run\n",
			expectedCleanupStdout: "this is a cleanup\n",
		},
		{
			name: "Errors on missing variable in step templating",
			content: `
name: template_step
inline: echo {[{.StepVars.run_message}]}
cleanup:
  inline: echo "this is a cleanup"
`,
			wantTemplateError: true,
		},
		{
			name: "Errors on missing variable in cleanup templating at cleanup time",
			content: `
name: template_step
inline: echo "this is a run"
cleanup:
  inline: echo {[{.StepVars.cleanup_message}]}
`,
			expectedExecuteStdout: "this is a run\n",
			wantCleanupError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s Step
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

			// parse the step
			err := yaml.Unmarshal([]byte(tc.content), &s)
			if tc.wantUnmarshalError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// validate the step
			err = s.Validate(execCtx)
			if tc.wantValidateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// template the step
			err = s.Template(execCtx)
			if tc.wantTemplateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// execute the step and check output
			result, err := s.Execute(execCtx)
			if tc.wantExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedExecuteStdout, result.Stdout)

			// run cleanup and check output
			cleanupResult, err := s.Cleanup(execCtx)
			if tc.wantCleanupError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCleanupStdout, cleanupResult.Stdout)
		})
	}
}

func TestCleanupDefault(t *testing.T) {
	testCases := []struct {
		name                        string
		contentFmtStr               string
		wantUnmarshalError          bool
		wantExecuteError            bool
		expectedFileContents        string
		wantCleanupError            bool
		fileShouldExistAfterCleanup bool
	}{
		{
			name: "create_file Default Cleanup",
			contentFmtStr: `name: create_file_step
description: creates a file and then deletes it
create_file: %v
contents: this is a test
cleanup: default`,
			expectedFileContents: "this is a test",
		},
		{
			name: "create_file with invalid cleanup",
			contentFmtStr: `name: create_file_step
description: invalid cleanup value
create_file: %v
contents: this is a test
cleanup: invalid`,
			wantUnmarshalError: true,
		},
		{
			name: "create_file with non-default cleanup",
			contentFmtStr: `name: create_file_step
description: non-default cleanup value
create_file: %v
contents: testing non default cleanup
cleanup:
  inline: echo "will not delete file"`,
			expectedFileContents:        "testing non default cleanup",
			fileShouldExistAfterCleanup: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s Step
			execCtx := NewTTPExecutionContext()

			// hack to get a valid temporary path without creating it
			tmpFile, err := os.CreateTemp("", "ttpforge-test-cleanup-default")
			require.NoError(t, err)
			filePath := tmpFile.Name()
			err = os.Remove(filePath)
			require.NoError(t, err)

			content := fmt.Sprintf(tc.contentFmtStr, filePath)
			err = yaml.Unmarshal([]byte(content), &s)
			if tc.wantUnmarshalError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// validate the step
			err = s.Validate(execCtx)
			require.NoError(t, err)

			// execute the step and check file contents
			_, err = s.Execute(execCtx)
			if tc.wantExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			contentBytes, err := os.ReadFile(filePath)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedFileContents, string(contentBytes))

			// run cleanup
			_, err = s.Cleanup(execCtx)
			if tc.wantCleanupError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// verify that file was deleted
			fsys := afero.NewOsFs()
			exists, err := afero.Exists(fsys, filePath)
			require.NoError(t, err)
			assert.Equal(t, tc.fileShouldExistAfterCleanup, exists)
		})
	}
}
