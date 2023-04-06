package blocks

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Top level struct
type TTP struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Steps       []Step            `yaml:"steps,omitempty,flow"`
	WorkDir     string            `yaml:"-"` // omit but expose for testing and other shenanigans
}

func (t *TTP) UnmarshalYAML(node *yaml.Node) error {

	// Top level struct
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

	for _, stepnode := range tmpl.Steps {
		// we do it piecemiel to build our struct
		basic := NewBasicStep()
		file := NewFileStep()
		subttp := NewSubTTPStep()
		var err, berr, ferr, serr error

		// decoding does not provide method for strict unwrapping so we have to check for validity

		berr = stepnode.Decode(&basic)
		if berr == nil && !basic.IsNil() {
			Logger.Sugar().Debugw("basic", "b", basic)
			t.Steps = append(t.Steps, basic)
			continue
		}

		ferr = stepnode.Decode(&file)

		if ferr == nil && !file.IsNil() {
			Logger.Sugar().Debugw("file", "f", file)
			t.Steps = append(t.Steps, file)
			continue
		}

		serr = stepnode.Decode(&subttp)

		if serr == nil && !subttp.IsNil() {
			Logger.Sugar().Debugw("subttp", "s", subttp)
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

func (t *TTP) Failed() (failed []string) {
	for _, s := range t.Steps {
		if !s.Success() {
			failed = append(failed, s.StepName())
		}
	}
	return failed
}

func (t *TTP) fetchEnv() {
	if t.Environment == nil {
		t.Environment = make(map[string]string)
	}
	Logger.Sugar().Debugw("environment for ttps", "env", t.Environment)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		t.Environment[pair[0]] = pair[1]
	}
}

func (t *TTP) RunSteps() error {

	if t.WorkDir == "" {
		path, err := os.Getwd()
		if err != nil {
			return err
		}

		t.WorkDir = path

	}

	Logger.Sugar().Info("[*] Validating Steps")

	for _, step := range t.Steps {

		stepCopy := step
		// pass in the directory
		stepCopy.SetDir(t.WorkDir)
		err := stepCopy.Validate()
		if err != nil {
			return err
		}
	}

	t.fetchEnv()
	Logger.Sugar().Info("[+] Finished validating steps")

	Logger.Sugar().Infof("[+] Running current TTP: %s", t.Name)
	availableSteps := make(map[string]Step)
	var cleanup []CleanupAct
	var err error

	for _, step := range t.Steps {
		stepCopy := step
		Logger.Sugar().Infof("[+] Running current step: %s", step.StepName())
		stepCopy.Setup(t.Environment, availableSteps)

		err = stepCopy.Execute()
		if err != nil {
			break
		}
		availableSteps[stepCopy.StepName()] = stepCopy
		// Enters in reverse order

		Logger.Sugar().Debugw("step data", "data", stepCopy)
		stepClean := stepCopy.GetCleanup()
		if len(stepClean) > 0 {
			Logger.Sugar().Debugw("adding cleanup step", "cleanup", stepClean)
			cleanup = append(stepClean, cleanup...)
		}
		Logger.Sugar().Debugw("available step data", "data", availableSteps[stepCopy.StepName()].GetOutput())
		Logger.Sugar().Infof("[+] Finished running step: %s", step.StepName())
	}
	// original error from step loop
	if err != nil {
		return err
	}

	Logger.Sugar().Info("[*] Completed TTP")

	if len(cleanup) > 0 {
		Logger.Sugar().Info("[*] Beginning Cleanup")
		t.Cleanup(availableSteps, cleanup)
		Logger.Sugar().Info("[*] Finished Cleanup")
	} else {
		Logger.Sugar().Info("[*] No Cleanup Steps Found")
	}

	return nil
}

func (t *TTP) Cleanup(availableSteps map[string]Step, cleanupSteps []CleanupAct) (err error) {
	for _, step := range cleanupSteps {
		stepCopy := step
		Logger.Sugar().Infof("[+] Running current cleanup step: %s", step.CleanupName())
		stepCopy.Setup(t.Environment, availableSteps)

		err = stepCopy.Cleanup()
		if err != nil {
			break
		}
		// Enters in reverse order
		Logger.Sugar().Infof("[+] Finished running cleanup step: %s", step.CleanupName())

	}
	// original error from step loop
	if err != nil {
		return err
	}
	return nil
}
