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
	"fmt"
	"os"
	"strconv"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/processutils"
)

// KillProcessConfig holds the configuration for the kill_process action
type KillProcessConfig struct {
	ID   string `yaml:"id,omitempty"`
	Name string `yaml:"name,omitempty"`
}

// KillProcessStep kills a process using ID/name
// Its intended use is simulating malicious programs stopping
// critical applications/processes
type KillProcessStep struct {
	actionDefaults            `yaml:",inline"`
	KillProcess               KillProcessConfig `yaml:"kill_process,omitempty"`
	ErrorOnFindProcessFailure bool              `yaml:"error_on_find_process_failure,omitempty"`
	ErrorOnKillFailure        bool              `yaml:"error_on_kill_failure,omitempty"`
}

// NewKillProcessStep creates a new KillProcessStep instance and returns a pointer to it.
func NewKillProcessStep() *KillProcessStep {
	return &KillProcessStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *KillProcessStep) IsNil() bool {
	switch {
	case s.KillProcess.ID == "" && s.KillProcess.Name == "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies.
func (s *KillProcessStep) Validate(_ TTPExecutionContext) error {
	if s.KillProcess.ID == "" && s.KillProcess.Name == "" {
		return fmt.Errorf("either id or name must be provided")
	}
	if s.KillProcess.ID != "" && s.KillProcess.Name != "" {
		return fmt.Errorf("only one of id or name can be provided, not both")
	}
	if s.KillProcess.ID != "" {
		processID, err := strconv.Atoi(s.KillProcess.ID)
		if err != nil {
			return fmt.Errorf("invalid process id: %v", s.KillProcess.ID)
		}
		if processID <= 0 {
			return fmt.Errorf("process id must be greater than 0")
		}
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *KillProcessStep) Template(execCtx TTPExecutionContext) error {
	var err error
	s.KillProcess.ID, err = execCtx.templateStep(s.KillProcess.ID)
	if err != nil {
		return err
	}
	s.KillProcess.Name, err = execCtx.templateStep(s.KillProcess.Name)
	if err != nil {
		return err
	}
	return nil
}

// extractPIDs - extracts process IDs based on the user input.
// Returns a slice of process IDs to be killed.
func (s *KillProcessStep) extractPIDs() ([]int, error) {
	if s.KillProcess.ID != "" {
		processID, _ := strconv.Atoi(s.KillProcess.ID)
		logging.L().Infof("Using Process ID: %v", processID)

		err := processutils.VerifyPIDExists(processID)
		if err != nil {
			logging.L().Errorf("Error while trying to verify PID exists: %+v", err)
			if s.ErrorOnFindProcessFailure {
				return nil, err
			}
			return []int{}, nil
		}
		return []int{processID}, nil
	}

	// s.KillProcess.Name must be set (validated in Validate())
	logging.L().Infof("Finding processes with name: %v", s.KillProcess.Name)

	processes, err := processutils.GetPIDsByName(s.KillProcess.Name)
	if err != nil {
		logging.L().Errorf("Error while trying to get PIDs from name: %+v", err)
		if s.ErrorOnFindProcessFailure {
			return nil, err
		}
		return []int{}, nil
	}

	pids := make([]int, len(processes))
	for i, pid := range processes {
		pids[i] = int(pid)
	}

	return pids, nil
}

// killProcesses - kills all processes with the given process IDs.
// Logs successful and unsuccessful kill actions.
// Returns an error if there's only one process to kill and it fails.
func (s *KillProcessStep) killProcesses(pids []int) error {
	logging.L().Infof("Killing the following processes: %v", pids)

	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			logging.L().Errorf("Error while trying to find process with ID: %v; %+v", pid, err)
			if s.ErrorOnFindProcessFailure {
				return err
			}
			continue
		}

		logging.L().Infof("Got process handle with PID: %d", pid)
		if err := proc.Kill(); err != nil {
			logging.L().Errorf("Failed to kill process with ID: %v; %+v", pid, err)
			if s.ErrorOnKillFailure {
				return err
			}
			continue
		}

		logging.L().Infof("Killed process with ID: %d", pid)
	}

	return nil
}

// Execute runs the step and returns an error if one occurs while extracting PIDs or killing processes.
func (s *KillProcessStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	if execCtx.Backend != nil {
		return s.executeRemote(execCtx)
	}

	pids, err := s.extractPIDs()
	if err != nil {
		return nil, err
	}

	if len(pids) == 0 {
		logging.L().Infof("No processes found to kill")
		return &ActResult{}, nil
	}
	if err := s.killProcesses(pids); err != nil {
		return nil, err
	}

	return &ActResult{}, nil
}

// executeRemote handles process killing via the execution backend.
func (s *KillProcessStep) executeRemote(execCtx TTPExecutionContext) (*ActResult, error) {
	backend := execCtx.Backend

	var pids []int

	if s.KillProcess.ID != "" {
		processID, _ := strconv.Atoi(s.KillProcess.ID)
		exists, err := backend.ProcessExists(processID)
		if err != nil || !exists {
			if s.ErrorOnFindProcessFailure {
				return nil, fmt.Errorf("process %d not found on remote host", processID)
			}
			logging.L().Infof("No processes found to kill")
			return &ActResult{}, nil
		}
		pids = []int{processID}
	} else {
		// s.KillProcess.Name must be set (validated in Validate())
		var err error
		pids, err = backend.FindProcessesByName(s.KillProcess.Name)
		if err != nil {
			if s.ErrorOnFindProcessFailure {
				return nil, err
			}
			logging.L().Infof("No processes found to kill")
			return &ActResult{}, nil
		}
	}

	for _, pid := range pids {
		if err := backend.KillProcess(pid); err != nil {
			logging.L().Errorf("Failed to kill remote process %d: %v", pid, err)
			if s.ErrorOnKillFailure {
				return nil, err
			}
			continue
		}
		logging.L().Infof("Killed remote process with ID: %d", pid)
	}

	return &ActResult{}, nil
}
