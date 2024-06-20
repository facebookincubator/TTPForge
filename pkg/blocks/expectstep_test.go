package blocks

import (
	"testing"
	"time"

	expect "github.com/Netflix/go-expect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalExpect(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple expect",
			content: `name: test
description: this is a test
steps:
  - name: testexpect
    action: expect
    expect_string: "Hello World!"`,
			wantError: false,
		},
		{
			name: "Expect missing string",
			content: `name: test
description: this is a test
steps:
  - name: testexpect
    action: expect`,
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpectStepExecuteWithOutput(t *testing.T) {
	// prepare step
	content := `name: test_expect_step
action: expect
expect_string: "Hello World!"`
	var s ExpectStep
	execCtx := NewTTPExecutionContext()
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	err = s.Validate(execCtx)
	require.NoError(t, err)

	// prepare the console
	console, err := expect.NewConsole(expect.WithStdout(execCtx.Cfg.Stdout))
	require.NoError(t, err)
	defer console.Close()

	done := make(chan struct{})

	// simulate console input
	go func() {
		defer close(done)
		time.Sleep(1 * time.Second)
		console.SendLine("Hello World!")
	}()

	// execute the step and check result
	result, err := s.Execute(execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)

	<-done // wait for the goroutine to finish
}

func TestExpectStepExecuteTimeout(t *testing.T) {
	// prepare step
	content := `name: test_expect_step
action: expect
expect_string: "Hello World!"
timeout: 1`
	var s ExpectStep
	execCtx := NewTTPExecutionContext()
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)
	err = s.Validate(execCtx)
	require.NoError(t, err)

	// prepare the console
	console, err := expect.NewConsole(expect.WithStdout(execCtx.Cfg.Stdout))
	require.NoError(t, err)
	defer console.Close()

	done := make(chan struct{})

	// simulate console input
	go func() {
		defer close(done)
		time.Sleep(2 * time.Second)
		console.SendLine("Hello World!")
	}()

	// execute the step and check result
	_, err = s.Execute(execCtx)
	assert.Error(t, err)

	<-done // wait for the goroutine to finish
}
