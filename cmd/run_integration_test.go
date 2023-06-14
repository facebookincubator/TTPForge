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

package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRunCommandVariadicArgs(t *testing.T) {
	testCases := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{"No Variadic Args", "run ttps/examples/variadic-params/delta.yaml", true},
		{"One Variadic Arg", "run ttps/examples/variadic-params/delta.yaml " +
			"--arg username=bob", false},
		{"Many Variadic Args", "run ttps/examples/variadic-params/delta.yaml " +
			"--arg username=bob --arg pass=12345pass", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := ".."
			args := append([]string{"run", "main.go"}, strings.Fields(tc.args)...)
			cmd := exec.Command("go", args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()

			if err != nil {
				if !tc.wantErr {
					t.Errorf("unexpected error status: got %v, want %v. Output: %s", err, tc.wantErr, string(output))
				}
			}

			if err == nil && !strings.Contains(string(output), ":") {
				t.Errorf("expected output not found in command output: %s", string(output))
			}
		})
	}
}
