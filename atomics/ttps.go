package atomics

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Top level struct
type TTP struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Steps       []*Step           `yaml:"steps,omitempty,flow"`
	altBaseDir  string
}

func (a *TTP) Failed() (failed []string) {
	for _, s := range a.Steps {
		if !s.Success {
			failed = append(failed, s.Name)
		}
	}
	return failed
}

func (a *TTP) SetEmbedHome(basename string) {
	if a.Environment == nil {
		a.Environment = make(map[string]string)
	}
	a.Environment["EMBEDHOME"] = basename
}

func (a *TTP) fetchEnv() {
	if a.Environment == nil {
		a.Environment = make(map[string]string)
	}
	Logger.Sugar().Debugw("environment for ttps", "env", a.Environment)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		a.Environment[pair[0]] = pair[1]
	}
}

func (a *TTP) RunSteps() (string, error) {
	Logger.Sugar().Info("[*] Validating Steps")
	err := a.checkSteps()
	if err != nil {
		return "", err
	}

	a.fetchEnv()

	Logger.Sugar().Infof("[+] Running current TTP: %s", a.Name)
	availableSteps := make(map[string]*Action)

	lastCompletedStepIdx := -1
	for stepIdx, step := range a.Steps {
		stepCopy := step.Action
		Logger.Sugar().Infof("[+] Running current step: %s", step.Name)
		stepCopy.setupStep(a.Environment, availableSteps)

		err = stepCopy.Exec()
		if err != nil {
			break
		}
		availableSteps[stepCopy.Name] = &stepCopy
		Logger.Sugar().Debugw("available step data", "data", availableSteps[stepCopy.Name].JsonOutput)
		Logger.Sugar().Infof("[+] Finished running step: %s", step.Name)
		lastCompletedStepIdx = stepIdx
	}
	// run cleanup even if there was an error
	Logger.Sugar().Info("========= Cleanup =========")
	for cleanupIdx := lastCompletedStepIdx; cleanupIdx >= 0; cleanupIdx-- {
		cleanup := a.Steps[cleanupIdx].Cleanup
		if cleanup == nil {
			continue
		}

		cleanup.setupEnv(a.Environment)
		err = cleanup.Exec()
		if err != nil {
			return "", fmt.Errorf("cleanup failed: %v", err)
		}
	}
	// original error from step loop
	if err != nil {
		return "", err
	}
	Logger.Sugar().Info("[*] Completed TTP")
	return "", nil
}

func (a *TTP) checkSteps() error {
	names := make(map[string]bool)
	for _, step := range a.Steps {
		stepCopy := step
		stepCopy.altBaseDir = a.altBaseDir
		Logger.Sugar().Infof("[*] Validating current step: %s", stepCopy.Name)
		if _, ok := names[step.Name]; ok {
			// name exists, conflict
			return errors.New(fmt.Sprintf("duplicate name used in steps %s", stepCopy.Name))
		}
		names[stepCopy.Name] = true

		err := stepCopy.validate()
		if err != nil {
			return err
		}
		if stepCopy.Cleanup != nil {
			stepCopy.Cleanup.validate()
		}
		Logger.Sugar().Infof("[*] Validated step: %s", stepCopy.Name)
	}
	return nil
}

// func parseSteps() {
// 	for _, ttp := range ttps {
// 		for _, step := range ttp.Steps {
// 			if _, err := os.Stat(step.FilePath); errors.Is(err, os.ErrNotExist) {
// 				Logger.Sugar().Warnf("file does not exist ttp: %s", ttp.TTP)
// 			}
// 		}
// 	}
// }
