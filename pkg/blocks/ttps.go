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

// UnmarshalYAML is a custom unmarshalling implementation for the TTP structure.
// It decodes a YAML Node into a TTP object, handling the decoding and
// validation of the individual steps within the TTP.
//
// **Parameters:**
//
// node: A pointer to a yaml.Node that represents the TTP structure
// to be unmarshalled.
//
// **Returns:**
//
// error: An error if the decoding process fails or if the TTP structure contains invalid steps.
func (t *TTP) UnmarshalYAML(node *yaml.Node) error {
	type TTPTmpl struct {
		Name        string            `yaml:"name,omitempty"`
		Description string            `yaml:"description"`
		Environment map[string]string `yaml:"env,flow,omitempty"`
		Steps       []yaml.Node       `yaml:"steps,omitempty,flow"`
		ArgSpecs    []args.Spec       `yaml:"args,omitempty,flow"`
	}

	var tmpl TTPTmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	t.Name = tmpl.Name
	t.Description = tmpl.Description
	t.Environment = tmpl.Environment
	t.ArgSpecs = tmpl.ArgSpecs

	// Check for and handle a mitre node
	var mitreNode *yaml.Node
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		if keyNode.Value == "mitre" {
			mitreNode = node.Content[i+1]
			break
		}
	}

	if mitreNode != nil {
		if err := mitreNode.Decode(&t.MitreAttackMapping); err != nil {
			return err
		}
		// if we have a MitreAttackMapping, ensure there's a tactic
		if len(t.MitreAttackMapping.Tactics) == 0 {
			return fmt.Errorf("TTP '%s' has a MitreAttackMapping but no Tactic is defined", t.Name)
		}
	}

	return t.decodeSteps(tmpl.Steps)
}

func (t *TTP) decodeSteps(steps []yaml.Node) error {
	for stepIdx, stepNode := range steps {
		decoded := false
		// these candidate steps are pointers, so this line
		// MUST be inside the outer step loop or horrible things will happen
		// #justpointerthings
		stepTypes := []Step{NewBasicStep(), NewFileStep(), NewSubTTPStep(), NewEditStep(), NewFetchURIStep()}
		for _, stepType := range stepTypes {
			err := stepNode.Decode(stepType)
			if err == nil && !stepType.IsNil() {
				// Must catch bad steps with ambiguous types, such as:
				// - name: hello
				//   file: bar
				//   ttp: foo
				//
				// we can't use KnownFields to solve this without a massive
				// refactor due to https://github.com/go-yaml/yaml/issues/460
				if decoded {
					return fmt.Errorf("step #%v has ambiguous type", stepIdx+1)
				}
				logging.L().Debugw("decoded step", "step", stepType)
				t.Steps = append(t.Steps, stepType)
				decoded = true
			}
		}

		if !decoded {
			return t.handleInvalidStepError(stepNode)
		}
	}

	return nil
}

func (t *TTP) handleInvalidStepError(stepNode yaml.Node) error {
	act := Act{}
	err := stepNode.Decode(&act)

	if act.Name != "" && err != nil {
		return fmt.Errorf("invalid step found, missing parameters for step types")
	}
	return fmt.Errorf("invalid step found with no name, missing parameters for step types")
}

func (t *TTP) setWorkingDirectory() error {
	if t.WorkDir != "" {
		return nil
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	t.WorkDir = path
	return nil
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
		// pass in the directory
		stepCopy.SetDir(t.WorkDir)
		if err := stepCopy.Validate(execCtx); err != nil {
			logging.L().Errorw("failed to validate %s step: %v", step, zap.Error(err))
			return err
		}
	}
	logging.L().Info("[+] Finished validating steps")
	return nil
}

func (t *TTP) executeSteps(execCtx TTPExecutionContext) (*StepResultsRecord, []CleanupAct, error) {
	logging.L().Infof("[+] Running current TTP: %s", t.Name)
	stepResults := NewStepResultsRecord()
	execCtx.StepResults = stepResults
	var cleanup []CleanupAct

	for _, step := range t.Steps {
		stepCopy := step
		logging.L().Infof("[+] Running current step: %s", step.StepName())

		execResult, err := stepCopy.Execute(execCtx)
		if err != nil {
			return stepResults, cleanup, err
		}
		stepResults.ByName[step.StepName()] = execResult
		stepResults.ByIndex = append(stepResults.ByIndex, execResult)

		// Enters in reverse order
		cleanup = append(stepCopy.GetCleanup(), cleanup...)
		logging.L().Infof("[+] Finished running step: %s", step.StepName())
	}
	return stepResults, cleanup, nil
}

// RunSteps executes all of the steps in the given TTP.
//
// **Parameters:**
//
// execCfg: The TTPExecutionConfig for the current TTP.
//
// **Returns:**
//
// *StepResultsRecord: A StepResultsRecord containing the results of each step.
// error: An error if any of the steps fail to execute.
func (t *TTP) RunSteps(execCfg TTPExecutionConfig) (*StepResultsRecord, error) {
	if err := t.setWorkingDirectory(); err != nil {
		return nil, err
	}

	execCtx := TTPExecutionContext{
		Cfg: execCfg,
	}

	if err := t.ValidateSteps(execCtx); err != nil {
		return nil, err
	}

	stepResults, cleanup, err := t.executeSteps(execCtx)
	if err != nil {
		// we need to run cleanup so we don't return here
		logging.L().Errorf("[*] Error executing TTP: %v", err)
	}

	logging.L().Info("[*] Completed TTP")

	if !execCtx.Cfg.NoCleanup {
		if execCtx.Cfg.CleanupDelaySeconds > 0 {
			logging.L().Infof("[*] Sleeping for Requested Cleanup Delay of %v Seconds", execCtx.Cfg.CleanupDelaySeconds)
			time.Sleep(time.Duration(execCtx.Cfg.CleanupDelaySeconds) * time.Second)
		}
		if len(cleanup) > 0 {
			logging.L().Info("[*] Beginning Cleanup")
			if err := t.executeCleanupSteps(execCtx, cleanup, *stepResults); err != nil {
				logging.L().Errorw("error encountered in cleanup step: %v", err)
				return nil, err
			}
			logging.L().Info("[*] Finished Cleanup")
		} else {
			logging.L().Info("[*] No Cleanup Steps Found")
		}
	}

	return stepResults, err
}

func (t *TTP) executeCleanupSteps(execCtx TTPExecutionContext, cleanupSteps []CleanupAct, stepResults StepResultsRecord) error {
	for cleanupIdx, step := range cleanupSteps {
		stepCopy := step

		cleanupResult, err := stepCopy.Cleanup(execCtx)
		if err != nil {
			logging.L().Errorw("error encountered in stepCopy cleanup: %v", err)
			return err
		}
		// since ByIndex and ByName both contain pointers to
		// the same underlying struct, this will update both
		stepResults.ByIndex[len(cleanupSteps)-cleanupIdx-1].Cleanup = cleanupResult
	}
	return nil
}
