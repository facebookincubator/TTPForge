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
func TestUnmarshalSimpleFile(t *testing.T) {

	var ttps blocks.TTP

	content := `name: test
description: this is a test
steps:
  - name: test_file
    file: test_file
  `

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Errorf("failed to unmarshal file step %v", err)
	}

	t.Logf("successfully unmarshalled data: %v", ttps)
}
