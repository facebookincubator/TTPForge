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
	"regexp"

	"github.com/spf13/afero"
)

// Edit represents a single old+new find-and-replace pair
type Edit struct {
	Old    string `yaml:"old,omitempty"`
	New    string `yaml:"new,omitempty"`
	Regexp bool   `yaml:"regexp,omitempty"`

	oldRegexp *regexp.Regexp
}

// EditStep represents one or more edits to a specific file
type EditStep struct {
	*Act       `yaml:",inline"`
	FileToEdit string   `yaml:"edit_file,omitempty"`
	Edits      []*Edit  `yaml:"edits,omitempty"`
	FileSystem afero.Fs `yaml:"-,omitempty"`
	BackupFile string   `yaml:"backup_file,omitempty"`
}

// NewEditStep creates a new EditStep instance with an initialized Act struct.
func NewEditStep() *EditStep {
	return &EditStep{
		Act: &Act{
			Type: StepEdit,
		},
	}
}

// GetCleanup returns the cleanup steps for a EditStep.
// Currently this is always empty because we use backup
// files instead for this type of step
func (s *EditStep) GetCleanup() []CleanupAct {
	return []CleanupAct{}
}

// GetType returns the step type for a EditStep.
func (s *EditStep) GetType() StepType {
	return s.Type
}

// IsNil checks if an EditStep is considered empty or uninitialized.
func (s *EditStep) IsNil() bool {
	switch {
	case s.Act.IsNil():
		return true
	case s.FileToEdit == "":
		return true
	default:
		return false
	}
}

// wrapped by exported Validate to standardize
// the error message prefix
func (s *EditStep) check() error {
	// Validate Act
	if err := s.Act.Validate(); err != nil {
		return err
	}

	var err error
	if len(s.Edits) == 0 {
		return fmt.Errorf("no edits specified")
	}

	// TODO: make this compatible with deleting lines
	for editIdx, edit := range s.Edits {
		if edit.Old == "" {
			return fmt.Errorf("edit #%d is missing 'old:'", editIdx+1)
		} else if edit.New == "" {
			return fmt.Errorf("edit #%d is missing 'new:'", editIdx+1)
		}

		if edit.Regexp {
			edit.oldRegexp, err = regexp.Compile(edit.Old)
			if err != nil {
				return fmt.Errorf("edit #%d has invalid regex for 'old:'", editIdx+1)
			}
		} else {
			edit.oldRegexp = regexp.MustCompile(regexp.QuoteMeta(edit.Old))
		}
	}
	return nil
}

// Validate validates the EditStep, checking for the necessary attributes and dependencies.
func (s *EditStep) Validate(execCtx TTPExecutionContext) error {
	err := s.check()
	if err != nil {
		return fmt.Errorf("[!] invalid editstep: [%s] %w", s.Name, err)
	}
	return nil
}

// Execute runs the EditStep and returns an error if any occur.
func (s *EditStep) Execute(execCtx TTPExecutionContext) (*ExecutionResult, error) {
	fileSystem := s.FileSystem
	targetPath := s.FileToEdit
	if fileSystem == nil {
		fileSystem = afero.NewOsFs()
		var err error
		targetPath, err = FetchAbs(targetPath, s.WorkDir)
		if err != nil {
			return nil, err
		}
	}
	rawContents, err := afero.ReadFile(fileSystem, targetPath)
	if err != nil {
		return nil, err
	}

	contents := string(rawContents)

	if s.BackupFile != "" {
		err = afero.WriteFile(fileSystem, s.BackupFile, []byte(contents), 0644)
		if err != nil {
			return nil, fmt.Errorf("could not write backup file %v: %v", s.BackupFile, err)
		}
	}

	// this is inefficient - searches string 2 * num_edits times -
	// but it's unlikely to be a performance issue in practice. If it is,
	// we can optimize
	for editIdx, edit := range s.Edits {
		matches := edit.oldRegexp.FindAllStringIndex(contents, -1)
		// we want to error here because otherwise ppl will be confused by silent
		// failures if the format of the file they're trying to edit changes
		// and their regexes no longer work
		if len(matches) == 0 {
			return nil, fmt.Errorf(
				"pattern '%v' from edit #%d was not found in file %v",
				edit.Old,
				editIdx+1,
				s.FileToEdit,
			)
		}
		contents = edit.oldRegexp.ReplaceAllString(contents, edit.New)
	}

	err = afero.WriteFile(fileSystem, targetPath, []byte(contents), 0644)
	if err != nil {
		return nil, err
	}

	return &ExecutionResult{}, nil
}
