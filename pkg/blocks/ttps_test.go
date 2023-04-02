package blocks_test

import (
	"testing"

	"github.com/facebookincubator/TTP-Runner/pkg/blocks"
	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestUnmarshalSimpleCleanupLarge(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
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
  `

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)

}

func TestUnmarshalScenario(t *testing.T) {

	var ttps blocks.TTP

	content := `
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
`

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("Successfully unmarshalled data: %v", ttps)
}

func TestTTP_RunSteps(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tt.content), &ttp)
			assert.NoError(t, err)

			err = ttp.RunSteps()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
