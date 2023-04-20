package cmd

import (
	"encoding/json"
	"fmt"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runE2ETest(t *testing.T, testFile string, stepOutputs []string) {
	ttp, err := ExecuteYAML("e2e-tests/" + testFile)
	assert.Nil(t, err)

	assert.Equal(t, len(stepOutputs), len(ttp.Steps), "step outputs should have correct length")

	for stepIdx, step := range ttp.Steps {
		output := step.GetOutput()
		b, err := json.Marshal(output)
		assert.Nil(t, err)
		assert.Equal(t, stepOutputs[stepIdx], string(b), "step output is incorrect")

	}

	assert.NotNil(t, ttp)
}

func TestE2E(t *testing.T) {

	userStruct, err := user.Current()
	assert.Nil(t, err)
	u := userStruct.Username

	scenarios := map[string][]string{
		"test_variable_expansion.yaml": {
			fmt.Sprintf("{\"output\":\"%v\"}", u),
			fmt.Sprintf("{\"another_key\":\"wut\",\"test_key\":\"%v\"}", u),
			"{\"output\":\"you said: wut\"}",
		},
	}
	for scenarioFile, stepOutputs := range scenarios {
		runE2ETest(t, scenarioFile, stepOutputs)
	}
}
