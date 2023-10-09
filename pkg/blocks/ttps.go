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
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/args"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TTP represents the top-level structure for a TTP
// (Tactics, Techniques, and Procedures) object.
//
// **Attributes:**
//
// Name: The name of the TTP.
// Description: A description of the TTP.
// MitreAttackMapping: A MitreAttack object containing mappings to the MITRE ATT&CK framework.
// Environment: A map of environment variables to be set for the TTP.
// Steps: An slice of steps to be executed for the TTP.
// ArgSpecs: An slice of argument specifications for the TTP.
// WorkDir: The working directory for the TTP.
type TTP struct {
	Name               string            `yaml:"name,omitempty"`
	Description        string            `yaml:"description"`
	MitreAttackMapping MitreAttack       `yaml:"mitre,omitempty"`
	Environment        map[string]string `yaml:"env,flow,omitempty"`
	Steps              []Step            `yaml:"steps,omitempty,flow"`
	ArgSpecs           []args.Spec       `yaml:"args,omitempty,flow"`
	// Omit WorkDir, but expose for testing.
	WorkDir string `yaml:"-"`
}

// MitreAttack represents mappings to the MITRE ATT&CK framework.
//
// **Attributes:**
//
// Tactics: A string slice containing the MITRE ATT&CK tactic(s) associated with the TTP.
// Techniques: A string slice containing the MITRE ATT&CK technique(s) associated with the TTP.
// SubTechniques: A string slice containing the MITRE ATT&CK sub-technique(s) associated with the TTP.
type MitreAttack struct {
	Tactics       []string `yaml:"tactics,omitempty"`
	Techniques    []string `yaml:"techniques,omitempty"`
	SubTechniques []string `yaml:"subtechniques,omitempty"`
}

// MarshalYAML is a custom marshalling implementation for the TTP structure.
// It encodes a TTP object into a formatted YAML string, handling the
// indentation and structure of the output YAML.
//
// **Returns:**
//
// interface{}: The formatted YAML string representing the TTP object.
// error: An error if the encoding process fails.
func (t *TTP) MarshalYAML() (interface{}, error) {
	marshaled, err := yaml.Marshal(*t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TTP to YAML: %v", err)
	}

	// This section is necessary to get the proper formatting.
	// Resource: https://pkg.go.dev/gopkg.in/yaml.v3#section-readme
	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal(marshaled, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %v", err)
	}

	b, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal back to YAML: %v", err)
	}

	formattedYAML := reduceIndentation(b, 2)

	return fmt.Sprintf("---\n%s", string(formattedYAML)), nil
}

func reduceIndentation(b []byte, n int) []byte {
	lines := bytes.Split(b, []byte("\n"))

	for i, line := range lines {
		// Replace tabs with spaces for consistent processing
		line = bytes.ReplaceAll(line, []byte("\t"), []byte("    "))

		trimmedLine := bytes.TrimLeft(line, " ")
		indentation := len(line) - len(trimmedLine)
		if indentation >= n {
			lines[i] = bytes.TrimPrefix(line, bytes.Repeat([]byte(" "), n))
		} else {
			lines[i] = trimmedLine
		}
	}

	return bytes.Join(lines, []byte("\n"))
}

// ValidateSteps iterates through each step in the TTP and validates it.
// It sets the working directory for each step before calling its Validate
// method. If any step fails validation, the method returns an error.
// If all steps are successfully validated, the method returns nil.
//
// **Parameters:**
//
// execCtx: The TTPExecutionContext for the current TTP.
//
// **Returns:**
//
// error: An error if any step validation fails, otherwise nil.
func (t *TTP) ValidateSteps(execCtx TTPExecutionContext) error {
	logging.L().Info("[*] Validating Steps")

	for _, step := range t.Steps {
		stepCopy := step
		if err := stepCopy.Validate(execCtx); err != nil {
			logging.L().Errorw("failed to validate %s step: %v", step, zap.Error(err))
			return err
		}
	}
	logging.L().Info("[+] Finished validating steps")
	return nil
}

func (t *TTP) chdir() (func(), error) {
	// note: t.WorkDir may not be set in tests but should
	// be set when actualy using `ttpforge run`
	if t.WorkDir == "" {
		return func() {}, nil
	}
	origDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := os.Chdir(t.WorkDir); err != nil {
		return nil, err
	}
	return func() {
		if err := os.Chdir(origDir); err != nil {
			logging.L().Errorf("could not restore original directory %v: %v", origDir, err)
		}
	}, nil
}

