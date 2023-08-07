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
		// 		{
		// 			name: "Ambiguous Edit+SubTTP Step",
		// 			content: `name: test
		// description: this is a test
		// steps:
		//   - name: ambiguous
		//     edit_file: hello
		//     ttp: world`,
		// 		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			assert.Error(t, err, "steps with ambiguous types should yield an error when parsed")
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

func TestCleanupAfterStepFailure(t *testing.T) {
	content := `name: test
description: verifies that cleanups run after step failures
steps:
    - name: step1
      inline: echo "step1"
      cleanup:
        inline: echo "cleanup1"
    - name: step2
      inline: echo "step2"
      cleanup:
        inline: echo "cleanup2"
    - name: step3
      inline: THIS WILL FAIL ON PURPOSE
      cleanup:
        inline: echo "cleanup3"
    - name: step4
      inline: echo "step4"
      cleanup:
        inline: echo "cleanup4"`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(blocks.TTPExecutionConfig{})
	assert.Error(t, err, "should get an error from step failure")

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

func TestTemplatingArgsAndConditionalExec(t *testing.T) {
	content := `name: test_variable_expansion
description: tests args + step result variable expansion functionality
args:
- name: arg1
- name: do_optional_step_1
  default: false
- name: do_optional_step_2
  default: false
steps:
  - name: mandatory_step
    inline: echo "arg value is {{ .Args.arg1 }}"
    {{ if .Args.do_optional_step_1 }}
  - name: optional_step_1
    inline: echo "optional step 1"
    {{ end }}
    {{ if .Args.do_optional_step_2 }}
  - name: optional_step_2
    inline: echo "optional step 2"
    {{ end }}`

	execCfg := blocks.TTPExecutionConfig{
		Args: map[string]string{
			"arg1":               "victory",
			"do_optional_step_1": "false",
			"do_optional_step_2": "true",
		},
	}
	ttp, err := blocks.RenderTemplatedTTP(content, &execCfg)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(execCfg)
	require.NoError(t, err)

	require.Equal(t, 2, len(stepResults.ByIndex))
	assert.Equal(t, "arg value is victory\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "optional step 2\n", stepResults.ByIndex[1].Stdout)

	require.Equal(t, 2, len(stepResults.ByName))
	assert.Equal(t, "arg value is victory\n", stepResults.ByName["mandatory_step"].Stdout)
	assert.Equal(t, "optional step 2\n", stepResults.ByName["optional_step_2"].Stdout)
}

func TestVariableExpansionArgsAndStepResults(t *testing.T) {
	content := `name: test_variable_expansion
description: tests args + step result variable expansion functionality
args:
- name: arg1
steps:
  - name: step1
    inline: echo {\"foo\":{\"bar\":\"baz\"}}
    outputs:
      first:
        filters:
        - json_path: foo.bar
  - name: step2
    inline: echo "first output is baz"
  - name: step3
    inline: echo "arg value is {{ .Args.arg1 }}"`

	execCfg := blocks.TTPExecutionConfig{
		Args: map[string]string{
			"arg1": "victory",
		},
	}
	ttp, err := blocks.RenderTemplatedTTP(content, &execCfg)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(execCfg)
	require.NoError(t, err)

	require.Equal(t, 3, len(stepResults.ByIndex))
	assert.Equal(t, "{\"foo\":{\"bar\":\"baz\"}}\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "first output is baz\n", stepResults.ByIndex[1].Stdout)
	assert.Equal(t, "arg value is victory\n", stepResults.ByIndex[2].Stdout)

	require.Equal(t, 3, len(stepResults.ByName))
	assert.Equal(t, "{\"foo\":{\"bar\":\"baz\"}}\n", stepResults.ByName["step1"].Stdout)
	assert.Equal(t, "first output is baz\n", stepResults.ByName["step2"].Stdout)
	assert.Equal(t, "arg value is victory\n", stepResults.ByName["step3"].Stdout)
}
