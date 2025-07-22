//go:build !windows
// +build !windows

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

package processutils

import (
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPidsByName(t *testing.T) {

	testCases := []struct {
		name                  string
		processName           string
		processExists         bool
		expectEntriesInResult bool
	}{
		{
			name:                  "Process Exists",
			processName:           "ping",
			processExists:         true,
			expectEntriesInResult: true,
		},
		{
			name:                  "Process Does Not Exist",
			processName:           "pingabc",
			processExists:         false,
			expectEntriesInResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Setup
			var pid int
			var error error
			if tc.processExists {
				pid, error = testutils.CreateProcessToTerminate()
				require.NoError(t, error)
			}
			// Testing
			result, err := GetPIDsByName(tc.processName)
			if len(result) > 0 {
				assert.Equal(t, tc.expectEntriesInResult, true)

				foundOurPid := false
				// Cleanup
				for _, p := range result {
					if int(p) == pid {
						foundOurPid = true
						process, err := os.FindProcess(int(p))
						if err != nil {
							continue
						}
						process.Kill()
						// Need to wait for the process to be reaped by the OS (Completely killed)
						_, err = process.Wait()
						require.NoError(t, err)
					}
				}
				assert.Equal(t, foundOurPid, true)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.expectEntriesInResult, false)
			}

		})
	}
}

// Only PID=1 is always running in Unix & Linux but not windows
func TestVerifyPIDExists(t *testing.T) {

	testCases := []struct {
		name        string
		pid         int
		expectError bool
	}{
		{
			name:        "Process Exists",
			pid:         1,
			expectError: false,
		},
		{
			name:        "Process Does Not Exist",
			pid:         124567890987654321,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Testing
			err := VerifyPIDExists(tc.pid)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