// Execute,executes all of the steps in the given TTP,
// then runs cleanup if appropriate
//
// **Parameters:**
//
// execCfg: The TTPExecutionConfig for the current TTP.
//
// **Returns:**
//
// *StepResultsRecord: A StepResultsRecord containing the results of each step.
// error: An error if any of the steps fail to execute.
func (t *TTP) Execute(execCfg TTPExecutionConfig) (*StepResultsRecord, error) {
	execCtx := &TTPExecutionContext{
		Cfg:     execCfg,
		WorkDir: t.WorkDir,
	}
	stepResults, lastStepToSucceedIdx, runErr := t.RunSteps(execCtx)
	if runErr != nil {
		// we need to run cleanup so we don't return here
		logging.L().Errorf("[*] Error executing TTP: %v", runErr)
	} else {
		logging.L().Info("[*] Completed TTP - No Errors :)")
	}
	if !execCtx.Cfg.NoCleanup {
		if execCtx.Cfg.CleanupDelaySeconds > 0 {
			logging.L().Infof("[*] Sleeping for Requested Cleanup Delay of %v Seconds", execCtx.Cfg.CleanupDelaySeconds)
			time.Sleep(time.Duration(execCtx.Cfg.CleanupDelaySeconds) * time.Second)
		}
		t.startCleanupAtStepIdx(lastStepToSucceedIdx, execCtx)
	}
	// still pass up the run error after our cleanup
	return stepResults, runErr
}

// RunSteps executes all of the steps in the given TTP.
//
// **Parameters:**
//
// execCtx: The current TTPExecutionContext
//
// **Returns:**
//
// *StepResultsRecord: A StepResultsRecord containing the results of each step.
// int: the index of the last successful step
// error: An error if any of the steps fail to execute.
func (t *TTP) RunSteps(execCtx *TTPExecutionContext) (*StepResultsRecord, int, error) {
	// go to the configuration directory for this TTP
	changeBack, err := t.chdir()
	if err != nil {
		return nil, -1, err
	}
	defer changeBack()

	// validate steps:
	// stop after validation for dry run
	if err := t.ValidateSteps(*execCtx); err != nil {
		return nil, -1, err
	}
	if execCtx.Cfg.DryRun {
		logging.L().Info("[*] Dry-Run Requested - Returning Early")
		return nil, -1, nil
	}

	// actually run all the steps
	logging.L().Infof("[+] Running current TTP: %s", t.Name)
	stepResults := NewStepResultsRecord()
	execCtx.StepResults = stepResults
	lastStepToSucceedIdx := -1
	for _, step := range t.Steps {
		stepCopy := step
		logging.L().Infof("[+] Running current step: %s", step.Name)

		stepResult, err := stepCopy.Execute(*execCtx)
		if err != nil {
			return stepResults, lastStepToSucceedIdx, err
		}
		lastStepToSucceedIdx += 1
		execResult := &ExecutionResult{
			ActResult: *stepResult,
		}
		stepResults.ByName[step.Name] = execResult
		stepResults.ByIndex = append(stepResults.ByIndex, execResult)

		// Enters in reverse order
		logging.L().Infof("[+] Finished running step: %s", step.Name)
	}
	return stepResults, lastStepToSucceedIdx, nil
}

func (t *TTP) startCleanupAtStepIdx(lastStepToSucceedIdx int, execCtx *TTPExecutionContext) error {
	// go to the configuration directory for this TTP
	changeBack, err := t.chdir()
	if err != nil {
		return err
	}
	defer changeBack()

	logging.L().Info("[*] Beginning Cleanup")
	for cleanupIdx := lastStepToSucceedIdx; cleanupIdx >= 0; cleanupIdx -= 1 {
		stepToCleanup := t.Steps[cleanupIdx]
		cleanupResult, err := stepToCleanup.Cleanup(*execCtx)
		if err != nil {
			logging.L().Errorw("error cleaning up step: %v", err)
			logging.L().Errorw("will continue to try to cleanup other steps", err)
			continue
		}
		// since ByIndex and ByName both contain pointers to
		// the same underlying struct, this will update both
		execCtx.StepResults.ByIndex[cleanupIdx].Cleanup = cleanupResult
	}
	logging.L().Info("[*] Finished Cleanup")
	return nil
}
