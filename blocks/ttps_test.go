package blocks_test

import (
	"testing"

	"github.com/facebookincubator/TTP-Runner/blocks"
	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
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

			err = ttp.RunSteps()
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

			err = ttp.ValidateSteps()
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
