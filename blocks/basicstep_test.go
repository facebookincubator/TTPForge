package blocks_test

import (
	"testing"
	"ttpforge/blocks"
	"ttpforge/utils"

	"gopkg.in/yaml.v3"
)

func init() {
	blocks.Logger = utils.Logger
	utils.ToggleDebug()
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

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
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

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
		t.Errorf("failed to unmarshal basic inline %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)

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
	err := yaml.Unmarshal([]byte(content), &ttps)
	if err == nil {
		t.Error("required parameter missing, passed unmarshal", ttps)
	}

	t.Log("successfully detected invalid format")

}
