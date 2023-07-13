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

func TestExecuteSubTtpSearchPath(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"ttps/test.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test sub ttp in search path
steps:
  - name: sub_step_1
    inline: echo sub_step_1_output
    cleanup:
      inline: echo cleanup_sub_step_1
  - name: sub_step_2
    inline: echo sub_step_2_output
    cleanup:
      inline: echo cleanup_sub_step_2`),
			},
		},
	}

	content := `
name: testing
ttp: test.yaml`

	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err, "invalid sub ttp step format")

	execCtx := blocks.TTPExecutionContext{
		Cfg: blocks.TTPExecutionConfig{
			TTPSearchPaths: []string{"ttps"},
		},
	}
	err = step.Validate(execCtx)
	require.NoError(t, err, "TTP failed to validate")

	// execute the step
	result, err := step.Execute(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "sub_step_1_output\nsub_step_2_output\n", result.Stdout)

	// cleanup the step
	cleanups := step.GetCleanup()
	require.NotNil(t, cleanups)
	cleanupResult, err := cleanups[0].Cleanup(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "cleanup_sub_step_2\ncleanup_sub_step_1\n", cleanupResult.Stdout)
}

func TestExecuteSubTtpCurrentDir(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"anotherTest.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test sub ttp in current dir
steps:
  - name: testing_sub_ttp
    inline: |
      echo -n in_current_dir`),
			},
		},
	}

	content := `
name: testing
ttp: anotherTest.yaml`

	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err, "invalid sub ttp step format")

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err, "TTP failed to validate")

	result, err := step.Execute(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "in_current_dir", result.Stdout)
}

func TestExecuteSubTtpWithArgs(t *testing.T) {
	step := blocks.SubTTPStep{
		FileSystem: fstest.MapFS{
			"test.yaml": &fstest.MapFile{
				Data: []byte(`name: test
description: test ttp sub step
args:
  - name: arg_number_one
  - name: arg_number_two
  - name: arg_number_three
    default: victory
steps:
  - name: testing_sub_ttp
    inline: |
      echo -n {{args.arg_number_one}} {{args.arg_number_two}} {{args.arg_number_three}}`),
			},
		},
	}

	content := `name: testing
ttp: test.yaml
args:
  arg_number_one: hello
  arg_number_two: world`

	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	result, err := step.Execute(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "hello world victory", result.Stdout)
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
	var execCtx blocks.TTPExecutionContext
	if err := ttps.Validate(execCtx); err == nil {
		t.Error("failure should occur here as file does not exist")
	}
}
