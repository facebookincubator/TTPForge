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
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/otiai10/copy"
	"github.com/spf13/afero"
)

// CopyPathStep creates a new file and populates it
// with the specified contents from an existing path.
// Its intended use is simulating malicious file copies
// via a C2, where there is no corresponding shell history
// telemetry.
type CopyPathStep struct {
	actionDefaults `yaml:",inline"`
	Source         string   `yaml:"copy_path,omitempty"`
	Destination    string   `yaml:"to,omitempty"`
	Recursive      bool     `yaml:"recursive,omitempty"`
	Overwrite      bool     `yaml:"overwrite,omitempty"`
	Mode           int      `yaml:"mode,omitempty"`
	FileSystem     afero.Fs `yaml:"-,omitempty"`
}

// NewCopyPathStep creates a new CopyPathStep instance and returns a pointer to it.
func NewCopyPathStep() *CopyPathStep {
	return &CopyPathStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *CopyPathStep) IsNil() bool {
	switch {
	case s.Source == "":
		return true
	case s.Destination == "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies
func (s *CopyPathStep) Validate(_ TTPExecutionContext) error {
	if s.Source == "" {
		return fmt.Errorf("src field cannot be empty")
	}
	if s.Destination == "" {
		return fmt.Errorf("dest field cannot be empty")
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *CopyPathStep) Template(execCtx TTPExecutionContext) error {
	var err error
	s.Source, err = execCtx.templateStep(s.Source)
	if err != nil {
		return err
	}
	s.Destination, err = execCtx.templateStep(s.Destination)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *CopyPathStep) Execute(_ TTPExecutionContext) (*ActResult, error) {
	logging.L().Infof("Copying file(s) from %v to %v", s.Source, s.Destination)
	fsys := s.FileSystem
	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	// check if source exists.
	sourceExists, err := afero.Exists(fsys, s.Source)
	if err != nil {
		return nil, err
	}

	// if source does not exist.
	if !sourceExists {
		return nil, fmt.Errorf("source %v does not exist", s.Source)
	}

	// if source is a directory but recursive is false
	srcInfo, err := fsys.Stat(s.Source)
	if err != nil {
		return nil, err
	}
	if srcInfo.IsDir() && !s.Recursive {
		return nil, fmt.Errorf("source %v is a directory, but the recursive flag is set to false", s.Source)
	}

	// check if destination exists.
	destExists, err := afero.Exists(fsys, s.Destination)
	if err != nil {
		return nil, err
	}
	// if destination exits, return error if overwrite flag is not true.
	if destExists && !s.Overwrite {
		return nil, fmt.Errorf("dest %v already exists and overwrite was not set", s.Destination)
	}

	// use the default umask
	// https://stackoverflow.com/questions/23842247/reading-default-filemode-when-using-os-o-create
	mode := s.Mode
	if mode == 0 {
		mode = 0666
	}

	// Copy a file
	err = copy.Copy(s.Source, s.Destination)
	if err != nil {
		return nil, err
	}

	return &ActResult{}, nil
}

// GetDefaultCleanupAction will instruct the calling code
// to remove the path created by this action
func (s *CopyPathStep) GetDefaultCleanupAction() Action {
	return &RemovePathAction{
		Path: s.Destination,
	}
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *CopyPathStep) CanBeUsedInCompositeAction() bool {
	return true
}
