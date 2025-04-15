/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
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

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
)

// ChangeDirectoryStep is a step that changes the current working directory
type ChangeDirectoryStep struct {
	actionDefaults `yaml:",inline"`
	Cd             string `yaml:"cd"`
	PreviousDir    string
	PreviousCDStep *ChangeDirectoryStep
	FileSystem     afero.Fs `yaml:"-,omitempty"`
}

// NewChangeDirectoryStep creates a new ChangeDirectoryStep instance with an initialized Act struct.
func NewChangeDirectoryStep() *ChangeDirectoryStep {
	return &ChangeDirectoryStep{}
}

// IsNil checks if a ChangeDirectoryStep is considered empty or unitializied
func (step *ChangeDirectoryStep) IsNil() bool {
	return step.Cd == ""
}

// Validate validates the ChangeDirectoryStep, checking for the necessary attributes and dependencies.
//
// **Returns:**
//
// error: error if validation fails, nil otherwise
func (step *ChangeDirectoryStep) Validate(_ TTPExecutionContext) error {
	// If this has a parent cd step, hold off on validation until execute
	if step.PreviousCDStep != nil {
		return nil
	}

	// Check if cd is provided
	if step.Cd == "" {
		err := errors.New("cd must be provided")
		return err
	}

	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (step *ChangeDirectoryStep) Template(execCtx TTPExecutionContext) error {
	var err error
	step.Cd, err = execCtx.templateStep(step.Cd)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the ChangeDirectoryStep, changing the current working directory and returns an error if any occur.
//
// **Returns:**
//
// ActResult: the result of the action
// error: error if execution fails, nil otherwise
func (step *ChangeDirectoryStep) Execute(ctx TTPExecutionContext) (*ActResult, error) {
	// If this has a parent, then it's a cleanup step, so we need to grab the previous dir from it
	if step.PreviousCDStep != nil {
		if step.PreviousCDStep.PreviousDir == "" {
			return nil, fmt.Errorf("no previous directory found in parent cd step")
		}
		step.Cd = step.PreviousCDStep.PreviousDir
	}

	// Check if cd is a valid directory
	fsys := step.FileSystem
	if fsys == nil {
		if step.PreviousCDStep != nil && step.PreviousCDStep.FileSystem != nil {
			fsys = step.PreviousCDStep.FileSystem
		} else {
			fsys = afero.NewOsFs()
		}
	}

	exists, err := afero.DirExists(fsys, step.Cd)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("directory \"%s\" does not exist", step.Cd)
	}

	logging.L().Infof("Changing directory to %s", step.Cd)

	if step.Cd == "" {
		return nil, fmt.Errorf("empty cd value in Execute(...)")
	}

	// Set workdir to the current cd value and store the previous workdir
	step.PreviousDir = ctx.Vars.WorkDir
	ctx.Vars.WorkDir = step.Cd

	return &ActResult{}, nil
}

// GetDefaultCleanupAction sets the directory back to the previous directory
func (step *ChangeDirectoryStep) GetDefaultCleanupAction() Action {
	return &ChangeDirectoryStep{
		PreviousCDStep: step,
	}
}
