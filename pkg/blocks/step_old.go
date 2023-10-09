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
	"strings"
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
	StepCreateFile = "createFileStep"
	StepUnset      = "unsetStep"
	StepFile       = "fileStep"
	StepFetchURI   = "fetchURIStep"
	StepBasic      = "basicStep"
	StepSubTTP     = "subTTPStep"
	StepCleanup    = "cleanupStep"
	StepEdit       = "editStep"
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
	Name string   `yaml:"name"`
	Type StepType `yaml:"-"`
}

// CleanupAct interface is implemented by anything that requires a cleanup step.
type CleanupAct interface {
	Cleanup(execCtx TTPExecutionContext) (*ActResult, error)
	StepName() string
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
type StepInterface interface {
	// Need list in case some steps are encapsulating many cleanup steps
	GetCleanup() []CleanupAct
	// Execute will need to take care of the condition checks/etc...
	Execute(execCtx TTPExecutionContext) (*ActResult, error)
	IsNil() bool
	ExplainInvalid() error
	Validate(execCtx TTPExecutionContext) error
	StepName() string
	GetType() StepType
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
