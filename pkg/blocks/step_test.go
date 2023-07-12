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

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMakeCleanupStep(t *testing.T) {
	tests := []struct {
		name          string
		yamlData      string
		expectedType  string
		expectedError string
	}{
		{
			name: "BasicStep",
			yamlData: `
name: "cleanup-test"
command: "echo 'cleanup'"
inline: true
`,
			expectedType: "BasicStep",
		},
		{
			name: "FileStep",
			yamlData: `
name: "cleanup-test"
src: "source/file"
dest: "destination/file"
filepath: true
`,
			expectedType:  "FileStep",
			expectedError: "empty FilePath provided",
		},
		{
			name: "InvalidStep",
			yamlData: `
invalid_key: "invalid_value"
`,
			expectedError: "invalid parameters for cleanup steps with basic [(inline) empty], file [empty FilePath provided]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(test.yamlData), &node))

			act := &blocks.Act{}
			cleanupAct, err := act.MakeCleanupStep(&node)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
				assert.Nil(t, cleanupAct)
			} else {
				assert.NoError(t, err)

				switch test.expectedType {
				case "BasicStep":
					_, ok := cleanupAct.(*blocks.BasicStep)
					assert.True(t, ok, "Expected BasicStep")
				case "FileStep":
					_, ok := cleanupAct.(*blocks.FileStep)
					assert.True(t, ok, "Expected FileStep")
				default:
					t.Fatalf("Unknown expected type: %s", test.expectedType)
				}
			}
		})
	}
}
