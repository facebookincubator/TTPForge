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

package blocks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// FileStep represents a step in a process that consists of a main action,
// a cleanup action, and additional metadata.
type FileStep struct {
	*Act        `yaml:",inline"`
	FilePath    string     `yaml:"file,omitempty"`
	Executor    string     `yaml:"executor,omitempty"`
	CleanupStep CleanupAct `yaml:"cleanup,omitempty,flow"`
	Args        []string   `yaml:"args,omitempty,flow"`
}

// NewFileStep creates a new FileStep instance and returns a pointer to it.
func NewFileStep() *FileStep {
	return &FileStep{
		Act: &Act{
			Type: StepFile,
		},
	}
}

// UnmarshalYAML decodes a YAML node into a FileStep instance. It uses
// the provided struct as a template for the YAML data, and initializes the
// FileStep instance with the decoded values.
//
// **Parameters:**
//
// node: A pointer to a yaml.Node representing the YAML data to decode.
//
// **Returns:**
//
// error: An error if there is a problem decoding the YAML data.
func (f *FileStep) UnmarshalYAML(node *yaml.Node) error {

	type fileStepTmpl struct {
		Act         `yaml:",inline"`
		FilePath    string    `yaml:"file,omitempty"`
		Executor    string    `yaml:"executor,omitempty"`
		CleanupStep yaml.Node `yaml:"cleanup,omitempty,flow"`
		Args        []string  `yaml:"args,omitempty,flow"`
	}

	// Decode the YAML node into the provided template.
	var tmpl fileStepTmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	// Initialize the FileStep instance with the decoded values.
	f.Act = &tmpl.Act
	f.Args = tmpl.Args
	f.FilePath = tmpl.FilePath
	f.Executor = tmpl.Executor

	// Check for invalid steps.
	if f.IsNil() {
		return f.ExplainInvalid()
	}

	// If there is no cleanup step or if this step is the cleanup step, exit.
	if tmpl.CleanupStep.IsZero() || f.Type == StepCleanup {
		return nil
	}

	// Create a CleanupStep instance and add it to the FileStep instance.
	logging.Logger.Sugar().Debugw("step", "name", tmpl.Name)
	cleanup, err := f.MakeCleanupStep(&tmpl.CleanupStep)
	logging.Logger.Sugar().Debugw("step", zap.Error(err))
	if err != nil {
		logging.Logger.Sugar().Errorw("error creating cleanup step", zap.Error(err))
		return err
	}

	f.CleanupStep = cleanup

	return nil
}

// GetType returns the type of the step as StepType.
func (f *FileStep) GetType() StepType {
	return StepFile
}

// Cleanup is a method to establish a link with the Cleanup interface.
// Assumes that the type is the cleanup step and is invoked by
// f.CleanupStep.Cleanup.
func (f *FileStep) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	result, err := f.Execute(execCtx)
	if err != nil {
		return nil, err
	}
	return &result.ActResult, err
}

// GetCleanup returns a slice of CleanupAct if the CleanupStep is not nil.
func (f *FileStep) GetCleanup() []CleanupAct {
	if f.CleanupStep != nil {
		return []CleanupAct{f.CleanupStep}
	}
	return []CleanupAct{}
}

// ExplainInvalid returns an error message explaining why the FileStep
// is invalid.
//
// **Returns:**
//
// error: An error message explaining why the FileStep is invalid.
func (f *FileStep) ExplainInvalid() error {
	var err error
	if f.FilePath == "" {
		err = errors.New("empty FilePath provided")
	}

	if f.Name != "" && err != nil {
		err = fmt.Errorf("[!] invalid FileStep: [%s] %v", f.Name, zap.Error(err))
	}

	return err
}

// IsNil checks if the FileStep is nil or empty and returns a boolean value.
func (f *FileStep) IsNil() bool {
	switch {
	case f.Act.IsNil():
		return true
	case f.FilePath == "":
		return true
	default:
		return false
	}
}

