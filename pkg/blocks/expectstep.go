package blocks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	expect "github.com/l50/go-expect"
	"go.uber.org/zap"
)

// ExpectStep represents an expect command
type ExpectStep struct {
	Type           string `yaml:"type"`
	actionDefaults `yaml:",inline"`
	Chdir          string     `yaml:"chdir,omitempty"`
	Responses      []Response `yaml:"responses,omitempty"`
	Timeout        int        `yaml:"timeout,omitempty"`
	Executor       string     `yaml:"executor,omitempty"`
	Inline         string     `yaml:"inline"`
	CleanupStep    string     `yaml:"cleanup,omitempty"`
}

type Response struct {
	Prompt   string `yaml:"prompt"`
	Response string `yaml:"response"`
}

// NewExpectStep creates a new ExpectStep instance
func NewExpectStep() *ExpectStep {
	return &ExpectStep{Type: "expect"}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *ExpectStep) IsNil() bool {
	return s.Inline == "" && len(s.Responses) == 0
}

// Validate validates the step, checking for the necessary attributes and dependencies.
func (s *ExpectStep) Validate(execCtx TTPExecutionContext) error {
	// Check if Inline is provided
	if s.Inline == "" {
		err := errors.New("inline must be provided")
		logging.L().Error(zap.Error(err))
		return err
	}

	// Set Executor to "bash" if it is not provided
	if s.Executor == "" && s.Inline != "" {
		logging.L().Debug("defaulting to bash since executor was not provided")
		s.Executor = ExecutorBash
	}

	// Return if Executor is ExecutorBinary
	if s.Executor == ExecutorBinary {
		return nil
	}

	// Check if the executor is in the system path
	if _, err := exec.LookPath(s.Executor); err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	logging.L().Debugw("command found in path", "executor", s.Executor)

	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *ExpectStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	if s == nil {
		return nil, fmt.Errorf("ExpectStep is nil")
	}

	if s.Chdir != "" {
		if err := os.Chdir(s.Chdir); err != nil {
			return nil, fmt.Errorf("failed to change directory: %w", err)
		}
	}

	console, err := expect.NewConsole(expect.WithStdout(os.Stdout), expect.WithStdin(os.Stdin))
	if err != nil {
		return nil, fmt.Errorf("failed to create new console: %w", err)
	}
	defer console.Close()

	envAsList := os.Environ()
	cmd := s.prepareCommand(context.Background(), execCtx, envAsList, s.Inline)
	cmd.Stdin = console.Tty()
	cmd.Stdout = console.Tty()
	cmd.Stderr = console.Tty()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		for _, response := range s.Responses {
			re := regexp.MustCompile(response.Prompt)
			if _, err := console.Expect(expect.Regexp(re)); err != nil {
				done <- fmt.Errorf("failed to expect %q: %w", re, err)
				return
			}
			if _, err := console.SendLine(response.Response); err != nil {
				done <- fmt.Errorf("failed to send response: %w", err)
				return
			}
		}
		// Close the console to send EOF
		console.Tty().Close()
		done <- nil
	}()

	timeout := 30 * time.Second
	if s.Timeout != 0 {
		timeout = time.Duration(s.Timeout) * time.Second
	}

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("error in Expect: %w", err)
		}
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for Expect")
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("command wait failed: %w", err)
	}

	if _, err := console.ExpectEOF(); err != nil {
		return nil, fmt.Errorf("failed to expect EOF: %w", err)
	}

	return &ActResult{}, nil
}

func (s *ExpectStep) prepareCommand(ctx context.Context, execCtx TTPExecutionContext, envAsList []string, inline string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, s.Executor, "-c", inline)
	cmd.Env = envAsList
	cmd.Dir = execCtx.WorkDir
	cmd.Stdin = strings.NewReader(inline) // Use inline as the command input
	return cmd
}

// Cleanup runs the cleanup command if specified
func (s *ExpectStep) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	if s.CleanupStep == "" {
		return &ActResult{}, nil
	}

	envAsList := os.Environ()
	cmd := s.prepareCommand(context.Background(), execCtx, envAsList, s.CleanupStep)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run cleanup command: %w", err)
	}

	return &ActResult{}, nil
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *ExpectStep) CanBeUsedInCompositeAction() bool {
	return true
}
