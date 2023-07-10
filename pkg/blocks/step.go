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
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Constants representing supported executor types.
const (
	ExecutorPython     = "python3"
	ExecutorBash       = "bash"
	ExecutorSh         = "sh"
	ExecutorPowershell = "powershell"
	ExecutorRuby       = "ruby"
	ExecutorBinary     = "binary"
	ExecutorCmd        = "cmd.exe"
)

// StepType denotes the type of a step in a TTP.
type StepType string

// Constants for defining the types of steps available.
const (
	StepUnset   = "unsetStep"
	StepFile    = "fileStep"
	StepBasic   = "basicStep"
	StepSubTTP  = "subTTPStep"
	StepCleanup = "cleanupStep"
	StepEdit    = "editStep"
)

// Act represents a single action within a TTP (Tactics, Techniques,
// and Procedures) step.
//
// Condition: The condition that needs to be satisfied for the Act to execute.
// Environment: Environment variables used during the Act's execution.
// Name: The unique name of the Act.
// WorkDir: The working directory of the Act.
// Type: The type of the Act (e.g., Command, File, or Setup).
// success: Indicates whether the execution of the Act was successful.
// stepRef: Reference to other steps in the sequence.
// output: The output of the Act's execution.
type Act struct {
	Condition   string            `yaml:"if,omitempty"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Name        string            `yaml:"name"`
	WorkDir     string            `yaml:"-"`
	Type        StepType          `yaml:"-"`
	stepRef     map[string]Step
	output      map[string]any
}

// NewAct is a constructor for the Act struct.
func NewAct() *Act {
	return &Act{
		output: make(map[string]interface{}),
	}
}

// CleanupAct interface is implemented by anything that requires a cleanup step.
type CleanupAct interface {
	Cleanup(execCtx TTPExecutionContext) (*ActResult, error)
	StepName() string
	SetDir(dir string)
	IsNil() bool
	Validate(execCtx TTPExecutionContext) error
}

// Step is an interface that represents a TTP step. Types that implement
// this interface must provide methods for setting up the environment and
// output references, setting the working directory, getting the cleanup
// actions, executing the step, checking if the step is empty, explaining
// validation errors, validating the step, fetching arguments, getting output,
// searching output, setting output success status, checking success status,
// returning the step name, and getting the step type.
type Step interface {
	SetDir(dir string)
	// Need list in case some steps are encapsulating many cleanup steps
	GetCleanup() []CleanupAct
	// Execute will need to take care of the condition checks/etc...
	Execute(execCtx TTPExecutionContext) (*ExecutionResult, error)
	IsNil() bool
	ExplainInvalid() error
	Validate(execCtx TTPExecutionContext) error
	GetOutput() map[string]any
	StepName() string
	GetType() StepType
}

// SetDir sets the working directory for the Act.
//
// **Parameters:**
//
// dir: A string representing the directory path to be set
// as the working directory.
func (a *Act) SetDir(dir string) {
	a.WorkDir = dir
}

// IsNil checks whether the Act is nil (i.e., it does not have a name).
//
// **Returns:**
//
// bool: True if the Act has no name, false otherwise.
func (a *Act) IsNil() bool {
	switch {
	case a.Name == "":
		return true
	default:
		return false
	}
}

// ExplainInvalid returns an error explaining why the Act is invalid.
//
// **Returns:**
//
// error: An error explaining why the Act is invalid, or nil
// if the Act is valid.
func (a *Act) ExplainInvalid() error {
	switch {
	case a.Name == "":
		return errors.New("no name provided for current step")
	default:
		return nil
	}
}

// StepName returns the name of the Act.
//
// **Returns:**
//
// string: The name of the Act.
func (a *Act) StepName() string {
	return a.Name
}

// GetOutput returns the output map of the Act.
//
// **Returns:**
//
// map[string]any: The output map of the Act.
func (a *Act) GetOutput() map[string]any {
	return a.output
}

// Validate checks the Act for any validation errors, such as the presence of
// spaces in the name.
//
// **Returns:**
//
// error: An error if any validation errors are found, or nil if
// the Act is valid.
func (a *Act) Validate() error {
	// Make sure name is of format we can index
	if strings.Contains(a.Name, " ") {
		return errors.New("name must not contain spaces")
	}

	return nil
}

// CheckCondition checks the condition specified for an Act and returns true
// if it matches the current OS, false otherwise. If the condition is "always",
// the function returns true.
// If an error occurs while checking the condition, it is returned.
//
// **Returns:**
//
// bool: true if the condition matches the current OS or the
// condition is "always", false otherwise.
//
// error: An error if an error occurs while checking the condition.
func (a *Act) CheckCondition() (bool, error) {
	switch a.Condition {
	case "windows":
		if runtime.GOOS == "windows" {
			return true, nil
		}
	case "darwin":
		if runtime.GOOS == "darwin" {
			return true, nil
		}
	case "linux":
		if runtime.GOOS == "linux" {
			return true, nil
		}
	// Run even if a previous step has failed.
	case "always":
		return true, nil

	default:
		return false, nil
	}
	return false, nil
}

// MakeCleanupStep creates a CleanupAct based on the given yaml.Node.
// If the node is empty or invalid, it returns nil. If the node contains a
// BasicStep or FileStep, the corresponding CleanupAct is created and returned.
//
// **Parameters:**
//
// node: A pointer to a yaml.Node containing the parameters to
// create the CleanupAct.
//
// **Returns:**
//
// CleanupAct: The created CleanupAct, or nil if the node is empty or invalid.
//
// error: An error if the node contains invalid parameters.
func (a *Act) MakeCleanupStep(node *yaml.Node) (CleanupAct, error) {
	if node.IsZero() {
		return nil, nil
	}

	basic, berr := a.tryDecodeBasicStep(node)
	if berr == nil && !basic.IsNil() {
		logging.Logger.Sugar().Debugw("cleanup step found", "basicstep", basic)
		return basic, nil
	}

	file, ferr := a.tryDecodeFileStep(node)
	if ferr == nil && !file.IsNil() {
		logging.Logger.Sugar().Debugw("cleanup step found", "filestep", file)
		return file, nil
	}

	err := fmt.Errorf("invalid parameters for cleanup steps with basic [%v], file [%v]", berr, ferr)
	logging.Logger.Sugar().Errorw(err.Error(), zap.Error(err))
	return nil, err
}

func (a *Act) tryDecodeBasicStep(node *yaml.Node) (*BasicStep, error) {
	basic := NewBasicStep()
	err := node.Decode(&basic)
	if err == nil && basic.Name == "" {
		basic.Name = fmt.Sprintf("cleanup-%s", a.Name)
		basic.Type = StepCleanup
	}
	return basic, err
}

func (a *Act) tryDecodeFileStep(node *yaml.Node) (*FileStep, error) {
	file := NewFileStep()
	err := node.Decode(&file)
	if err == nil && file.Name == "" {
		file.Name = fmt.Sprintf("cleanup-%s", a.Name)
		file.Type = StepCleanup
	}
	return file, err
}
