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
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
)

// CreateFileStep creates a new file and populates it
// with the specified contents.
// Its intended use is simulating malicious file creation
// through an editor program or via a C2, where there is no
// corresponding shell history telemetry
type CreateFileStep struct {
	Path       string   `yaml:"create_file,omitempty"`
	Contents   string   `yaml:"contents,omitempty"`
	Overwrite  bool     `yaml:"overwrite,omitempty"`
	Mode       int      `yaml:"mode,omitempty"`
	FileSystem afero.Fs `yaml:"-,omitempty"`
}

// NewCreateFileStep creates a new CreateFileStep instance and returns a pointer to it.
func NewCreateFileStep() *CreateFileStep {
	return &CreateFileStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *CreateFileStep) IsNil() bool {
	switch {
	case s.Path == "":
		return true
	default:
		return false
	}
}

// Execute runs the step and returns an error if any occur.
func (s *CreateFileStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
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
	// use the default umask
	// https://stackoverflow.com/questions/23842247/reading-default-filemode-when-using-os-o-create
	mode := s.Mode
	if mode == 0 {
		mode = 0666
	}
	f, err = fsys.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE, os.FileMode(mode))
	if err != nil {
		return nil, err
	}
	_, err = f.Write([]byte(s.Contents))
	if err != nil {
		return nil, err
	}

	return &ActResult{}, nil
}

// Validate validates the step
//
// **Returns:**
//
// error: An error if any validation checks fail.
func (s *CreateFileStep) Validate(execCtx TTPExecutionContext) error {
	if s.Path == "" {
		return fmt.Errorf("path field cannot be empty")
	}
	return nil
}
