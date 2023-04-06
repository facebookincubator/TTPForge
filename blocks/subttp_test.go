package blocks_test

import (
	"testing"
	"testing/fstest"
	"ttpforge/blocks"
	"ttpforge/utils"

	"gopkg.in/yaml.v3"
)

func init() {
	blocks.Logger = utils.Logger
	utils.ToggleDebug()
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

	err := yaml.Unmarshal([]byte(content), &ttps)
	t.Log(err)
	if err != nil {
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

	err := yaml.Unmarshal([]byte(content), &ttps)
	t.Log(err)
	if err != nil {
		t.Error("unmarshalling will not check for existence quite yet, should not fail here")
	}

	err = ttps.Validate()
	if err == nil {
		t.Error("failure should occur here as file does not exist")
	}

}
