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
	Append string `yaml:"append,omitempty"`
	Delete string `yaml:"delete,omitempty"`
	Regexp bool   `yaml:"regexp,omitempty"`

	oldRegexp *regexp.Regexp
}

// EditStep represents one or more edits to a specific file
type EditStep struct {
	actionDefaults `yaml:",inline"`
	FileToEdit     string   `yaml:"edit_file,omitempty"`
	Edits          []*Edit  `yaml:"edits,omitempty"`
	FileSystem     afero.Fs `yaml:"-,omitempty"`
	BackupFile     string   `yaml:"backup_file,omitempty"`
}

// NewEditStep creates a new EditStep instance with an initialized Act struct.
func NewEditStep() *EditStep {
	return &EditStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *EditStep) IsNil() bool {
	switch {
	case s.FileToEdit == "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies
func (s *EditStep) Validate(execCtx TTPExecutionContext) error {
	if len(s.Edits) == 0 {
		return fmt.Errorf("no edits specified")
	}

	targetPath := s.FileToEdit
	fileSystem := s.FileSystem

	if fileSystem == nil {
		_, err := FetchAbs(targetPath, execCtx.Vars.WorkDir)
		if err != nil {
			return err
		}
	}

	for editIdx, edit := range s.Edits {

		if edit.Append == "" && edit.Delete == "" {
			if edit.Old == "" {
				return fmt.Errorf("edit #%d is missing 'old:'", editIdx+1)
			} else if edit.New == "" {
				return fmt.Errorf("edit #%d is missing 'new:'", editIdx+1)
			}
		} else if edit.Append != "" {
			if edit.Old != "" {
				return fmt.Errorf("append is not to be used in conjunction with 'old:'")
			} else if edit.New != "" {
				return fmt.Errorf("append is not to be used in conjunction with 'new:'")

			} else if edit.Regexp {
				return fmt.Errorf("append is not to be used in conjunction with 'regexp:'")

			}
		} else if edit.Delete != "" {
			if edit.Old != "" {
				return fmt.Errorf("delete is not to be used in conjunction with 'old:'")
			} else if edit.New != "" {
				return fmt.Errorf("delete is not to be used in conjunction with 'new:'")

			}
		}

		oldStr := edit.Old
		if edit.Delete != "" {
			oldStr = edit.Delete
		}
		var err error
		if edit.Regexp {
			edit.oldRegexp, err = regexp.Compile(oldStr)
			if err != nil {
				return fmt.Errorf("edit #%d has invalid regex for 'old:'", editIdx+1)
			}
		} else {
			edit.oldRegexp = regexp.MustCompile(regexp.QuoteMeta(oldStr))
		}
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *EditStep) Template(execCtx TTPExecutionContext) error {
	var err error
	s.FileToEdit, err = execCtx.templateStep(s.FileToEdit)
	if err != nil {
		return err
	}
	s.BackupFile, err = execCtx.templateStep(s.BackupFile)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *EditStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	fileSystem := s.FileSystem
	targetPath := s.FileToEdit
	backupPath := s.BackupFile

	if fileSystem == nil {
		fileSystem = afero.NewOsFs()
		var err error
		targetPath, err = FetchAbs(targetPath, execCtx.Vars.WorkDir)
		if err != nil {
			return nil, err
		}
		if backupPath != "" {
			backupPath, err = FetchAbs(backupPath, execCtx.Vars.WorkDir)
			if err != nil {
				return nil, err
			}
		}
	}

	rawContents, err := afero.ReadFile(fileSystem, targetPath)
	if err != nil {
		return nil, err
	}

	contents := string(rawContents)

	if backupPath != "" {
		err = afero.WriteFile(fileSystem, backupPath, []byte(contents), 0644)
		if err != nil {
			return nil, fmt.Errorf("could not write backup file %v: %w", s.BackupFile, err)
		}
	}

	// this is inefficient - searches string 2 * num_edits times -
	// but it's unlikely to be a performance issue in practice. If it is,
	// we can optimize
	for editIdx, edit := range s.Edits {
		if edit.Append != "" {
			contents += "\n" + edit.Append
			continue
		}

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
		newStr := edit.New
		if edit.Delete != "" {
			newStr = ""
		}
		contents = edit.oldRegexp.ReplaceAllString(contents, newStr)
	}

	err = afero.WriteFile(fileSystem, targetPath, []byte(contents), 0644)
	if err != nil {
		return nil, err
	}

	return &ActResult{}, nil
}

// GetDefaultCleanupAction will instruct the calling code
// to copy the file to the backup file to the original path on cleanup.
func (s *EditStep) GetDefaultCleanupAction() Action {
	if s.BackupFile != "" {
		return &CompositeAction{
			actions: []Action{
				&CopyPathStep{
					Source:      s.BackupFile,
					Destination: s.FileToEdit,
					Overwrite:   true,
					FileSystem:  s.FileSystem,
				},
				&RemovePathAction{
					Path: s.BackupFile,
				},
			},
		}
	}
	return nil
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *EditStep) CanBeUsedInCompositeAction() bool {
	return true
}
