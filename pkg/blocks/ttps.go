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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	cp "github.com/otiai10/copy"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TTP represents the top-level structure for a TTP (Tactics, Techniques, and Procedures) object.
// DOES NOT DO ANYTHING USEFUL UNTIL InitializeFromYAML IS CALLED!!
// Q: "Why don't you just make a NewTTP function and encapsulate it in there?"
// A: Due to how this overlaps with cobra/viper, the current design is the most DRY setup
type TTP struct {
	// these fields are all expected to be set by deserialization TTP YAML file
	// via InitializeFromYAML
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Steps       []Step            `yaml:"steps,omitempty,flow"`

	// behaviors we expect to be set via top-level CLI flags/configs, not via
	// deserialization of yaml config
	// Just parse cobra flags/viper configs directly into these fields
	NoCleanup bool   `yaml:"-"`
	WorkDir   string `yaml:"-"`
	// singular as opposed to plural to be like unix $PATH - has multiple elements
	InventoryPath []string `yaml:"-"`

	// set by InitializeFromYAML
	yamlFile string
}

// MarshalYAML is a custom marshalling implementation for the TTP structure. It encodes a TTP object into a formatted
// YAML string, handling the indentation and structure of the output YAML.
//
// Returns:
//
// interface{}: The formatted YAML string representing the TTP object.
//
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

// UnmarshalYAML is a custom unmarshalling implementation for the TTP structure. It decodes a YAML Node into a TTP object,
// handling the decoding and validation of the individual steps within the TTP.
//
// Parameters:
//
// node: A pointer to a yaml.Node that represents the TTP structure to be unmarshalled.
//
// Returns:
//
// error: An error if the decoding process fails or if the TTP structure contains invalid steps.
func (t *TTP) UnmarshalYAML(node *yaml.Node) error {
	type TTPTmpl struct {
		Name        string            `yaml:"name,omitempty"`
		Description string            `yaml:"description"`
		Environment map[string]string `yaml:"env,flow,omitempty"`
		Steps       []yaml.Node       `yaml:"steps,omitempty,flow"`
	}

	var tmpl TTPTmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	t.Name = tmpl.Name
	t.Description = tmpl.Description
	t.Environment = tmpl.Environment

	return t.decodeAndValidateSteps(tmpl.Steps)
}

