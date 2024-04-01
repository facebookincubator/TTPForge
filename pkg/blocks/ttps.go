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
	"runtime"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/checks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/platforms"
	"gopkg.in/yaml.v3"
)

// TTP represents the top-level structure for a TTP
// (Tactics, Techniques, and Procedures) object.
//
// **Attributes:**
//
// Environment: A map of environment variables to be set for the TTP.
// Steps: An slice of steps to be executed for the TTP.
// WorkDir: The working directory for the TTP.
type TTP struct {
	PreambleFields `yaml:",inline"`
	Environment    map[string]string `yaml:"env,flow,omitempty"`
	Steps          []Step            `yaml:"steps,omitempty,flow"`
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

// Validate ensures that all components of the TTP are valid
// It checks key fields, then iterates through each step
// and validates them in turn
func (t *TTP) Validate(execCtx TTPExecutionContext) error {
	logging.L().Debugf("Validating TTP %q...", t.Name)

	// Validate preamble fields
	err := t.PreambleFields.Validate(false)
	if err != nil {
		return err
	}

	// Validate steps
	for _, step := range t.Steps {
		stepCopy := step
		if err := stepCopy.Validate(execCtx); err != nil {
			return err
		}
	}
	logging.L().Debug("...finished validating TTP.")
	return nil
}

// Execute executes all of the steps in the given TTP,
// then runs cleanup if appropriate
func (t *TTP) Execute(execCtx TTPExecutionContext) error {
	logging.L().Infof("RUNNING TTP: %v", t.Name)

	if err := t.verifyPlatform(); err != nil {
		return fmt.Errorf("TTP requirements not met: %w", err)
	}

	err := t.RunSteps(execCtx)
	if err == nil {
		logging.L().Info("All TTP steps completed successfully! ✅")
	}
	return err
}

// RunSteps executes all of the steps in the given TTP.
func (t *TTP) RunSteps(execCtx TTPExecutionContext) error {
	// go to the configuration directory for this TTP
	changeBack, err := t.chdir()
	if err != nil {
		return err
	}
	defer changeBack()

	var stepError error
	var verifyError error
	var shutdownFlag bool

	// actually run all the steps
	for stepIdx, step := range t.Steps {
		logging.DividerThin()
		logging.L().Infof("Executing Step #%d: %q", stepIdx+1, step.Name)
		// core execution - run the step action
		go func(step Step) {
			_, err := step.Execute(execCtx)
			if err != nil {
				// This error was logged by the step itself
				logging.L().Debugf("Error executing step %s: %v", step.Name, err)
			}
		}(step)

		// await one of three outcomes:
		// 1. step execution successful
		// 2. step execution failed
		// 3. shutdown signal received
		select {
		case stepResult := <-execCtx.actionResultsChan:
			// step execution successful - record results
			execResult := &ExecutionResult{
				ActResult: *stepResult,
			}
			execCtx.StepResults.ByName[step.Name] = execResult
			execCtx.StepResults.ByIndex = append(execCtx.StepResults.ByIndex, execResult)

		case stepError = <-execCtx.errorsChan:
			// this part is tricky - SubTTP steps
			// must be cleaned up even on failure
			// (because substeps may have succeeded)
			// so in those cases, we need to save the result
			// even if nil
			if step.ShouldCleanupOnFailure() {
				logging.L().Infof("[+] Cleaning up failed step %s", step.Name)
				logging.L().Infof("[+] Full Cleanup will Run Afterward")
				_, cleanupErr := step.Cleanup(execCtx)
				if cleanupErr != nil {
					logging.L().Errorf("Error cleaning up failed step %v: %v", step.Name, cleanupErr)
				}
			}

		case shutdownFlag = <-execCtx.shutdownChan:
			// TODO[nesusvet]: We should propagate signal to child processes if any
			logging.L().Warn("Shutting down due to signal received")
		}

		// if the user specified custom success checks, run them now
		verifyError = step.VerifyChecks()

		if stepError != nil || verifyError != nil || shutdownFlag {
			logging.L().Debug("[*] Stopping TTP Early")
			break
		}
	}

	logging.DividerThin()
	if stepError != nil {
		logging.L().Errorf("[*] Error executing TTP: %v", stepError)
		return stepError
	}
	if verifyError != nil {
		logging.L().Errorf("[*] Error verifying TTP: %v", verifyError)
		return verifyError
	}
	if shutdownFlag {
		return fmt.Errorf("[*] Shutting Down now")
	}

	return nil
}

// RunCleanup executes all required cleanup for steps in the given TTP.
func (t *TTP) RunCleanup(execCtx TTPExecutionContext) error {
	if execCtx.Cfg.NoCleanup {
		logging.L().Info("[*] Skipping Cleanup as requested by Config")
		return nil
	}

	if execCtx.Cfg.CleanupDelaySeconds > 0 {
		logging.L().Infof("[*] Sleeping for Requested Cleanup Delay of %v Seconds", execCtx.Cfg.CleanupDelaySeconds)
		time.Sleep(time.Duration(execCtx.Cfg.CleanupDelaySeconds) * time.Second)
	}

	// TODO[nesusvet]: We also should catch signals in clean ups
	cleanupResults, err := t.startCleanupForCompletedSteps(execCtx)
	if err != nil {
		return err
	}
	// since ByIndex and ByName both contain pointers to
	// the same underlying struct, this will update both
	for cleanupIdx, cleanupResult := range cleanupResults {
		execCtx.StepResults.ByIndex[cleanupIdx].Cleanup = cleanupResult
	}

	return nil
}

func (t *TTP) chdir() (func(), error) {
	// note: t.WorkDir may not be set in tests but should
	// be set when actually using `ttpforge run`
	if t.WorkDir == "" {
		logging.L().Info("Not changing working directory in tests")
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

// verify that we actually meet the necessary requirements to execute this TTP
func (t *TTP) verifyPlatform() error {
	verificationCtx := checks.VerificationContext{
		Platform: platforms.Spec{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
	return t.Requirements.Verify(verificationCtx)
}

func (t *TTP) startCleanupForCompletedSteps(execCtx TTPExecutionContext) ([]*ActResult, error) {
	// go to the configuration directory for this TTP
	changeBack, err := t.chdir()
	if err != nil {
		return nil, err
	}
	defer changeBack()

	logging.DividerThick()
	n := len(execCtx.StepResults.ByIndex)
	logging.L().Infof("CLEANING UP %v steps of TTP: %q", n, t.Name)
	cleanupResults := make([]*ActResult, n)
	for cleanupIdx := n - 1; cleanupIdx >= 0; cleanupIdx-- {
		stepToCleanup := t.Steps[cleanupIdx]
		logging.DividerThin()
		logging.L().Infof("Cleaning Up Step #%d: %q", cleanupIdx+1, stepToCleanup.Name)
		cleanupResult, err := stepToCleanup.Cleanup(execCtx)
		// must be careful to put these in step order, not in execution (reverse) order
		cleanupResults[cleanupIdx] = cleanupResult
		if err != nil {
			logging.L().Errorf("error cleaning up step: %v", err)
			logging.L().Errorf("will continue to try to cleanup other steps")
			continue
		}
	}
	logging.DividerThin()
	logging.L().Info("Finished Cleanup Successfully ✅")
	return cleanupResults, nil
}
