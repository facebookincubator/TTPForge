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

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestMarshalYAML(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedOutput blocks.TTP
		expectErr      bool
	}{
		{
			name: "Successful Marshal",
			content: `---
name: paramtest
description: Test variadiac parameter handling
steps:
  - name: "paramtest"
    inline: |
      set -e

      user="$(echo {{user}} | tr -d '\n\t\r')"
      if [[ "{{user}}" == *'{{'* ]]; then
          user=""
      fi

      password="$(echo {{password}} | tr -d '\n\t\r')"
      if [[ "{{password}}" == *'{{'* ]]; then
          password=""
      fi

      if [[ (-z "$user") || (-z "$password") ]]; then
          echo "Error: Both user and password must have a value."
          exit 1
      fi

      go run variadicParameterExample.go \
        --user $user \
        --password $password`,
			expectedOutput: blocks.TTP{
				Name: "paramtest",
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Unmarshal the YAML string to a TTP instance
			ttp := blocks.TTP{}
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if err != nil {
				t.Errorf("Failed to unmarshal YAML: %v", err)
				return
			}

			// Test the MarshalYAML function
			outputIface, err := ttp.MarshalYAML()
			if (err != nil) != tc.expectErr {
				t.Errorf("MarshalYAML() error = %v, expectErr %v", err, tc.expectErr)
				return
			}

			// Assert output to string
			output, ok := outputIface.(string)
			if !ok {
				t.Errorf("Failed to assert output to string")
				return
			}

			// Unmarshal the output back to a struct
			outStruct := blocks.TTP{}
			err = yaml.Unmarshal([]byte(output), &outStruct)
			if err != nil {
				t.Errorf("Failed to unmarshal output: %v", err)
				return
			}

			// Compare the relevant fields
			if outStruct.Name != tc.expectedOutput.Name {
				t.Errorf("MarshalYAML() got = %v, want %v", outStruct.Name, tc.expectedOutput.Name)
			}
		})
	}
}

func TestUnmarshalSimpleCleanupLarge(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		expectErr bool
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
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.expectErr {
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
		expectErr bool
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
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.expectErr {
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
		expectErr bool
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
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = ttp.ValidateSteps(blocks.TTPExecutionContext{})
			if tc.expectErr {
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
		expectErr bool
	}{
		{
			name: "Empty steps",
			content: `
name: test
description: this is a test
steps: []
`,
			expectErr: false,
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
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			assert.NoError(t, err)

			err = ttp.RunSteps(blocks.TTPExecutionContext{})
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTTP_Cleanup(t *testing.T) {
	testCases := []struct {
		name           string
		execCtx        blocks.TTPExecutionContext
		availableSteps map[string]blocks.Step
		cleanupSteps   []blocks.CleanupAct
		expectErr      bool
	}{
		{
			name:           "No cleanup steps",
			execCtx:        blocks.TTPExecutionContext{},
			availableSteps: map[string]blocks.Step{},
			cleanupSteps:   []blocks.CleanupAct{},
			expectErr:      false,
		},
		{
			name:           "One cleanup step",
			execCtx:        blocks.TTPExecutionContext{},
			availableSteps: map[string]blocks.Step{},
			cleanupSteps:   []blocks.CleanupAct{},
			expectErr:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// act := &blocks.Act{}
			// cleanupAct, err := act.MakeCleanupStep(&node)

			ttp := blocks.TTP{}
			err := ttp.Cleanup(tc.execCtx, tc.availableSteps, tc.cleanupSteps)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
