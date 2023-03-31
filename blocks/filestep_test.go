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
func TestUnmarshalSimpleFile(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
description: this is a test
steps:
  - name: test_file
    file: test_file
  `

	err := yaml.Unmarshal([]byte(content), &ttps)
	if err != nil {
		t.Errorf("failed to unmarshal file step %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)

}
