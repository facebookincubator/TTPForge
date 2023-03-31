package blocks_test

import (
	"testing"
	"testing/fstest"

	"github.com/facebookincubator/TTP-Runner/blocks"
	"github.com/facebookincubator/TTP-Runner/logging"

	"gopkg.in/yaml.v3"
)

func init() {
	blocks.Logger = logging.Logger
	logging.ToggleDebug()
}

func TestUnmarshalSubTtp(t *testing.T) {
	ttps := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`
name: test
description: test ttp sub step
steps:
  name: testing_sub_ttp
  inline: |
    ls
        `),
			},
		},
	}

	content := `
name: testing
ttp: test.yaml
  `

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Error("invalid sub ttp format", ttps)
	}

	t.Logf("sub ttp step populated with data: %v", ttps)
}

func TestUnmarshalSubTtpInvalid(t *testing.T) {
	ttps := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`
name: test
description: test ttp sub step
steps:
  name: testing_sub_ttp
  inline: |
    ls
        `),
			},
		},
	}

	content := `
name: testing
ttp: bad.yaml
  `

	if err := yaml.Unmarshal([]byte(content), &ttps); err != nil {
		t.Error("unmarshalling will not check for existence quite yet, should not fail here")
	}

	if err := ttps.Validate(); err == nil {
		t.Error("failure should occur here as file does not exist")
	}
}
