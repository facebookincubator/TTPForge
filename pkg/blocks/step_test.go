package blocks_test

import (
	"bytes"
	"testing"

	"github.com/facebookincubator/TTP-Runner/pkg/blocks"
	"github.com/stretchr/testify/assert"
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
