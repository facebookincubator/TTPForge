package blocks_test

import (
	"testing"

	"github.com/facebookincubator/TTP-Runner/blocks"
	"github.com/facebookincubator/TTP-Runner/logging"

	"gopkg.in/yaml.v3"
)

func init() {
	blocks.Logger = logging.Logger
	logging.ToggleDebug()
}

func TestUnmarshalSimpleBasic(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
description: this is a test
steps:
  - name: testinline
    inline: |
      ls
`

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Log("successfully unmarshalled data")

}

func TestUnmarshalSimpleCleanupBasic(t *testing.T) {

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

  `

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("Successfully unmarshalled data: %v", ttps)

}

func TestUnmarshalInvalidBasic(t *testing.T) {
	var ttps blocks.TTP

	content := `
name: test
description: this is a test
steps:
  - noname: testinline
    inline: |
      ls
  `
	if err := yaml.Unmarshal([]byte(content), &ttps); err == nil {
		t.Error("required parameter missing, passed unmarshal", ttps)
	}

	t.Log("successfully detected invalid format")

}
