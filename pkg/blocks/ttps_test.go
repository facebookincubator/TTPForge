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
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

// TestAmbiguousStepType verifies that we error
// out appropriately when an ambiguously-typed
// step is provided. Should probably live in step_test.go
// eventually but for current code structure
// it is better to have it live here.
func TestAmbiguousStepType(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "Ambiguous Inline+File Step",
			content: `name: test
description: this is a test
steps:
  - name: ambiguous
    inline: foo
    file: bar`,
		},
		{
			name: "Ambiguous Edit+SubTTP Step",
			content: `name: test
description: this is a test
steps:
  - name: ambiguous
    edit_file: hello
    ttp: world`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			assert.Error(t, err, "steps with ambiguous types should yield an error when parsed")
		})
	}
}

func TestUnmarshalSimpleCleanupLarge(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple cleanup large",
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
  - name: test_cleanup_two
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la
  - name: test_cleanup_three
    inline: |
      ls
    cleanup:
      name: test_cleanup
      inline: |
        ls -la
  `,
			wantError: false,
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

func TestUnmarshalScenario(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Hello World scenario",
			content: `
---
name: Hello World
description: |
  Print hello world
steps:
  - name: hello
    file: ./ttps/privilege-escalation/credential-theft/hello-world/hello-world.sh
    cleanup:
      name: cleanup
      inline: |
        echo "cleaned up!"
        - name: hello_inline
        inline: |
          ./ttps/privilege-escalation/credential-theft/hello-world/hello-world.sh
`,
			wantError: false,
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

func TestTTP_RunSteps(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Empty steps",
			content: `
name: test
description: this is a test
steps: []
`,
			wantError: false,
		},
		{
			name: "Valid steps with cleanup",
			content: `
name: test
description: this is a test
steps:
  - name: step1
    inline: |
      echo "step1"
    cleanup:
      name: cleanup1
      inline: |
        echo "cleanup1"
  - name: step2
    inline: |
      echo "step2"
    cleanup:
      name: cleanup2
      inline: |
        echo "cleanup2"
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			assert.NoError(t, err)

			_, err = ttp.RunSteps(blocks.TTPExecutionConfig{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTTP_ValidateSteps(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Valid steps",
			content: `
name: test
description: this is a test
steps:
  - name: step1
    inline: |
      echo "step1"
  - name: step2
    inline: |
      echo "step2"
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = ttp.ValidateSteps(blocks.TTPExecutionContext{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteStepsBasic(t *testing.T) {
	content := `name: test
description: this is a test
steps:
    - name: step1
      inline: echo "step1"
      cleanup:
        inline: echo "cleanup1"
    - name: step2
      inline: echo "step2"
      cleanup:
        inline: echo "cleanup2"`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(blocks.TTPExecutionConfig{})
	require.NoError(t, err)

	require.Equal(t, 2, len(stepResults.ByIndex))
	assert.Equal(t, "step1\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "step2\n", stepResults.ByIndex[1].Stdout)

	require.Equal(t, 2, len(stepResults.ByName))
	assert.Equal(t, "step1\n", stepResults.ByName["step1"].Stdout)
	assert.Equal(t, "step2\n", stepResults.ByName["step2"].Stdout)

	require.NotNil(t, stepResults.ByIndex[0].Cleanup)
	require.NotNil(t, stepResults.ByIndex[1].Cleanup)
	assert.Equal(t, "cleanup1\n", stepResults.ByIndex[0].Cleanup.Stdout)
	assert.Equal(t, "cleanup2\n", stepResults.ByIndex[1].Cleanup.Stdout)

	require.NotNil(t, stepResults.ByName["step1"].Cleanup)
	require.NotNil(t, stepResults.ByName["step2"].Cleanup)
	assert.Equal(t, "cleanup1\n", stepResults.ByName["step1"].Cleanup.Stdout)
	assert.Equal(t, "cleanup2\n", stepResults.ByName["step2"].Cleanup.Stdout)
}
