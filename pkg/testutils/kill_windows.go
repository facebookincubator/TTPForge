//go:build windows
// +build windows

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

package testutils

import (
	"fmt"
	"os/exec"
)

// CreateProcessToTerminate creates a process that can be killed for testing on Windows systems
func CreateProcessToTerminate() (int, error) {
	cmd := exec.Command("cmd.exe", "/C", "ping -n 60 127.0.0.1 > nul")
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to start process:", err)
		return 0, fmt.Errorf("Failed to start process: %w", err)
	}
	pid := cmd.Process.Pid
	return pid, nil
}
