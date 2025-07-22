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
	"fmt"
	"github.com/shirou/gopsutil/process"
)

// GetPIDsByName returns a list of process IDs that match the given process name
func GetPIDsByName(processName string) ([]int32, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}
	var pids []int32
	for _, proc := range processes {
		name, err := proc.Name()
		if err == nil && name == processName {
			pids = append(pids, proc.Pid)
		}
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("No process found with name: %s", processName)
	}
	return pids, nil
}

// VerifyPIDExists returns a boolean basis if a process with the input PID exists
func VerifyPIDExists(pid int) error {
	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, proc := range processes {
		if int(proc.Pid) == pid {
			return nil
		}
	}
	return fmt.Errorf("No process found with PID: %d", pid)
}
