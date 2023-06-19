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
	"testing"
	"testing/fstest"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestUnmarshalAndValidateSubTtp(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test ttp sub step
steps:
  - name: testing_sub_ttp
    inline: |
      ls`),
			},
		},
	}

	content := `
name: testing
ttp: test.yaml`

	if err := yaml.Unmarshal([]byte(content), &step); err != nil {
		t.Error("invalid sub ttp step format", step)
	}

	if err := step.Validate(); err != nil {
		t.Error("TTP failed to validate", err)
	}
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
