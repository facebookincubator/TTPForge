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

	"github.com/facebookincubator/ttpforge/pkg/logging"
)

// SubTTPStep represents a step within a parent TTP that references a separate TTP file.
type SubTTPStep struct {
	TtpFile string            `yaml:"ttp"`
	Args    map[string]string `yaml:"args"`

	ttp        *TTP
	subExecCtx TTPExecutionContext
}

// NewSubTTPStep creates a new SubTTPStep and returns a pointer to it.
func NewSubTTPStep() *SubTTPStep {
	return &SubTTPStep{}
}

func aggregateResults(results []*ExecutionResult) *ActResult {
	var subStdouts []string
	var subStderrs []string
	for _, result := range results {
		subStdouts = append(subStdouts, result.Stdout)
		subStderrs = append(subStderrs, result.Stderr)
	}

	return &ActResult{
		Stdout: strings.Join(subStdouts, ""),
		Stderr: strings.Join(subStderrs, ""),
	}
}

func (s *SubTTPStep) processSubTTPArgs(execCtx TTPExecutionContext) ([]string, error) {
	var argKvStrs []string
	for k, v := range s.Args {
		argKvStrs = append(argKvStrs, k+"="+v)
	}

	expandedArgKvStrs, err := execCtx.ExpandVariables(argKvStrs)
	if err != nil {
		return nil, err
	}
	return expandedArgKvStrs, nil
}

// Execute runs each step of the TTP file associated with the SubTTPStep
// and manages the outputs and cleanup steps.
func (s *SubTTPStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	logging.L().Infof("[*] Executing Sub TTP: %s", s.TtpFile)

	var runErr error
	stepResults, lastStepToSucceedIdx, runErr := s.ttp.RunSteps(&execCtx)
	if runErr != nil {
		// for subTTPs, we need to start cleanup now - cleanup of previous steps
		// will proceed according to the normal LIFO process
		logging.L().Errorf("[*] Error executing SubTTP: %v", runErr)
		logging.L().Errorf("[*] Beginning SubTTP cleanup")
		s.ttp.startCleanupAtStepIdx(lastStepToSucceedIdx, &execCtx)
		return nil, runErr
	}
	logging.L().Info("[*] Completed TTP - No Errors :)")
	return aggregateResults(stepResults.ByIndex), nil
}

// loadSubTTP loads a TTP file into a SubTTPStep instance
// and validates the contained steps.
func (s *SubTTPStep) loadSubTTP(execCtx TTPExecutionContext) error {
	repo := execCtx.Cfg.Repo
	subTTPAbsPath, err := execCtx.Cfg.Repo.FindTTP(s.TtpFile)
	if err != nil {
		return err
	}

	subArgsKv, err := s.processSubTTPArgs(execCtx)
	if err != nil {
		return err
	}

	ttps, err := LoadTTP(subTTPAbsPath, repo.GetFs(), &s.subExecCtx.Cfg, subArgsKv)
	if err != nil {
		return err
	}
	s.ttp = ttps
	return nil
}

// IsNil checks if the SubTTPStep is empty or uninitialized.
func (s *SubTTPStep) IsNil() bool {
	switch {
	case s.TtpFile == "":
		return true
	default:
		return false
	}
}

// Validate checks the validity of the SubTTPStep by ensuring
// the following conditions are met:
// The associated Act is valid.
// The TTP file associated with the SubTTPStep can be successfully unmarshalled.
// The TTP file path is not empty.
// The steps within the TTP file do not contain any nested SubTTPSteps.
// If any of these conditions are not met, an error is returned.
func (s *SubTTPStep) Validate(execCtx TTPExecutionContext) error {
	if s.TtpFile == "" {
		return errors.New("a TTP file path is required and must not be empty")
	}

	if err := s.loadSubTTP(execCtx); err != nil {
		return err
	}

	return s.ttp.ValidateSteps(execCtx)
}
