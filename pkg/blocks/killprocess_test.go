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
	"os"
	"strconv"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestKillProcessExecute(t *testing.T) {

	testCases := []struct {
		name                string
		description         string
		step                *KillProcessStep
		stepVars            map[string]string
		createProcess       bool
		expectTemplateError bool
		expectExecuteError  bool
		expectValidateError bool
	}{
		{
			name:        "Kill non-existent process with id - throw error",
			description: "Trying to kill a process with id that doesn't exist",
			step: &KillProcessStep{
				KillProcess:               KillProcessConfig{ID: "123456789"},
				ErrorOnFindProcessFailure: true,
			},
			createProcess:      false,
			expectExecuteError: true,
		},
		{
			name:        "Kill non-existent process with id - continue on error",
			description: "Trying to kill a process with id that doesn't exist",
			step: &KillProcessStep{
				KillProcess:               KillProcessConfig{ID: "123456789"},
				ErrorOnFindProcessFailure: false,
				ErrorOnKillFailure:        false,
			},
			createProcess:      false,
			expectExecuteError: false,
		},
		{
			name:        "Kill non-existent process with name - throw error",
			description: "Trying to kill a process with name that doesn't exist",
			step: &KillProcessStep{
				KillProcess:               KillProcessConfig{Name: "ttpforge_nonexistent_process_12345"},
				ErrorOnFindProcessFailure: true,
			},
			createProcess:      false,
			expectExecuteError: true,
		},
		{
			name:        "Kill non-existent process with name - continue on error",
			description: "Trying to kill a process with name that doesn't exist",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{Name: "ttpforge_nonexistent_process_12345"},
			},
			createProcess:      false,
			expectExecuteError: false,
		},
		{
			name:        "Kill existent process with id",
			description: "Trying to kill a process with id that exists",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "123456789"},
			},
			createProcess:      true,
			expectExecuteError: false,
		},
		{
			name:        "Kill existent process with name",
			description: "Trying to kill a process with name that exists",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{Name: "ping"},
			},
			createProcess:      true,
			expectExecuteError: false,
		},
		{
			name:        "Kill process invalid id",
			description: "Trying to kill a process with negative id",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "-100"},
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process with no id and name",
			description: "Trying to kill a process with null id and name",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "", Name: ""},
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process with both id and name",
			description: "Trying to kill a process with both id and name set",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "123", Name: "ping"},
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process with char id",
			description: "Trying to kill a process with character id",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "ABCD"},
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process id with templating",
			description: "Trying to kill a process with templating",
			step: &KillProcessStep{
				KillProcess:               KillProcessConfig{ID: "{[{.StepVars.pid}]}"},
				ErrorOnFindProcessFailure: true,
			},
			stepVars: map[string]string{
				"pid": "123456789",
			},
			expectTemplateError: false,
			createProcess:       false,
			expectExecuteError:  true,
		},
		{
			name:        "Kill process name with templating",
			description: "Trying to kill a process with name and templating",
			step: &KillProcessStep{
				KillProcess:               KillProcessConfig{Name: "{[{.StepVars.processName}]}"},
				ErrorOnFindProcessFailure: true,
			},
			stepVars: map[string]string{
				"processName": "touch",
			},
			expectTemplateError: false,
			createProcess:       false,
			expectExecuteError:  true,
		},
		{
			name:        "Kill process id with templating error",
			description: "Trying to kill a process id without proper templating",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{ID: "{[{.StepVars.pid}]}"},
			},
			expectTemplateError: true,
		},
		{
			name:        "Kill process name with templating error",
			description: "Trying to kill a process name without proper templating",
			step: &KillProcessStep{
				KillProcess: KillProcessConfig{Name: "{[{.StepVars.processName}]}"},
			},
			expectTemplateError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// prep execution context with step vars
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

			// template and check error
			err := tc.step.Template(execCtx)
			if tc.expectTemplateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// validate
			err2 := tc.step.Validate(execCtx)
			if tc.expectValidateError {
				require.Error(t, err2)
				return
			}
			require.NoError(t, err2)

			if tc.createProcess {
				// Create a process to kill
				pid, error1 := testutils.CreateProcessToTerminate()
				require.NoError(t, error1)

				// set the process id for execution
				tc.step.KillProcess.ID = strconv.Itoa(pid)
				_, err = os.FindProcess(pid)
				require.NoError(t, err)
			}
			// execute and check error
			_, err = tc.step.Execute(execCtx)
			if tc.expectExecuteError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
