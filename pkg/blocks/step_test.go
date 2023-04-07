package blocks_test

import (
	"bytes"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSetOutputSuccess(t *testing.T) {
	testCases := []struct {
		name            string
		output          string
		exit            int
		expectedSuccess bool
		expectedOutput  map[string]interface{}
	}{
		{
			name:            "valid JSON output, exit 0",
			output:          `{"key": "value", "num": 42}`,
			exit:            0,
			expectedSuccess: true,
			expectedOutput: map[string]interface{}{
				"key": "value",
				"num": float64(42),
			},
		},
		{
			name:            "valid JSON output, exit non-zero",
			output:          `{"key": "value", "num": 42}`,
			exit:            1,
			expectedSuccess: false,
			expectedOutput: map[string]interface{}{
				"key": "value",
				"num": float64(42),
			},
		},
		{
			name:            "invalid JSON output, exit 0",
			output:          "Hello, world!",
			exit:            0,
			expectedSuccess: true,
			expectedOutput: map[string]interface{}{
				"output": "Hello, world!",
			},
		},
		{
			name:            "invalid JSON output, exit non-zero",
			output:          "Hello, world!",
			exit:            1,
			expectedSuccess: false,
			expectedOutput: map[string]interface{}{
				"output": "Hello, world!",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			act := blocks.NewAct() // Assuming there is a constructor for Act

			outputBuffer := bytes.NewBufferString(tc.output)
			act.SetOutputSuccess(outputBuffer, tc.exit)

			assert.Equal(t, tc.expectedSuccess, act.Success())
			assert.Equal(t, tc.expectedOutput, act.GetOutput())
		})
	}
}

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