func (t *TTP) decodeAndValidateSteps(steps []yaml.Node) error {
	for _, stepNode := range steps {
		decoded := false
		// these candidate steps are pointers, so this line
		// MUST be inside the outer step loop or horrible things will happen
		// #justpointerthings
		stepTypes := []Step{NewBasicStep(), NewFileStep(), NewSubTTPStep()}
		for _, stepType := range stepTypes {
			err := stepNode.Decode(stepType)
			if err == nil && !stepType.IsNil() {
				logging.Logger.Sugar().Debugw("decoded step", "step", stepType)
				t.Steps = append(t.Steps, stepType)
				decoded = true
				break
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

// Failed returns a slice of strings containing the names of failed steps in the TTP.
func (t *TTP) Failed() (failed []string) {
	for _, s := range t.Steps {
		if !s.Success() {
			failed = append(failed, s.StepName())
		}
	}
	return failed
}

// fetchEnv retrieves the environment variables and populates the TTP's Environment map.
func (t *TTP) fetchEnv() {
	if t.Environment == nil {
		t.Environment = make(map[string]string)
	}
	logging.Logger.Sugar().Debugw("environment for ttps", "env", t.Environment)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		t.Environment[pair[0]] = pair[1]
	}
}

// ValidateSteps iterates through each step in the TTP and validates it. It sets the working directory
// for each step before calling its Validate method. If any step fails validation, the method returns
// an error. If all steps are successfully validated, the method returns nil.
//
// Returns:
//
// error: An error if any step validation fails, otherwise nil.
func (t *TTP) ValidateSteps() error {
	logging.Logger.Sugar().Info("[*] Validating Steps")

	for _, step := range t.Steps {
		stepCopy := step
		if err := stepCopy.Validate(); err != nil {
			logging.Logger.Sugar().Errorw("failed to validate %s step: %v", step, zap.Error(err))
			return err
		}
	}
	logging.Logger.Sugar().Info("[+] Finished validating steps")
	return nil
}

func (t *TTP) executeSteps() (map[string]Step, []CleanupAct, error) {
	logging.Logger.Sugar().Infof("[+] Running current TTP: %s", t.Name)
	availableSteps := make(map[string]Step)
	var cleanup []CleanupAct

	for _, step := range t.Steps {
		stepCopy := step
		logging.Logger.Sugar().Infof("[+] Running current step: %s", step.StepName())
		stepCopy.Setup(t.Environment, availableSteps)

		if err := stepCopy.Execute(); err != nil {
			logging.Logger.Sugar().Errorw("error encountered in stepCopy execution: %v", err)
			return nil, nil, err
		}
		// Enters in reverse order
		availableSteps[stepCopy.StepName()] = stepCopy
		cleanup = append(stepCopy.GetCleanup(), cleanup...)
		logging.Logger.Sugar().Infof("[+] Finished running step: %s", step.StepName())
	}
	return availableSteps, cleanup, nil
}

// RunSteps executes all of the steps in the given TTP.
//
// Parameters:
//
// t: The TTP to execute the steps for.
//
// Returns:
//
// error: An error if any of the steps fail to execute.
func (t *TTP) RunSteps() error {

	// create temporary working directory if needed
	if t.WorkDir == "" {
		// no pattern, that would be an indicator of compromise :P
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			return err
		}

		// NoCleanup means we want to save TTP results, don't delete them
		if !t.NoCleanup {
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					logging.Logger.Sugar().Errorf("Failed to delete workdir: %v", tmpDir)
				}
			}()
		}
		t.WorkDir = tmpDir
	}
	logging.Logger.Sugar().Infof("[*] Using working directory: %v", t.WorkDir)

	// initialize the working directory for execution
	err := t.copyTTPtoWorkDir()
	if err != nil {
		return err
	}

	// chdir
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Chdir(curDir); err != nil {
			logging.Logger.Sugar().Errorf("failed to revert directory: %v", err)
		}
	}()
	if err := os.Chdir(t.WorkDir); err != nil {
		return err
	}

	// validate
	if err := t.ValidateSteps(); err != nil {
		return err
	}

	t.fetchEnv()

	availableSteps, cleanup, err := t.executeSteps()
	if err != nil {
		return err
	}

	logging.Logger.Sugar().Info("[*] Completed TTP")

	if len(cleanup) > 0 && !t.NoCleanup {
		logging.Logger.Sugar().Info("[*] Beginning Cleanup")
		if err := t.Cleanup(availableSteps, cleanup); err != nil {
			logging.Logger.Sugar().Errorw("error encountered in cleanup step: %v", err)
			return err
		}
		logging.Logger.Sugar().Info("[*] Finished Cleanup")
	} else {
		logging.Logger.Sugar().Info("[*] No Cleanup Steps Found")
	}

	return nil
}

func (t *TTP) executeCleanupSteps(availableSteps map[string]Step, cleanupSteps []CleanupAct) error {
	for _, step := range cleanupSteps {
		stepCopy := step
		logging.Logger.Sugar().Infof("[+] Running current cleanup step: %s", step.CleanupName())
		stepCopy.Setup(t.Environment, availableSteps)

		if err := stepCopy.Cleanup(); err != nil {
			logging.Logger.Sugar().Errorw("error encountered in stepCopy cleanup: %v", err)
			return err
		}
		logging.Logger.Sugar().Infof("[+] Finished running cleanup step: %s", step.CleanupName())
	}
	return nil
}

// Cleanup executes all of the cleanup steps in the given TTP.
//
// Parameters:
//
// t: The TTP to execute the cleanup steps for.
//
// Returns:
//
// error: An error if any of the cleanup steps fail to execute.
func (t *TTP) Cleanup(availableSteps map[string]Step, cleanupSteps []CleanupAct) error {
	err := t.executeCleanupSteps(availableSteps, cleanupSteps)
	if err != nil {
		logging.Logger.Sugar().Errorw("error encountered in cleanup step loop: %v", err)
		return err
	}
	return nil
}

// InitializeFromYAML reads a TTP file initialize the TTP instance based on its contents.
// If the file is empty or contains invalid data, it returns an error.
//
// Parameters:
//
// filename: A string representing the path to the TTP file.
//
// Returns:
//
// err: An error if the file contains invalid data or cannot be read.
func (t *TTP) InitializeFromYAML(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, t)
	if err != nil {
		return
	}

	t.yamlFile = filename

	return
}

// copy TTP config to a temporary working directory
func (t *TTP) copyTTPtoWorkDir() error {

	dir := filepath.Dir(t.yamlFile)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// seriously you need a library for this wtf golang
	err = cp.Copy(absDir, t.WorkDir)
	return err
}
