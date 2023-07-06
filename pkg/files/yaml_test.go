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

package files_test

// func TestExecuteYAML(t *testing.T) {
// 	home, err := os.UserHomeDir()
// 	assert.Nil(t, err)

// 	testVariableExpansionYAML := `---
// name: Test Variable Expansion
// description: |
//   Tests environment + step output variable expansion
// steps:
//   - name: test_env_inline
//     inline: |
//       echo $HOME
//     cleanup:
//       inline: |
//         echo "cleaning up now"
//   - name: test_step_output_as_input_env
//     inline: |
//       echo "{\"test_key\" : \"$input\", \"another_key\":\"wut\"}"
//     env:
//       input: steps.test_env_inline.output
//   - name: test_step_output_in_arg_list
//     file: test-variable-expansion.sh
//     args:
//       - steps.test_step_output_as_input_env.another_key
// `

// 	testVariableExpansionSH := `#!/bin/bash

// echo "you said: $1"
// `

// 	tests := []struct {
// 		name        string
// 		testFile    string
// 		stepOutputs []string
// 	}{
// 		{
// 			name:     "test_variable_expansion",
// 			testFile: "test_variable_expansion.yaml",
// 			stepOutputs: []string{
// 				fmt.Sprintf("{\"output\":\"%v\"}", home),
// 				fmt.Sprintf("{\"another_key\":\"wut\",\"test_key\":\"%v\"}", home),
// 				"{\"output\":\"you said: wut\"}",
// 			},
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			testDir, err := os.MkdirTemp("", "e2e-tests")
// 			assert.NoError(t, err, "failed to create temporary directory")
// 			// Clean up the temporary directory
// 			defer os.RemoveAll(testDir)

// 			// Navigate to test dir
// 			if err := os.Chdir(testDir); err != nil {
// 				t.Fatalf("failed to change into test directory: %v", err)
// 			}

// 			createTestInventory(t, testDir)
// 			inventoryPath := filepath.Join(testDir, "ttps")
// 			if err := os.MkdirAll(inventoryPath, 0755); err != nil {
// 				t.Fatalf("failed to create inventoryPath (%s): %v", inventoryPath, err)
// 			}

// 			// config for the test
// 			testConfigYAML := `---
// logfile: ""
// nocolor: false
// stacktrace: false
// verbose: false
// `

// 			// Write the config to a temporary file
// 			testConfigYAMLPath := filepath.Join(testDir, "config.yaml")
// 			err = os.WriteFile(testConfigYAMLPath, []byte(testConfigYAML), 0644)
// 			assert.NoError(t, err, "failed to write the temporary YAML file")

// 			relTestYAMLPath := filepath.Join("ttps", tc.testFile)
// 			err = os.WriteFile(relTestYAMLPath, []byte(testVariableExpansionYAML), 0644)
// 			assert.NoError(t, err, "failed to write the temporary YAML file")

// 			relScriptPath := filepath.Join("ttps", "test-variable-expansion.sh")
// 			err = os.WriteFile(relScriptPath, []byte(testVariableExpansionSH), 0755)
// 			assert.NoError(t, err, "failed to write the temporary shell script")

// 			ttp, err := files.ExecuteYAML(relTestYAMLPath, blocks.TTPExecutionConfig{})
// 			assert.NoError(t, err, "execution of the testFile should not cause an error")
// 			assert.Equal(t, len(tc.stepOutputs), len(ttp.Steps), "step outputs should have correct length")

// 			for stepIdx, step := range ttp.Steps {
// 				output := step.GetOutput()
// 				b, err := json.Marshal(output)
// 				assert.Nil(t, err)
// 				assert.Equal(t, tc.stepOutputs[stepIdx], string(b), "step output is incorrect")
// 			}

// 			assert.NotNil(t, ttp)
// 		})
// 	}
// }
