package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	expect "github.com/Netflix/go-expect"
)

// ExpectStep represents an expect command
type ExpectStep struct {
	actionDefaults `yaml:",inline"`
	ExpectString   string `yaml:"expect_string,omitempty"`
	SendCommands   string `yaml:"send_commands,omitempty"`
	Command        string `yaml:"command,omitempty"`
}

// NewExpectStep creates a new ExpectStep instance with an initialized Act struct.
func NewExpectStep() *ExpectStep {
	return &ExpectStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *ExpectStep) IsNil() bool {
	return s.ExpectString == "" && s.SendCommands == "" && s.Command == ""
}

// Validate validates the step, checking for the necessary attributes and dependencies.
func (s *ExpectStep) Validate(execCtx TTPExecutionContext) error {
	if s.ExpectString == "" {
		return fmt.Errorf("expect_string is not specified")
	}
	if s.SendCommands == "" {
		return fmt.Errorf("send_commands are not specified")
	}
	if s.Command == "" {
		return fmt.Errorf("command is not specified")
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *ExpectStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	stdout := execCtx.Cfg.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	console, err := expect.NewConsole(expect.WithStdout(stdout))
	if err != nil {
		return nil, fmt.Errorf("failed to create new console: %w", err)
	}
	defer console.Close()

	fmt.Printf("Console created: %+v\n", console)

	cmd := exec.Command("sh", "-c", s.Command)
	cmd.Stdin = console.Tty()
	cmd.Stdout = console.Tty()
	cmd.Stderr = console.Tty()

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	time.Sleep(time.Second)

	// Send commands
	sendCommands := strings.Split(s.SendCommands, "\n")
	for _, sendCmd := range sendCommands {
		if sendCmd == "" {
			continue
		}
		sendCmd = strings.TrimPrefix(sendCmd, "send ")
		if strings.HasPrefix(sendCmd, "send_line ") {
			sendCmd = strings.TrimPrefix(sendCmd, "send_line ")
			if _, err := console.SendLine(sendCmd); err != nil {
				return nil, fmt.Errorf("failed to send line command: %w", err)
			}
		} else {
			if _, err := console.Send(sendCmd); err != nil {
				return nil, fmt.Errorf("failed to send command: %w", err)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	done := make(chan error, 1)
	go func() {
		_, err := console.Expect(expect.String(s.ExpectString), expect.EOF)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("error in Expect: %w", err)
		}
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for Expect")
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command wait failed: %w", err)
	}

	return &ActResult{}, nil
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *ExpectStep) CanBeUsedInCompositeAction() bool {
	return true
}
