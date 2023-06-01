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
	"go.uber.org/zap"
)

type Edit struct {
	Old string `yaml:"old,omitempty"`
	New string `yaml:"new,omitempty"`
}

// EditStep represents one or more edits to a specific file
type EditStep struct {
	*Act       `yaml:",inline"`
	FileToEdit string `yaml:"edit_file,omitempty"`
	Edits      []Edit `yaml:"edits,omitempty"`
}

// NewEditStep creates a new EditStep instance with an initialized Act struct.
func NewEditStep() *EditStep {
	return &EditStep{
		Act: &Act{
			Type: StepEdit,
		},
	}
}

// GetCleanup returns the cleanup steps for a BasicStep.
func (s *EditStep) GetCleanup() []CleanupAct {
	// TODO: implement
	return []CleanupAct{}
}

// CleanupName returns the name of the cleanup step.
func (s *EditStep) CleanupName() string {
	return s.Name
}

// GetType returns the step type for a BasicStep.
func (s *EditStep) GetType() StepType {
	return s.Type
}

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

// Validate validates the BasicStep, checking for the necessary attributes and dependencies.
func (s *EditStep) Validate() error {
	// Validate Act
	if err := s.Act.Validate(); err != nil {
		logging.Logger.Sugar().Error(zap.Error(err))
		return err
	}

	var err error
	if len(s.Edits) == 0 {
		err = fmt.Errorf("no edits specified")
	} else {
		for editIdx, edit := range s.Edits {
			if edit.Old == "" {
				err = fmt.Errorf("edit #%d is missing 'old:'", editIdx+1)
			} else if edit.New == "" {
				err = fmt.Errorf("edit #%d is missing 'new:'", editIdx+1)
			}
		}
	}

	if s.Name != "" && err != nil {
		return fmt.Errorf("[!] invalid editstep: [%s] %w", s.Name, err)
	}
	return nil
}

// Execute runs the EditStep and returns an error if any occur.
func (s *EditStep) Execute(inputs map[string]string) (err error) {
	logging.Logger.Sugar().Panic("not implemented")
	return nil
}
