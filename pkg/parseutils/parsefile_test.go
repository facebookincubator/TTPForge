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

package parseutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTTP(t *testing.T) {
	tests := []struct {
		name        string
		ttpStr      string
		expectError bool
	}{
		{
			name: "valid ttp",
			ttpStr: `name: valid ttp
description: should pass parsing
args:
- name: arg1
steps:
- name: step1
  inline: echo "arg value is {{ .Args.arg1 }}"
- name: step2
  inline: echo "step two"`,
			expectError: false,
		},
		{
			name: "duplicate args key",
			ttpStr: `name: duplicate args key
description: should fail parsing
args:
- name: step1
  default: echo "step one"
args:
- name: step2
  default: echo "step two"`,
			expectError: true,
		},
		{
			name: "No Steps Field",
			ttpStr: `description: should fail linting due to args after steps
args:
- name: arg1"
`,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTTP([]byte(tc.ttpStr), tc.name+".yaml")
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
