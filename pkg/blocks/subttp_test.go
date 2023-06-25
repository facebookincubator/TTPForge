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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func init() {
	logging.ToggleDebug()
}

func TestExecuteSubTtp(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			// "ttps/test.yaml": &fstest.MapFile{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test ttp sub step
steps:
  - name: testing_sub_ttp
    inline: |
      echo victory`),
			},
		},
	}

	content := `
name: testing
ttp: test.yaml`

	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err, "invalid sub ttp step format")

	err = step.Validate()
	require.NoError(t, err, "TTP failed to validate")

	// TODO: remove Setup() call after upcoming ExecutionContext refactor
	step.Setup(nil, nil)
	if err := step.Execute(map[string]string{}); err != nil {
		t.Error("TTP failed to execute", err)
	}

	// TODO: clean this up after output handling refactor
	stepOutput := step.GetOutput()
	subStepOutputMap := stepOutput["testing_sub_ttp"].(map[string]interface{})
	subStepOutput := subStepOutputMap["output"].(string)

	assert.Equal(t, "victory", subStepOutput)
}

func TestExecuteSubTtpWithArgs(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test ttp sub step
steps:
  - name: testing_sub_ttp
    inline: |
      echo {{arg_number_one}} {{arg_number_two}}`),
			},
		},
	}

	content := `name: testing
ttp: test.yaml
args:
  arg_number_one: hello
  arg_number_two: world`

	if err := yaml.Unmarshal([]byte(content), &step); err != nil {
		t.Error("invalid sub ttp step format", step)
	}

	if err := step.Validate(); err != nil {
		t.Error("TTP failed to validate", err)
	}

	// TODO: remove Setup() call after upcoming ExecutionContext refactor
	step.Setup(nil, nil)
	if err := step.Execute(map[string]string{}); err != nil {
		t.Error("TTP failed to execute", err)
	}

	// TODO: clean this up after output handling refactor
	stepOutput := step.GetOutput()
	subStepOutputMap := stepOutput["testing_sub_ttp"].(map[string]interface{})
	subStepOutput := subStepOutputMap["output"].(string)

	assert.Equal(t, "hello world", subStepOutput)
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
