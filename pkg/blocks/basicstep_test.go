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
