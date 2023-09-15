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

package blocks_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

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
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBasicStepExecuteWithOutput(t *testing.T) {
	// prepare step
	content := `name: test_basic_step
inline: echo {\"foo\":{\"bar\":\"baz\"}}
outputs:
  first:
    filters:
    - json_path: foo.bar`
	var s blocks.BasicStep
	execCtx := blocks.TTPExecutionContext{
		Cfg: blocks.TTPExecutionConfig{
			Args: map[string]any{
				"myarg": "baz",
			},
		},
	}
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	err = s.Validate(execCtx)
	require.NoError(t, err)

	// execute and check result
	result, err := s.Execute(execCtx)
	require.NoError(t, err)
	require.Equal(t, 1, len(result.Outputs))
	assert.Equal(t, "baz", result.Outputs["first"], "first output should be correct")
}

func TestBasicStepUnmarshalIgnoreErrors(t *testing.T) {
	data := `
name: testStep
inline: echo "Hello World"
ignore_errors: true
`
	step := &blocks.BasicStep{}
	err := yaml.Unmarshal([]byte(data), step)
	assert.NoError(t, err)
	assert.True(t, step.IgnoreErrors)
}

func TestExecuteWithIgnoreErrors(t *testing.T) {
	step := &blocks.BasicStep{
		Act: &blocks.Act{
			Type: blocks.StepBasic,
			Name: "errorStep",
		},
		Executor:     "bash",
		Inline:       "exit 1",
		IgnoreErrors: true,
	}

	ctx := blocks.TTPExecutionContext{}

	result, err := step.Execute(ctx)
	assert.NoError(t, err) // Since IgnoreErrors is true, we shouldn't get an error.
	assert.NotNil(t, result)
}

func TestExecuteWithoutIgnoreErrors(t *testing.T) {
	step := &blocks.BasicStep{
		Act: &blocks.Act{
			Type: blocks.StepBasic,
			Name: "errorStep",
		},
		Executor:     "bash",
		Inline:       "exit 1",
		IgnoreErrors: false,
	}

	ctx := blocks.TTPExecutionContext{}

	_, err := step.Execute(ctx)
	assert.Error(t, err) // Since IgnoreErrors is false, we should get an error.
}

func TestValidExecuteWithIgnoreErrors(t *testing.T) {
	step := &blocks.BasicStep{
		Act: &blocks.Act{
			Type: blocks.StepBasic,
			Name: "validStep",
		},
		Executor:     "bash",
		Inline:       "echo Hello",
		IgnoreErrors: true,
	}

	ctx := blocks.TTPExecutionContext{}

	result, err := step.Execute(ctx)
	assert.NoError(t, err) // Even if IgnoreErrors is true, we shouldn't get an error since the step is valid.
	assert.NotNil(t, result)
	assert.Equal(t, "Hello\n", result.Stdout)
}
