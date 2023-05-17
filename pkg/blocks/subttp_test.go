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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"

	"gopkg.in/yaml.v3"
)

func init() {
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
	var ttps blocks.SubTTPStep
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

func TestSubTTPStep(t *testing.T) {
	// Prepare a temporary TTP file for the SubTTPStep to reference
	subTTPData := `
---
name: sub_ttp
description: "SubTTP for testing SubTTPStep execution."
steps:
  - name: sub_ttp_step
    file: test.sh
    args:
      - sub_ttp_arg
`
	tmpSubTTPFile, err := os.CreateTemp(os.TempDir(), "sub_ttp_*.ttp")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpSubTTPFile.Name())

	if err := os.WriteFile(tmpSubTTPFile.Name(), []byte(subTTPData), 0644); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpSubTTPFile.Name(), err)
	}

	// Create a temporary test.sh file
	testShContent := `#!/bin/bash

echo testing "$@"
`
	tmpTestShFile, err := os.CreateTemp("", "test_*.sh")
	if err != nil {
		t.Fatalf("error creating temp test.sh file: %v", err)
	}
	defer os.Remove(tmpTestShFile.Name())

	if err := os.WriteFile(tmpTestShFile.Name(), []byte(testShContent), 0755); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpTestShFile.Name(), err)
	}

	// Update InventoryPath to include the directory containing the test.sh file
	origInventoryPath := blocks.InventoryPath
	testDir := filepath.Dir(tmpTestShFile.Name())
	blocks.InventoryPath = append(blocks.InventoryPath, testDir)
	defer func() {
		blocks.InventoryPath = origInventoryPath
	}()

	t.Run("Test SubTTPStep unmarshalling and validation", func(t *testing.T) {
		// Create a temporary test.sh file
		testShContent := `#!/bin/bash

echo testing "$@"
`
		tmpTestShFile, err := os.CreateTemp("/tmp", "test_*.sh")
		if err != nil {
			t.Fatalf("error creating temp test.sh file: %v", err)
		}
		defer os.Remove(tmpTestShFile.Name())

		if err := os.WriteFile(tmpTestShFile.Name(), []byte(testShContent), 0755); err != nil {
			t.Fatalf("failed to write to %s: %v", tmpTestShFile.Name(), err)
		}
		yamlContent := fmt.Sprintf(`
---
name: test_subttpstep
description: "Test unmarshalling and validation of SubTTPStep."
steps:
  - name: test_sub_ttp
    file: %s
    ttp: %s
`, tmpTestShFile.Name(), tmpSubTTPFile.Name())

		var parsedTTP blocks.TTP
		if err := yaml.Unmarshal([]byte(yamlContent), &parsedTTP); err != nil {
			t.Fatalf("error unmarshalling yaml: %v", err)
		}

		subTTPStep := parsedTTP.Steps[0]

		if err := subTTPStep.Validate(); err != nil {
			t.Errorf("validation failed: %v", err)
		}
	})

	t.Run("Test SubTTPStep with invalid TTP file path", func(t *testing.T) {
		yamlContent := `
---
name: test_subttpstep_invalid_path
description: "Test SubTTPStep with an invalid TTP file path."
steps:
  - name: test_sub_ttp_invalid
    ttp: non_existent_file.ttp
`
		var parsedTTP blocks.TTP
		if err := yaml.Unmarshal([]byte(yamlContent), &parsedTTP); err != nil {
			t.Fatalf("error unmarshalling yaml: %v", err)
		}

		subTTPStep := parsedTTP.Steps[0]

		if err := subTTPStep.Validate(); err == nil {
			t.Error("expected validation to fail due to invalid TTP file path")
		}
	})

	t.Run("Test SubTTPStep with empty TtpFile", func(t *testing.T) {
		subTTPStep := &blocks.SubTTPStep{
			Act: &blocks.Act{
				Name: "test_sub_ttp_empty",
			},
		}

		err := subTTPStep.Validate()
		if err == nil {
			t.Error("expected validation to fail due to empty TtpFile")
		}
	})
}
