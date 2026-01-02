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
	actionDefaults `yaml:",inline"`
	TtpRef         string            `yaml:"ttp"`
	Args           map[string]string `yaml:"args"`

	ttp        *TTP
	subExecCtx *TTPExecutionContext
}

// NewSubTTPStep creates a new SubTTPStep and returns a pointer to it.
func NewSubTTPStep() *SubTTPStep {
	return &SubTTPStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *SubTTPStep) IsNil() bool {
	switch s.TtpRef {
	case "":
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
	if s.TtpRef == "" {
		return errors.New("a TTP reference is required and must not be empty")
	}

	// validate subttp
	if err := s.loadSubTTP(execCtx); err != nil {
		return err
	}

	return s.ttp.Validate(execCtx)
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *SubTTPStep) Template(execCtx TTPExecutionContext) error {
	var err error

	for key, value := range s.Args {
		s.Args[key], err = execCtx.templateStep(value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Execute runs each step of the TTP file associated with the SubTTPStep
// and manages the outputs and cleanup steps.
func (s *SubTTPStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	logging.L().Infof("[*] Executing Sub TTP: %s", s.TtpRef)
	logging.IncreaseIndentLevel()
	runErr := s.ttp.RunSteps(*s.subExecCtx)
	if runErr != nil {
		return &ActResult{}, runErr
	}
	logging.DecreaseIndentLevel()
	logging.L().Info("[*] Completed SubTTP - No Errors :)")

	// just a little annoying plumbing due to subtle type differences
	actResults := make([]*ActResult, len(s.subExecCtx.StepResults.ByIndex))
	for index, execResult := range s.subExecCtx.StepResults.ByIndex {
		actResults[index] = &execResult.ActResult
	}
	result := aggregateResults(actResults)

	// Send stdout to the output variable in the parent execution context
	if s.OutputVar != "" {
		execCtx.Vars.StepVars[s.OutputVar] = strings.TrimSuffix(result.Stdout, "\n")
	}

	return result, nil
}

// GetDefaultCleanupAction will instruct the calling code
// to cleanup all successful steps of this subTTP
func (s *SubTTPStep) GetDefaultCleanupAction() Action {
	return &subTTPCleanupAction{
		step: s,
	}
}

func aggregateResults(results []*ActResult) *ActResult {
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

// loadSubTTP loads a TTP file into a SubTTPStep instance
// and validates the contained steps.
func (s *SubTTPStep) loadSubTTP(execCtx TTPExecutionContext) error {
	repo := execCtx.Cfg.Repo
	subTTPAbsPath, err := repo.FindTTP(s.TtpRef)
	if err != nil {
		return err
	}

	subArgsKv, err := s.processSubTTPArgs(execCtx)
	if err != nil {
		return err
	}

	ttps, ctx, err := LoadTTP(subTTPAbsPath, repo.GetFs(), &execCtx.Cfg, execCtx.Vars.StepVars, subArgsKv)
	if err != nil {
		return err
	}
	s.ttp = ttps
	s.subExecCtx = ctx

	return nil
}
