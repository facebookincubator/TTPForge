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

// // TestAmbiguousStepType verifies that we error
// // out appropriately when an ambiguously-typed
// // step is provided. Should probably live in step_test.go
// // eventually but for current code structure
// // it is better to have it live here.
// func TestAmbiguousStepType(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		content string
// 	}{
// 		{
// 			name: "Ambiguous Inline+File Step",
// 			content: `name: test
// description: this is a test
// steps:
//   - name: ambiguous
//     inline: foo
//     file: bar`,
// 		},
// 		{
// 			name: "Ambiguous Edit+SubTTP Step",
// 			content: `name: test
// description: this is a test
// steps:
//   - name: ambiguous
//     edit_file: hello
//     ttp: world`,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var ttps blocks.TTP
// 			err := yaml.Unmarshal([]byte(tc.content), &ttps)
// 			assert.Error(t, err, "steps with ambiguous types should yield an error when parsed")
// 		})
// 	}
// }

func TestUnmarshalYAML(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "File scenario",
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
		{
			name: "Basic Mitre scenario",
			content: `
---
name: Leverage mdfind to search for aws credentials on disk.
description: |
  This TTP runs a search using mdfind to search for AKIA strings in files,
  which would likely indicate that the file is an aws key.
mitre:
  tactics:
    - TA0006 Credential Access
  techniques:
    - T1552 Unsecured Credentials
  subtechniques:
    - "T1552.001 Unsecured Credentials: Credentials In Files"
steps:
  - name: mdfind_aws_keys
    inline: |
      echo -e "Searching for aws keys on disk using mdfind..."
      mdfind "kMDItemTextContent == '*AKIA*' || kMDItemDisplayName == '*AKIA*' -onlyin ~"
      echo "[+] TTP Done!"
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

func TestTTP(t *testing.T) {
	testCases := []struct {
		name               string
		content            string
		execConfig         blocks.TTPExecutionConfig
		expectedByIndexOut map[int]string
		expectedByNameOut  map[string]string
		wantError          bool
	}{
		{
			name: "Cleanup After Step Failure",
			content: `name: test
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
        inline: echo "cleanup4"`,
			execConfig: blocks.TTPExecutionConfig{},
			expectedByIndexOut: map[int]string{
				0: "step1\n",
				1: "step2\n",
			},
			expectedByNameOut: map[string]string{
				"step1": "step1\n",
				"step2": "step2\n",
			},
			wantError: true,
		},
		{
			name: "Templating Args And Conditional Exec",
			content: `name: test_variable_expansion
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
{{ end }}`,
			execConfig: blocks.TTPExecutionConfig{
				Args: map[string]interface{}{
					"arg1":               "victory",
					"do_optional_step_2": true,
				},
			},
			expectedByIndexOut: map[int]string{
				0: "arg value is victory\n",
				1: "optional step 2\n",
			},
			expectedByNameOut: map[string]string{
				"mandatory_step":  "arg value is victory\n",
				"optional_step_2": "optional step 2\n",
			},
			wantError: false,
		},
		{
			name: "Variable Expansion Args And Step Results",
			content: `name: test_variable_expansion
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
    inline: echo "arg value is {{ .Args.arg1 }}"`,
			execConfig: blocks.TTPExecutionConfig{
				Args: map[string]interface{}{
					"arg1": "victory",
				},
			},
			expectedByIndexOut: map[int]string{
				0: "{\"foo\":{\"bar\":\"baz\"}}\n",
				1: "first output is baz\n",
				2: "arg value is victory\n",
			},
			expectedByNameOut: map[string]string{
				"step1": "{\"foo\":{\"bar\":\"baz\"}}\n",
				"step2": "first output is baz\n",
				"step3": "arg value is victory\n",
			},
			wantError: false,
		},
		{
			name: "Metacharacters in step contents",
			content: `name: test_metacharacters
steps:
  - name: step1
    inline: |
      cat <<EOF
      A
      B
      EOF`,
			execConfig: blocks.TTPExecutionConfig{},
			expectedByIndexOut: map[int]string{
				0: "A\nB\n",
			},
			expectedByNameOut: map[string]string{
				"step1": "A\nB\n",
			},
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Render the templated TTP first
			ttp, err := blocks.RenderTemplatedTTP(tc.content, &tc.execConfig)
			if err != nil {
				t.Fatalf("failed to render and unmarshal templated TTP: %v", err)
				return
			}

			stepResults, err := ttp.Execute(tc.execConfig)
			if tc.wantError && err == nil {
				t.Error("expected an error from step execution but got none")
				return
			}
			if !tc.wantError && err != nil {
				t.Errorf("didn't expect an error from step execution but got: %s", err)
				return
			}

			for index, output := range tc.expectedByIndexOut {
				require.Equal(t, output, stepResults.ByIndex[index].Stdout)
			}
			for name, output := range tc.expectedByNameOut {
				require.Equal(t, output, stepResults.ByName[name].Stdout)
			}
		})
	}
}

// func TestMitreAttackMapping(t *testing.T) {
// 	testCases := []struct {
// 		name      string
// 		content   string
// 		wantError bool
// 	}{
// 		{
// 			name: "Valid MITRE Mapping",
// 			content: `
// name: TestTTP
// description: Test description
// mitre:
//   tactics:
//     - Initial Access
//     - Execution
//   techniques:
//     - Spearphishing Link
//   subtechniques:
//     - Attachment
// `,
// 			wantError: false,
// 		},
// 		{
// 			name: "Invalid MITRE Mapping - Missing Tactic",
// 			content: `
// name: TestTTP
// description: Test description
// mitre:
//   techniques:
//     - Spearphishing Link
// `,
// 			wantError: true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var ttp blocks.TTP
// 			err := yaml.Unmarshal([]byte(tc.content), &ttp)
// 			if tc.wantError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, ttp.MitreAttackMapping.Tactics)
// 			}
// 		})
// 	}
// }
