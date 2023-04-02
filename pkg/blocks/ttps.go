package blocks

import (
	"fmt"
	"os"
	"strings"

	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TTP represents the top-level structure for a TTP (Tactics, Techniques, and Procedures) object.
type TTP struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Steps       []Step            `yaml:"steps,omitempty,flow"`
	// Omit WorkDir, but expose for testing.
	WorkDir string `yaml:"-"`
}

// UnmarshalYAML custom unmarshalling implementation for the TTP structure.
func (t *TTP) UnmarshalYAML(node *yaml.Node) error {
	// TTPTmpl is a temporary structure to assist in unmarshalling a TTP object.
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

	// Decode and validate each step in the TTP
	for _, stepnode := range tmpl.Steps {
		basic := NewBasicStep()
		file := NewFileStep()
		subttp := NewSubTTPStep()
		var err, berr, ferr, serr error

		berr = stepnode.Decode(&basic)
		if berr == nil && !basic.IsNil() {
			logging.Logger.Sugar().Debugw("basic", "b", basic)
			t.Steps = append(t.Steps, basic)
			continue
		}

		ferr = stepnode.Decode(&file)
		if ferr == nil && !file.IsNil() {
			logging.Logger.Sugar().Debugw("file", "f", file)
			t.Steps = append(t.Steps, file)
			continue
		}

		serr = stepnode.Decode(&subttp)
		if serr == nil && !subttp.IsNil() {
			logging.Logger.Sugar().Debugw("subttp", "s", subttp)
			t.Steps = append(t.Steps, subttp)
			continue
		}

		act := Act{}
		err = stepnode.Decode(&act)
		if act.Name != "" && err != nil {
			return fmt.Errorf("invalid step found, missing parameters for following types: %w, %w, %w", berr, ferr, serr)
		}
		return fmt.Errorf("invalid step found with no name, missing parameters for following types: %w, %w, %w", berr, ferr, serr)
	}

	return nil
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
	if t.WorkDir == "" {
		path, err := os.Getwd()
		if err != nil {
			return err
		}
		t.WorkDir = path
	}

	logging.Logger.Sugar().Info("[*] Validating Steps")

	for _, step := range t.Steps {
		fmt.Println(step)
		stepCopy := step
		// pass in the directory
		stepCopy.SetDir(t.WorkDir)
		if err := stepCopy.Validate(); err != nil {
			logging.Logger.Sugar().Errorw("failed to validate %s step: %v", step, zap.Error(err))
			return err
		}
	}

	t.fetchEnv()
	logging.Logger.Sugar().Info("[+] Finished validating steps")

	logging.Logger.Sugar().Infof("[+] Running current TTP: %s", t.Name)
	availableSteps := make(map[string]Step)
	var cleanup []CleanupAct
	var err error

	for _, step := range t.Steps {
		stepCopy := step
		logging.Logger.Sugar().Infof("[+] Running current step: %s", step.StepName())
		stepCopy.Setup(t.Environment, availableSteps)

		if err := stepCopy.Execute(); err != nil {
			logging.Logger.Sugar().Errorw("error encountered in stepCopy execution: %v", err)
			break
		}
		// Enters in reverse order
		availableSteps[stepCopy.StepName()] = stepCopy

		logging.Logger.Sugar().Debugw("step data", "data", stepCopy)
		stepClean := stepCopy.GetCleanup()
		if len(stepClean) > 0 {
			logging.Logger.Sugar().Debugw("adding cleanup step", "cleanup", stepClean)
			cleanup = append(stepClean, cleanup...)
		}
		logging.Logger.Sugar().Debugw("available step data", "data", availableSteps[stepCopy.StepName()].GetOutput())
		logging.Logger.Sugar().Infof("[+] Finished running step: %s", step.StepName())
	}
	// original error from step loop
	if err != nil {
		logging.Logger.Sugar().Errorw("error encountered in step loop: %v", err)
		return err
	}

	logging.Logger.Sugar().Info("[*] Completed TTP")

	if len(cleanup) > 0 {
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

// Cleanup executes all of the cleanup steps in the given TTP.
//
// Parameters:
//
// t: The TTP to execute the cleanup steps for.
//
// Returns:
//
// error: An error if any of the cleanup steps fail to execute.
func (t *TTP) Cleanup(availableSteps map[string]Step, cleanupSteps []CleanupAct) (err error) {
	for _, step := range cleanupSteps {
		stepCopy := step
		logging.Logger.Sugar().Infof("[+] Running current cleanup step: %s", step.CleanupName())
		stepCopy.Setup(t.Environment, availableSteps)

		if err := stepCopy.Cleanup(); err != nil {
			logging.Logger.Sugar().Errorw("error encountered in stepCopy cleanup: %v", err)
			break
		}
		// Enters in reverse order
		logging.Logger.Sugar().Infof("[+] Finished running cleanup step: %s", step.CleanupName())

	}

	// original error from cleanup step loop
	if err != nil {
		logging.Logger.Sugar().Errorw("error encountered in cleanup step loop: %v", err)
		return err
	}

	return nil
}
