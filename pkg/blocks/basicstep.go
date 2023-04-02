package blocks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"gopkg.in/yaml.v3"
)

// BasicStep is a type that represents a basic execution step.
type BasicStep struct {
	*Act        `yaml:",inline"`
	Executor    string     `yaml:"executor,omitempty"`
	Inline      string     `yaml:"inline,flow"`
	Args        []string   `yaml:"args,omitempty,flow"`
	CleanupStep CleanupAct `yaml:"cleanup,omitempty"`
}

// NewBasicStep creates a new BasicStep instance with an initialized Act struct.
func NewBasicStep() *BasicStep {
	return &BasicStep{
		Act: &Act{
			Type: StepBasic,
		},
	}
}

// UnmarshalYAML custom unmarshaler for BasicStep to handle decoding from YAML.
func (b *BasicStep) UnmarshalYAML(node *yaml.Node) error {
	type BasicStepTmpl struct {
		Act         `yaml:",inline"`
		Executor    string    `yaml:"executor,omitempty"`
		Inline      string    `yaml:"inline,flow"`
		Args        []string  `yaml:"args,omitempty,flow"`
		CleanupStep yaml.Node `yaml:"cleanup,omitempty"`
	}

	var tmpl BasicStepTmpl
	// there is an issue with strict fields not being managed https://github.com/go-yaml/yaml/issues/460
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	b.Act = &tmpl.Act
	b.Args = tmpl.Args
	b.Executor = tmpl.Executor
	b.Inline = tmpl.Inline

	if b.IsNil() {
		return b.ExplainInvalid()
	}

	// we do it piecemiel to build our struct
	if tmpl.CleanupStep.IsZero() || b.Type == StepCleanup {
		return nil
	}

	logging.Logger.Sugar().Debugw("step", "name", tmpl.Name)
	cleanup, err := b.MakeCleanupStep(&tmpl.CleanupStep)
	logging.Logger.Sugar().Debugw("step", "err", err)
	if err != nil {
		return err
	}

	b.CleanupStep = cleanup

	return nil
}

// Cleanup is an implementation of the CleanupAct interface's Cleanup method.
func (b *BasicStep) Cleanup() error {
	return b.Execute()
}

// GetCleanup returns the cleanup steps for a BasicStep.
func (b *BasicStep) GetCleanup() []CleanupAct {
	if b.CleanupStep != nil {
		b.CleanupStep.SetDir(b.WorkDir)
		return []CleanupAct{b.CleanupStep}
	}
	return []CleanupAct{}
}

// CleanupName returns the name of the cleanup step.
func (b *BasicStep) CleanupName() string {
	return b.Name
}

// GetType returns the step type for a BasicStep.
func (b *BasicStep) GetType() StepType {
	return b.Type
}

// ExplainInvalid returns an error with an explanation of why a BasicStep is invalid.
func (b *BasicStep) ExplainInvalid() error {
	var err error
	if b.Inline == "" {
		err = fmt.Errorf("(inline) empty")
	}
	if b.Name != "" && err != nil {
		return fmt.Errorf("[!] invalid basicstep: [%s] %w", b.Name, err)
	}
	return err
}

// IsNil checks if a BasicStep is considered empty or uninitialized.
func (b *BasicStep) IsNil() bool {
	switch {
	case b.Act.IsNil():
		return true
	case b.Inline == "":
		return true
	default:
		return false
	}
}

// Validate validates the BasicStep, checking for the necessary attributes and dependencies.
func (b *BasicStep) Validate() error {
	err := b.Act.Validate()
	if err != nil {
		return err
	}

	if b.Inline == "" {
		return errors.New("inline must be provided")
	}

	if b.Executor == "" && b.Inline != "" {
		// TODO: add os handling using the runtime (ezpz)
		logging.Logger.Sugar().Debug("defaulting to bash since executor was not provided")
		b.Executor = "bash"
	}

	if b.Executor == ExecutorBinary {
		return nil
	}
	_, err = exec.LookPath(b.Executor)
	if err != nil {
		return err
	}
	if b.CleanupStep != nil {
		err := b.CleanupStep.Validate()
		if err != nil {
			return err
		}
	}
	logging.Logger.Sugar().Debugw("command found in path", "executor", b.Executor)

	return nil
}

func (b *BasicStep) Execute() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()
	logging.Logger.Sugar().Debugw("available data", "outputs", b.output)
	logging.Logger.Sugar().Info("========= Executing ==========")
	if b.Inline != "" {
		err = b.executeBashStdin(ctx)
	}
	logging.Logger.Sugar().Info("========= Result ==========")
	return err
}

func (b *BasicStep) executeBashStdin(ptx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ptx)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.Executor)
	cmd.Env = FetchEnv(b.Environment)
	cmd.Dir = b.WorkDir
	var inline bytes.Buffer

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"json": JSONString,
	}

	tmpl, err := template.New("inline").Funcs(funcMap).Parse(b.Inline)
	if err != nil {
		logging.Logger.Sugar().Warnw("failed to parse template", "err", err)
		return err
	}

	err = tmpl.Execute(&inline, b.output)
	if err != nil {
		logging.Logger.Sugar().Warnw("failed to execute template", "err", err)
		return err
	}

	logging.Logger.Sugar().Debugw("value of inline parsed", "inline", inline.String())
	// if b.FetchArgs() != nil {
	// 	inline = fmt.Sprintf("%s %s", b.Inline, strings.Join(b.FetchArgs(), " "))
	// }
	input := strings.NewReader(inline.String())

	logging.Logger.Sugar().Debugw("check input", "input", inline)
	cmd.Stdin = input

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	if err != nil {
		logging.Logger.Sugar().Warnw("bad exit of process", "stdout", outStr, "stderr", errStr, "exit code", cmd.ProcessState.ExitCode())
		return err
	}
	logging.Logger.Sugar().Debugw("output of process", "stdout", outStr, "stderr", errStr, "status", cmd.ProcessState.ExitCode())

	b.SetOutputSuccess(&stdoutBuf, cmd.ProcessState.ExitCode())

	return nil
}
