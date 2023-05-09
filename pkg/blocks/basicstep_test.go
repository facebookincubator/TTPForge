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
	"strings"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/logging"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestUnmarshalBasic(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple basic",
			content: `name: test
description: this is a test
steps:
  - name: testinline
    inline: |
      ls
`,
			wantError: false,
		},
		{
			name: "Simple cleanup basic",
			content: `name: test
description: this is a test
steps:
  - name: testinline
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la
  `,
			wantError: false,
		},
		{
			name: "Invalid basic",
			content: `
name: test
description: this is a test
steps:
  - noname: testinline
    inline: |
      ls
  `,
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArgsExpansion(t *testing.T) {
	testCases := []struct {
		name  string
		basic BasicStep
		want  string
	}{
		{
			name: "Simple basic",
			basic: BasicStep{
				Inline: `ls
{{test}}
{{test_second}}`,
			},
			want: `ls
test_value
test_value_second`,
		},
		{
			name: "Simple basic",
			basic: BasicStep{
				Inline: `ls
{{test_not_expanded}}`,
			},
			want: `ls
{{test_not_expanded}}`,
		},
	}

	InputArgs["test"] = "test_value"
	InputArgs["test_second"] = "test_value_second"
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			replaced := tc.basic.replaceInput(InputArgs)
			if tc.want != strings.TrimSpace(replaced) {
				t.Errorf("expected output does not match return of replacement %v != %v", tc.want, replaced)
			}
		})
	}
}
