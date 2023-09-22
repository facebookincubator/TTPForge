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
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// CreateFileStep creates a new file and populates it
// with the specified contents.
// Its intended use is simulating malicious file creation
// through an editor program or via a C2, where there is no
// corresponding shell history telemetry
type CreateFileStep struct {
	Act         `yaml:",inline"`
	Path        string      `yaml:"create_file,omitempty"`
	Contents    string      `yaml:"contents,omitempty"`
	Overwrite   bool        `yaml:"overwrite,omitempty"`
	Perm        os.FileMode `yaml:"perm,omitempty"`
	CleanupStep CleanupAct  `yaml:"cleanup,omitempty,flow"`
	FileSystem  afero.Fs    `yaml:"-,omitempty"`
}

// NewCreateFileStep creates a new CreateFileStep instance and returns a pointer to it.
func NewCreateFileStep() *FetchURIStep {
	return &FetchURIStep{
		Act: &Act{
			Type: StepFetchURI,
		},
	}
}

// UnmarshalYAML decodes a YAML node into a CreateFileStep instance. It uses
// the provided struct as a template for the YAML data, and initializes the
// CreateFileStep instance with the decoded values.
//
// **Parameters:**
//
// node: A pointer to a yaml.Node representing the YAML data to decode.
//
// **Returns:**
//
// error: An error if there is a problem decoding the YAML data.
func (s *CreateFileStep) UnmarshalYAML(node *yaml.Node) error {

	type createFileStepTmpl struct {
		Act         `yaml:",inline"`
		Path        string    `yaml:"create_file,omitempty"`
		Contents    string    `yaml:"contents,omitempty"`
		Overwrite   bool      `yaml:"overwrite,omitempty"`
		CleanupStep yaml.Node `yaml:"cleanup,omitempty,flow"`
	}

	// Decode the YAML node into the provided template.
	var tmpl createFileStepTmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	// Initialize the instance with the decoded values.
	s.Act = tmpl.Act
	s.Path = tmpl.Path
	s.Contents = tmpl.Contents
	s.Overwrite = tmpl.Overwrite

	// Check for invalid steps.
	if s.IsNil() {
		return s.ExplainInvalid()
	}

	// If there is no cleanup step or if this step is the cleanup step, exit.
	if tmpl.CleanupStep.IsZero() || s.Type == StepCleanup {
		return nil
	}

	// Create a CleanupStep instance and add it to this step
	cleanup, err := s.MakeCleanupStep(&tmpl.CleanupStep)
	if err != nil {
		return err
	}

	s.CleanupStep = cleanup

	return nil
}

// GetType returns the type of the step as StepType.
func (s *CreateFileStep) GetType() StepType {
	return StepCreateFile
}

// Cleanup is a method to establish a link with the Cleanup interface.
// Assumes that the type is the cleanup step and is invoked by
// s.CleanupStep.Cleanup.
func (s *CreateFileStep) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	result, err := s.Execute(execCtx)
	if err != nil {
		return nil, err
	}
	return &result.ActResult, err
}

// GetCleanup returns a slice of CleanupAct if the CleanupStep is not nil.
func (s *CreateFileStep) GetCleanup() []CleanupAct {
	if s.CleanupStep != nil {
		return []CleanupAct{s.CleanupStep}
	}
	return []CleanupAct{}
}

// ExplainInvalid returns an error message explaining why the step
// is invalid.
//
// **Returns:**
//
// error: An error message explaining why the step is invalid.
func (s *CreateFileStep) ExplainInvalid() error {
	if s.Path == "" {
		return errors.New("empty FetchURI provided")
	}
	return nil
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *CreateFileStep) IsNil() bool {
	switch {
	case s.Act.IsNil():
		return true
	case s.Path == "":
		return true
	default:
		return false
	}
}

// Execute runs the step and returns an error if any occur.
func (s *CreateFileStep) Execute(execCtx TTPExecutionContext) (*ExecutionResult, error) {
	logging.L().Infof("Creating file %v", s.Path)
	fsys := s.FileSystem
	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	exists, err := afero.Exists(fsys, s.Path)
	if err != nil {
		return nil, err
	}

	var f afero.File
	if exists && !s.Overwrite {
		return nil, fmt.Errorf("path %v already exists and overwrite was not set", s.Path)
	}
	f, err = fsys.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE, s.Perm)
	if err != nil {
		return nil, err
	}
	_, err = f.Write([]byte(s.Contents))
	if err != nil {
		return nil, err
	}

	return &ExecutionResult{}, nil
}

// Validate validates the step
//
// **Returns:**
//
// error: An error if any validation checks fail.
func (s *CreateFileStep) Validate(execCtx TTPExecutionContext) error {
	if err := s.Act.Validate(); err != nil {
		return err
	}

	if s.Path == "" {
		return fmt.Errorf("path field cannot be empty")
	}

	if s.CleanupStep != nil {
		if err := s.CleanupStep.Validate(execCtx); err != nil {
			logging.L().Errorw("error validating cleanup step", zap.Error(err))
			return err
		}
	}

	return nil
}