// Execute runs the FileStep and returns an error if any occur.
func (f *FileStep) Execute(execCtx TTPExecutionContext) (*ExecutionResult, error) {
	logging.Logger.Sugar().Info("========= Executing ==========")

	if f.FilePath != "" {
		if err := f.fileExec(execCtx); err != nil {
			logging.Logger.Sugar().Error(zap.Error(err))
			return nil, err
		}
	}

	logging.Logger.Sugar().Info("========= Result ==========")

	return &ExecutionResult{}, nil
}

// fileExec executes the FileStep with the specified executor and arguments,
// and returns an error if any occur.
func (f *FileStep) fileExec(execCtx TTPExecutionContext) error {
	var cmd *exec.Cmd
	expandedArgs, err := execCtx.ExpandVariables(f.Args)
	if err != nil {
		return err
	}
	if f.Executor == ExecutorBinary {
		cmd = exec.Command(f.FilePath, expandedArgs...)
	} else {
		args := []string{f.FilePath}
		args = append(args, expandedArgs...)

		logging.Logger.Sugar().Debugw("command line execution:", "exec", f.Executor, "args", args)
		cmd = exec.Command(f.Executor, args...)
	}
	envAsList := FetchEnv(f.Environment)
	expandedEnvAsList, err := execCtx.ExpandVariables(envAsList)
	if err != nil {
		return err
	}
	cmd.Env = expandedEnvAsList
	cmd.Dir = f.WorkDir
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	if err != nil {
		logging.Logger.Sugar().Errorw("bad exit of process", "stdout", outStr, "stderr", errStr, "exit code", cmd.ProcessState.ExitCode())
		return err
	}
	logging.Logger.Sugar().Debugw("output of process", "stdout", outStr, "stderr", errStr, "status", cmd.ProcessState.ExitCode())

	return nil
}

// Validate validates the FileStep. It checks that the
// Act field is valid, and that either FilePath is set with
// a valid file path, or InlineLogic is set with valid code.
//
// If FilePath is set, it ensures that the file exists and retrieves
// its absolute path.
//
// If Executor is not set, it infers the executor based on the file extension.
// It then checks that the executor is in the system path, and if CleanupStep
// is not nil, it validates the cleanup step as well.
// It logs any errors and returns them.
//
// **Returns:**
//
// error: An error if any validation checks fail.
func (f *FileStep) Validate(execCtx TTPExecutionContext) error {
	if err := f.Act.Validate(); err != nil {
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	if f.FilePath == "" {
		err := errors.New("a TTP must include inline logic or path to a file with the logic")
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	// If FilePath is set, ensure that the file exists.
	fullpath, err := FindFilePath(f.FilePath, f.WorkDir, nil)
	if err != nil {
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	// Retrieve the absolute path to the file.
	f.FilePath, err = FetchAbs(fullpath, f.WorkDir)
	if err != nil {
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	// Infer executor if it's not set.
	if f.Executor == "" {
		f.Executor = InferExecutor(f.FilePath)
		logging.Logger.Sugar().Infow("executor set via extension", "exec", f.Executor)
	}

	if f.Executor == ExecutorBinary {
		return nil
	}

	if _, err := exec.LookPath(f.Executor); err != nil {
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	if f.CleanupStep != nil {
		if err := f.CleanupStep.Validate(execCtx); err != nil {
			logging.Logger.Sugar().Errorw("error validating cleanup step", zap.Error(err))
			return err
		}
	}
	logging.Logger.Sugar().Debugw("command found in path", "executor", f.Executor)

	return nil
}

// InferExecutor infers the executor based on the file extension and
// returns it as a string.
func InferExecutor(filePath string) string {
	ext := filepath.Ext(filePath)
	logging.Logger.Sugar().Debugw("file extension inferred", "filepath", filePath, "ext", ext)
	switch ext {
	case ".sh":
		return ExecutorSh
	case ".py":
		return ExecutorPython
	case ".rb":
		return ExecutorRuby
	case ".pwsh":
		return ExecutorPowershell
	case ".ps1":
		return ExecutorPowershell
	case ".bat":
		return ExecutorCmd
	case "":
		return ExecutorBinary
	default:
		if runtime.GOOS == "windows" {
			return ExecutorCmd
		}
		return ExecutorSh
	}
}
