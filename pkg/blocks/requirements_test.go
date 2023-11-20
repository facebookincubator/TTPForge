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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRequirements(t *testing.T) {
	testCases := []struct {
		name                 string
		content              string
		expectValidateError  bool
		expectExecuteError   bool
		expectedRequirements *RequirementsConfig
	}{
		{
			name: "Omitted requirements section",
			content: `
name: TestTTP
description: Test description
steps:
  - name: hello
    print_str: hello world`,
			expectedRequirements: nil,
		},
		{
			name: "Valid requirements section",
			content: `
name: TestTTP
description: Test description
requirements:
  superuser: false
steps:
  - name: hello
    inline: echo "hello world"`,
			expectedRequirements: &RequirementsConfig{
				ExpectSuperuser: false,
			},
		},
		{
			name: "Invalid requirements section - cannot become root in tests",
			content: `
name: TestTTP
description: Test description
requirements:
  superuser: true
steps:
  - name: hello
    print_str: hello world`,
			expectExecuteError: true,
			expectedRequirements: &RequirementsConfig{
				ExpectSuperuser: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			require.NoError(t, err)
			var ctx TTPExecutionContext
			err = ttp.Validate(ctx)
			if tc.expectValidateError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			_, err = ttp.Execute(&ctx)
			if tc.expectExecuteError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
