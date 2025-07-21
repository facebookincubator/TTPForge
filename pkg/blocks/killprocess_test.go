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
			name:        "Kill non-existent process with process id - throw error",
			description: "Trying to kill a process with process id that doesn't exist",
			step: &KillProcessStep{
				ProcessID:                 "123",
				ProcessName:               "ping",
				ErrorOnFindProcessFailure: true,
			},
			createProcess:      false,
			expectExecuteError: true,
		},
		{
			name:        "Kill non-existent process with process id - continue on error",
			description: "Trying to kill a process with process id that doesn't exist",
			step: &KillProcessStep{
				ProcessID:                 "123456789",
				ProcessName:               "ping",
				ErrorOnFindProcessFailure: false,
				ErrorOnKillFailure:        false,
			},
			createProcess:      false,
			expectExecuteError: false,
		},
		{
			// Test might throw an error on stress run
			name:        "Kill non-existent process with process name - throw error",
			description: "Trying to kill a process with process name that doesn't exist",
			step: &KillProcessStep{
				ProcessID:                 "",
				ProcessName:               "ping",
				ErrorOnFindProcessFailure: true,
			},
			createProcess:      false,
			expectExecuteError: true,
		},
		{
			name:        "Kill non-existent process with process name - continue on error",
			description: "Trying to kill a process with process name that doesn't exist",
			step: &KillProcessStep{
				ProcessID:   "",
				ProcessName: "ping",
			},
			createProcess:      false,
			expectExecuteError: false,
		},
		{
			name:        "Kill non-existent process with process id",
			description: "Trying to kill a process with process id that exists",
			step: &KillProcessStep{
				ProcessID:   "123456789",
				ProcessName: "ping",
			},
			createProcess:      true,
			expectExecuteError: false,
		},
		{
			name:        "Kill existent process with process name",
			description: "Trying to kill a process with process name that exists",
			step: &KillProcessStep{
				ProcessID:   "",
				ProcessName: "ping",
			},
			createProcess:      true,
			expectExecuteError: false,
		},
		{
			name:        "Kill process invalid process id",
			description: "Trying to kill a process with negative process id",
			step: &KillProcessStep{
				ProcessID:   "-100",
				ProcessName: "ping",
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process with no process id and name",
			description: "Trying to kill a process with null process id and name",
			step: &KillProcessStep{
				ProcessID:   "",
				ProcessName: "",
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process with char process id",
			description: "Trying to kill a process with character process id",
			step: &KillProcessStep{
				ProcessID:   "ABCD",
				ProcessName: "",
			},
			expectValidateError: true,
		},
		{
			name:        "Kill process ID with templating",
			description: "Trying to kill a process with templating",
			step: &KillProcessStep{
				ProcessID:                 "{[{.StepVars.pid}]}",
				ProcessName:               "ping",
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
			description: "Trying to kill a process with process name and templating",
			step: &KillProcessStep{
				ProcessID:                 "",
				ProcessName:               "{[{.StepVars.processName}]}",
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
			name:        "Kill process ID with templating error",
			description: "Trying to kill a process ID without proper templating",
			step: &KillProcessStep{
				ProcessID:   "{[{.StepVars.pid}]}",
				ProcessName: "ping",
			},
			expectTemplateError: true,
		},
		{
			name:        "Kill process name with templating error",
			description: "Trying to kill a process name without proper templating",
			step: &KillProcessStep{
				ProcessID:   "",
				ProcessName: "{[{.StepVars.processName}]}",
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
				tc.step.ProcessID = strconv.Itoa(pid)
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
