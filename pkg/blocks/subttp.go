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
	"io/fs"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"gopkg.in/yaml.v3"
)

// SubTTPStep represents a step within a parent TTP that references a separate TTP file.
type SubTTPStep struct {
	*Act       `yaml:",inline"`
	TtpFile    string    `yaml:"ttp"`
	FileSystem fs.StatFS `yaml:"-,omitempty"`
	// Omitting because the sub steps will contain the cleanups.
	CleanupSteps []CleanupAct `yaml:"-,omitempty"`
	ttp          *TTP
}

// NewSubTTPStep creates a new SubTTPStep and returns a pointer to it.
func NewSubTTPStep() *SubTTPStep {
	return &SubTTPStep{}
}

// GetCleanup returns a slice of CleanupAct associated with the SubTTPStep.
func (s *SubTTPStep) GetCleanup() []CleanupAct {
	return s.CleanupSteps
}

// UnmarshalYAML is a custom unmarshaller for SubTTPStep which decodes
// a YAML node into a SubTTPStep instance.
func (s *SubTTPStep) UnmarshalYAML(node *yaml.Node) error {
	type Subtmpl struct {
		Act     `yaml:",inline"`
		TtpFile string `yaml:"ttp"`
	}
	var substep Subtmpl

	if err := node.Decode(&substep); err != nil {
		return err
	}
	logging.Logger.Sugar().Debugw("step found", "substep", substep)

	s.Act = &substep.Act
	s.TtpFile = substep.TtpFile

	return nil
}

// UnmarshalSubTTP loads a TTP file associated with a SubTTPStep
// and stores it in the instance.
func (s *SubTTPStep) UnmarshalSubTTP() error {
	logging.Logger.Sugar().Debugw("parameters used to grab file", "filename", s.TtpFile, "workdir", s.WorkDir)
	fullpath, err := FindFilePath(s.TtpFile, s.WorkDir, s.FileSystem)
	if err != nil {
		return err
	}

	s.TtpFile = fullpath

	if err := s.loadSubTTP(); err != nil {
		return err
	}

	return nil
}

// Execute runs each step of the TTP file associated with the SubTTPStep
// and manages the outputs and cleanup steps.
func (s *SubTTPStep) Execute(inputs map[string]string) error {
	logging.Logger.Sugar().Infof("[*] Executing Sub TTP: %s", s.Name)
	availableSteps := make(map[string]Step)

	for key, val := range s.ttp.InputMap {
		inputs[key] = val
	}

	for _, step := range s.ttp.Steps {
		stepCopy := step
		stepCopy.Setup(s.Environment, availableSteps)
		logging.Logger.Sugar().Infof("[+] Running current step: %s", step.StepName())

		if err := stepCopy.Execute(inputs); err != nil {
			return err
		}

		output := stepCopy.GetOutput()

		availableSteps[stepCopy.StepName()] = stepCopy
		s.output[stepCopy.StepName()] = output

		stepClean := stepCopy.GetCleanup()
		if stepClean != nil {
			logging.Logger.Sugar().Debugw("adding cleanup step", "cleanup", stepClean)
			s.CleanupSteps = append(stepCopy.GetCleanup(), s.CleanupSteps...)
		}

		logging.Logger.Sugar().Debugw("available step data", "data", availableSteps[stepCopy.StepName()].GetOutput())
		logging.Logger.Sugar().Infof("[+] Finished running step: %s", stepCopy.StepName())
	}

	logging.Logger.Sugar().Info("Finished execution of sub ttp file")

	return nil
}

// loadSubTTP loads a TTP file into a SubTTPStep instance
// and validates the contained steps.
func (s *SubTTPStep) loadSubTTP() error {
	ttps, err := LoadTTP(s.TtpFile)
	if err != nil {
		return err
	}
	s.ttp = ttps

	// run validate to flesh out issues
	logging.Logger.Sugar().Infof("[*] Validating Sub TTP: %s", s.Name)
	for _, step := range s.ttp.Steps {
		stepCopy := step
		if err := stepCopy.Validate(); err != nil {
			return err
		}
	}
	logging.Logger.Sugar().Infof("[*] Finished validating Sub TTP")

	return nil
}

// GetType returns the type of the step (StepSubTTP for SubTTPStep).
func (s *SubTTPStep) GetType() StepType {
	return StepSubTTP
}

// ExplainInvalid checks for invalid data in the SubTTPStep
// and returns an error explaining any issues found.
// Currently, it checks if the TtpFile field is empty.
func (s *SubTTPStep) ExplainInvalid() error {
	if s.TtpFile == "" {
		err := fmt.Errorf("error: TtpFile is empty")
		if s.Name != "" {
			return fmt.Errorf("invalid SubTTPStep [%s]: %w", s.Name, err)
		}
		return err
	}
	return nil
}

// IsNil checks if the SubTTPStep is empty or uninitialized.
func (s *SubTTPStep) IsNil() bool {
	logging.Logger.Sugar().Info(s.Act)
	switch {
	case s.Act.IsNil():
		return true
	case s.TtpFile == "":
		return true
	default:
		return false
	}
}

// Validate checks the validity of the SubTTPStep by ensuring the following conditions are met:
// 1. The associated Act is valid.
// 2. The TTP file associated with the SubTTPStep can be successfully unmarshalled.
// 3. The TTP file path is not empty.
// 4. The steps within the TTP file do not contain any nested SubTTPSteps.
// If any of these conditions are not met, an error is returned.
func (s *SubTTPStep) Validate() error {
	if err := s.Act.Validate(); err != nil {
		return err
	}

	if err := s.UnmarshalSubTTP(); err != nil {
		return err
	}

	if s.TtpFile == "" {
		return errors.New("a TTP file path is required and must not be empty")
	}

	// Check if steps contain any SubTTPSteps. If they do, return an error.
	for _, steps := range s.ttp.Steps {
		if steps.GetType() == StepSubTTP {
			return errors.New(
				"nested SubTTPStep detected within a SubTTPStep, " +
					"please remove it for successful execution")
		}
	}

	return nil
}
