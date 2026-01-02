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

	"github.com/facebookincubator/ttpforge/pkg/fileutils"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
)

// RemovePathAction is invoked by
// adding remove_path to a given YAML step.
// It will delete the file at the specified path
// You must pass `recursive: true` to delete directories
type RemovePathAction struct {
	actionDefaults `yaml:",inline"`
	Path           string   `yaml:"remove_path,omitempty"`
	Recursive      bool     `yaml:"recursive,omitempty"`
	FileSystem     afero.Fs `yaml:"-,omitempty"`
}

// NewRemovePathAction creates a new RemovePathAction.
func NewRemovePathAction() *RemovePathAction {
	return &RemovePathAction{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *RemovePathAction) IsNil() bool {
	switch s.Path {
	case "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies
func (s *RemovePathAction) Validate(_ TTPExecutionContext) error {
	if s.Path == "" {
		return fmt.Errorf("path field cannot be empty")
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *RemovePathAction) Template(execCtx TTPExecutionContext) error {
	var err error
	s.Path, err = execCtx.templateStep(s.Path)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *RemovePathAction) Execute(_ TTPExecutionContext) (*ActResult, error) {
	logging.L().Infof("Removing path %v", s.Path)
	fsys := s.FileSystem
	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	// cannot remove a non-existent path
	pathToRemove, err := fileutils.ExpandPath(s.Path)
	if err != nil {
		return nil, err
	}
	exists, err := afero.Exists(fsys, pathToRemove)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("path %v does not exist", pathToRemove)
	}

	// afero fsys.Remove(...) appears to be buggy
	// and will remove a directory even if it is not empty
	// so we check manually - we use the semantics
	// of the macOS `rm` command and refuse to remove even
	// empty directories unless recursive is specified
	isDir, err := afero.IsDir(fsys, pathToRemove)
	if err != nil {
		return nil, err
	}

	if isDir && !s.Recursive {
		return nil, fmt.Errorf("path %v is a directory and `recursive: true` was not specified - refusing to remove", s.Path)
	}

	// actually remove the file
	err = fsys.RemoveAll(pathToRemove)
	if err != nil {
		return nil, err
	}
	return &ActResult{}, nil
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *RemovePathAction) CanBeUsedInCompositeAction() bool {
	return true
}
