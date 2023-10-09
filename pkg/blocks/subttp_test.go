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

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func makeTestFsForSubTTPs(t *testing.T) afero.Fs {
	fsys, err := testutils.MakeAferoTestFs(map[string][]byte{
		"repos/a/" + repos.RepoConfigFileName: []byte(`ttp_search_paths: ["myttps"]`),
		"repos/a/myttps/awesome-sub-ttp.yaml": []byte(`name: test
description: test sub ttp basic execution
steps:
- name: testing_sub_ttp
  inline: |
    echo -n awesome-sub-ttp ran successfully`),
		"repos/a/myttps/another/args.yaml": []byte(`name: test
description: test ttp sub step
args:
- name: arg_number_one
- name: arg_number_two
- name: arg_number_three
  default: victory
steps:
- name: testing_sub_ttp
  inline: |
    echo -n {{ .Args.arg_number_one}} {{ .Args.arg_number_two}} {{ .Args.arg_number_three }}`),
		"repos/b/" + repos.RepoConfigFileName: []byte(`ttp_search_paths: ["ttps"]`),
		"repos/b/ttps/with/cleanup.yaml": []byte(`name: with-cleanup
description: test sub ttp with cleanup steps
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
	)
	require.NoError(t, err)
	return fsys
}

// func TestSubTTPExecution(t *testing.T) {

// 	tests := []struct {
// 		name                  string
// 		spec                  repos.Spec
// 		fsys                  afero.Fs
// 		stepYAML              string
// 		expectError           bool
// 		expectedOutput        string
// 		expectedCleanupOutput string
// 	}{
// 		{
// 			name: "Simple Sub TTP Execution",
// 			spec: repos.Spec{
// 				Name: "default",
// 				Path: "repos/a",
// 			},
// 			fsys: makeTestFsForSubTTPs(t),
// 			stepYAML: `name: testing
// ttp: awesome-sub-ttp.yaml`,
// 			expectedOutput: "awesome-sub-ttp ran successfully",
// 		},
// 		{
// 			name: "Sub TTP Execution with Args",
// 			spec: repos.Spec{
// 				Name: "default",
// 				Path: "repos/a",
// 			},
// 			fsys: makeTestFsForSubTTPs(t),
// 			stepYAML: `name: with-args
// ttp: another/args.yaml
// args:
//   arg_number_one: hello
//   arg_number_two: world`,
// 			expectedOutput: "hello world victory",
// 		},
// 		{
// 			name: "Sub TTP Execution with Cleanups",
// 			spec: repos.Spec{
// 				Name: "b",
// 				Path: "repos/b",
// 			},
// 			fsys: makeTestFsForSubTTPs(t),
// 			stepYAML: `name: with-cleanup
// ttp: with/cleanup.yaml`,
// 			expectedOutput:        "sub_step_1_output\nsub_step_2_output\n",
// 			expectedCleanupOutput: "cleanup_sub_step_2\ncleanup_sub_step_1\n",
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var step blocks.SubTTPStep
// 			err := yaml.Unmarshal([]byte(tc.stepYAML), &step)
// 			require.NoError(t, err, "step YAML should unmarshal safely")

// 			repo, err := tc.spec.Load(tc.fsys, "")
// 			require.NoError(t, err)

// 			execCtx := blocks.TTPExecutionContext{
// 				Cfg: blocks.TTPExecutionConfig{
// 					Repo: repo,
// 				},
// 			}
// 			err = step.Validate(execCtx)
// 			require.NoError(t, err, "step failed to validate")

// 			result, err := step.Execute(execCtx)
// 			require.NoError(t, err)
// 			assert.Equal(t, tc.expectedOutput, result.Stdout)

// 			if tc.expectedCleanupOutput != "" {
// 				cleanups := step.GetCleanup()
// 				require.NotNil(t, cleanups)
// 				cleanupResult, err := cleanups[0].Cleanup(execCtx)
// 				require.NoError(t, err)
// 				assert.Equal(t, tc.expectedCleanupOutput, cleanupResult.Stdout)
// 			}
// 		})
// 	}
// }
