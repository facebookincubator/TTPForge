/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
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
	"context"
	"errors"
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/backends"
	"github.com/facebookincubator/ttpforge/pkg/checks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// CommonStepFields contains the fields
// common to every type of step (such as Name).
// It centralizes validation to simplify the code
type CommonStepFields struct {
	Name   string         `yaml:"name,omitempty"`
	Remote string         `yaml:"remote,omitempty"`
	Checks []checks.Check `yaml:"checks,omitempty"`

	// CleanupSpec is exported so that UnmarshalYAML
	// can see it - however, it should be considered
	// to be a private detail of this file
	// and not referenced elsewhere in the codebase
	CleanupSpec yaml.Node `yaml:"cleanup,omitempty"`
}

// Step contains a TTPForge executable action
// and its associated cleanup action (if specified)
type Step struct {
	CommonStepFields

	// These are where the actual executable content
	// of the step (and its associated cleanup process)
	// live - they are not deserialized directly from YAML
	// but rather must be decoded by ParseAction
	action  Action
	cleanup Action

	// cleanupRemote is the remote override for custom cleanup actions.
	// When set, the cleanup action runs on this named connection
	// instead of locally.
	cleanupRemote string

	// isDefaultCleanup is true when cleanup: default was used,
	// meaning the cleanup should inherit the step's remote: field.
	isDefaultCleanup bool
}

func isDefaultCleanup(cleanupNode *yaml.Node) (bool, error) {
	var testStr string
	// is it a string? if not, let the subsequent decoding
	// in the calling function deal with it
	if err := cleanupNode.Decode(&testStr); err != nil {
		return false, nil
	}

	// if it is a string, it must be a valid string
	if testStr == "default" {
		return true, nil
	}
	return false, fmt.Errorf("invalid cleanup value specified: %v", testStr)
}

// ShouldCleanupOnFailure specifies that this step should be cleaned
// up even if its Execute(...)  failed.
// We usually don't want to do this - for example,
// you shouldn't try to remove_path a create_file that failed)
// However, certain step types (especially SubTTPs) need to run cleanup even if they fail
func (s *Step) ShouldCleanupOnFailure() bool {
	switch s.action.(type) {
	case *SubTTPStep:
		return true
	default:
		return false
	}
}

// ShouldUseImplicitDefaultCleanup is a hack
// to make subTTPs always run their default
// cleanup process even when `cleanup: default` is
// not explicitly specified - this is purely for backward
// compatibility
func ShouldUseImplicitDefaultCleanup(action Action) bool {
	switch action.(type) {
	case *SubTTPStep:
		return true
	default:
		return false
	}
}

// UnmarshalYAML implements custom deserialization
// process to ensure that the step action and its
// cleanup action are decoded to the correct struct type
func (s *Step) UnmarshalYAML(node *yaml.Node) error {

	// Decode all of the shared fields.
	// Use of this auxiliary type prevents infinite recursion
	var csf CommonStepFields
	err := node.Decode(&csf)
	if err != nil {
		return err
	}
	s.CommonStepFields = csf

	if s.Name == "" {
		return errors.New("no name specified for step")
	}

	// figure out what kind of action is
	// associated with executing this step
	s.action, err = s.ParseAction(node)
	if err != nil {
		return fmt.Errorf("could not parse action for step %q: %w", s.Name, err)
	}

	// figure out what kind of action is
	// associated with cleaning up this step
	if csf.CleanupSpec.IsZero() {
		// hack for subTTPs - they should always use their default cleanup
		if ShouldUseImplicitDefaultCleanup(s.action) {
			s.cleanup = s.action.GetDefaultCleanupAction()
			s.isDefaultCleanup = true
		}
	} else {
		useDefaultCleanup, err := isDefaultCleanup(&csf.CleanupSpec)
		if err != nil {
			return err
		}
		if useDefaultCleanup {
			s.isDefaultCleanup = true
			if dca := s.action.GetDefaultCleanupAction(); dca != nil {
				s.cleanup = dca
				return nil
			}
			return fmt.Errorf("`cleanup: default` was specified but step %v is not an action type that has a default cleanup action", s.Name)
		}

		// Extract remote: from custom cleanup mapping before parsing the action
		var cleanupMeta struct {
			Remote string `yaml:"remote,omitempty"`
		}
		if err := csf.CleanupSpec.Decode(&cleanupMeta); err == nil {
			s.cleanupRemote = cleanupMeta.Remote
		}

		s.cleanup, err = s.ParseAction(&csf.CleanupSpec)
		if err != nil {
			return fmt.Errorf("could not parse cleanup action for step %q: %w", s.Name, err)
		}
	}
	return nil
}

// Validate checks that both the step action and cleanup
// action are valid
func (s *Step) Validate(execCtx TTPExecutionContext) error {
	if err := s.action.Validate(execCtx); err != nil {
		return err
	}
	if s.cleanup != nil {
		if err := s.cleanup.Validate(execCtx); err != nil {
			return err
		}
	}
	return nil
}

// Template replaces variables in the step action
func (s *Step) Template(execCtx TTPExecutionContext) error {
	if err := s.action.Template(execCtx); err != nil {
		return err
	}
	return nil
}

// swapToRemote sets up the backend on the execution context for the given
// remote connection name. It returns a restore function that must be called
// to restore the original backend. If remoteName is empty, it is a no-op.
//
// When switching to a remote backend, the working directory is reset to "/"
// because the local working directory (typically the TTP file's directory)
// will not exist on the remote host.
func (s *Step) swapToRemote(execCtx *TTPExecutionContext, remoteName string) (func(), error) {
	if remoteName == "" {
		return func() {}, nil
	}
	if execCtx.ConnPool == nil {
		return nil, fmt.Errorf("remote: specified on step %q but connection pool is not initialized", s.Name)
	}

	originalBackend := execCtx.Backend
	originalWorkDir := execCtx.Vars.WorkDir
	backend, err := execCtx.ConnPool.GetByName(remoteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get backend for step %q: %w", s.Name, err)
	}
	execCtx.Backend = backend
	// Reset working directory for remote execution — the local TTP directory
	// path does not exist on the remote host.
	execCtx.Vars.WorkDir = "/"
	return func() {
		execCtx.Backend = originalBackend
		execCtx.Vars.WorkDir = originalWorkDir
	}, nil
}

// swapBackend sets up the backend on the execution context if this step
// has a remote: connection name. It returns a restore function that must
// be called to restore the original backend.
func (s *Step) swapBackend(execCtx *TTPExecutionContext) (func(), error) {
	return s.swapToRemote(execCtx, s.Remote)
}

// Execute runs the action associated with this step and sends result/error to channels of the context
func (s *Step) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	// Swap backend if remote: is specified
	restore, err := s.swapBackend(&execCtx)
	if err != nil {
		execCtx.errorsChan <- err
		return nil, err
	}
	defer restore()

	desc := s.action.GetDescription()
	if desc != "" {
		logging.L().Infof("Description: %v", desc)
	}
	result, err := s.action.Execute(execCtx)
	if err != nil {
		logging.L().Errorf("Failed to execute step %v: %v", s.Name, err)
		execCtx.errorsChan <- err
	} else {
		logging.L().Debugf("Successfully executed step %v", s.Name)
		execCtx.actionResultsChan <- result
	}

	return result, err
}

// Cleanup runs the cleanup action associated with this step
func (s *Step) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	if s.cleanup != nil {
		// Determine remote target for cleanup:
		// - cleanup: default → inherit step's remote (same host)
		// - custom cleanup → use cleanup's own remote: (local if not specified)
		remoteForCleanup := ""
		if s.isDefaultCleanup {
			remoteForCleanup = s.Remote
		} else {
			remoteForCleanup = s.cleanupRemote
		}

		restore, err := s.swapToRemote(&execCtx, remoteForCleanup)
		if err != nil {
			return nil, err
		}
		defer restore()

		desc := s.cleanup.GetDescription()
		if desc != "" {
			logging.L().Infof("Description: %v", desc)
		}
		if err := s.cleanup.Template(execCtx); err != nil {
			return nil, err
		}
		return s.cleanup.Execute(execCtx)
	}
	logging.L().Infof("No Cleanup Action Defined for Step %v", s.Name)
	return &ActResult{}, nil
}

// ParseAction decodes an action (from step or cleanup) in YAML
// format into the appropriate struct
func (s *Step) ParseAction(node *yaml.Node) (Action, error) {
	var typeField struct {
		Inline    string       `yaml:"inline"`
		File      string       `yaml:"file"`
		TTP       string       `yaml:"ttp"`
		EditFile  string       `yaml:"edit_file"`
		Responses []Response   `yaml:"responses"`
		Connect   *ConnectStep `yaml:"connect"`
	}

	if err := node.Decode(&typeField); err != nil {
		return nil, err
	}

	// Check for ambiguous types
	typesCount := 0
	if typeField.Inline != "" {
		typesCount++
	}
	if typeField.File != "" {
		typesCount++
	}
	if typeField.TTP != "" {
		typesCount++
	}
	if typeField.EditFile != "" {
		typesCount++
	}
	if typeField.Connect != nil && !typeField.Connect.IsNil() {
		typesCount++
	}
	if typesCount > 1 {
		return nil, fmt.Errorf("step %v has ambiguous type", s.Name)
	}

	// Check for ConnectStep
	if typeField.Connect != nil && !typeField.Connect.IsNil() {
		return typeField.Connect, nil
	}

	// Check for ExpectStep
	if len(typeField.Responses) > 0 {
		expectStep := NewExpectStep()
		if err := node.Decode(expectStep); err != nil {
			return nil, err
		}
		return expectStep, nil
	}

	// Otherwise, treat it as a BasicStep
	if typeField.Inline != "" {
		basicStep := NewBasicStep()
		if err := node.Decode(basicStep); err != nil {
			return nil, err
		}
		return basicStep, nil
	}

	actionCandidates := []Action{
		NewBasicStep(),
		NewChangeDirectoryStep(),
		NewFileStep(),
		NewSubTTPStep(),
		NewEditStep(),
		NewFetchURIStep(),
		NewCreateFileStep(),
		NewCopyPathStep(),
		NewRemovePathAction(),
		NewPrintStrAction(),
		NewExpectStep(),
		NewHTTPRequestStep(),
		NewKillProcessStep(),
	}

	var action Action
	for _, actionType := range actionCandidates {
		err := node.Decode(actionType)
		if err == nil && !actionType.IsNil() {
			if action != nil {
				// Must catch bad steps with ambiguous types, such as:
				// - name: hello
				//   file: bar
				//   ttp: foo
				//
				// we can't use KnownFields to solve this without a massive
				// refactor due to https://github.com/go-yaml/yaml/issues/460
				// note: we check for non-empty name earlier so s.Name will be non-empty
				return nil, fmt.Errorf("step %v has ambiguous type", s.Name)
			}
			action = actionType
		}
	}

	if action == nil {
		return nil, errors.New("action fields did not match any valid action type")
	}

	return action, nil
}

// buildVerificationContext creates a VerificationContext for the given remote
// connection name. If remoteName is empty, the context targets the local machine.
// stepOutput is the combined stdout+stderr from the step that just ran.
func (s *Step) buildVerificationContext(execCtx TTPExecutionContext, remoteName string, stepOutput string) (checks.VerificationContext, error) {
	var activeBackend backends.ExecutionBackend
	if remoteName != "" && execCtx.ConnPool != nil {
		var err error
		activeBackend, err = execCtx.ConnPool.GetByName(remoteName)
		if err != nil {
			return checks.VerificationContext{}, fmt.Errorf("failed to get backend for check on step %q: %w", s.Name, err)
		}
	} else if remoteName == "" {
		activeBackend = execCtx.Backend
	}

	var fsys afero.Fs
	if activeBackend != nil {
		var err error
		fsys, err = activeBackend.GetFs()
		if err != nil {
			return checks.VerificationContext{}, fmt.Errorf("failed to get filesystem for checks: %w", err)
		}
	} else {
		fsys = afero.NewOsFs()
	}

	verificationCtx := checks.VerificationContext{
		FileSystem: fsys,
		StepOutput: stepOutput,
	}

	if activeBackend != nil {
		shellName := "sh"
		shellArgs := []string{"-c"}
		if remoteName != "" && execCtx.ConnPool != nil {
			if remoteCfg := execCtx.ConnPool.GetConfigByName(remoteName); remoteCfg != nil {
				switch remoteCfg.Shell {
				case "powershell":
					shellName = "powershell"
					shellArgs = []string{"-Command"}
				case "cmd":
					shellName = "cmd"
					shellArgs = []string{"/c"}
				}
			}
		}
		verificationCtx.RunCommand = func(command string) (string, int, error) {
			ctx := context.Background()
			args := append(shellArgs, command)
			stdout, stderr, err := activeBackend.RunCommand(ctx, shellName, "", args, nil, "", nil, nil)
			output := stdout + stderr
			if err != nil {
				return output, 1, nil //nolint:nilerr // non-zero exit is not a check error
			}
			return output, 0, nil
		}
	}

	return verificationCtx, nil
}

// VerifyChecks runs all checks and returns an error if any of them fail.
// result is the ActResult from the step execution; it may be nil.
func (s *Step) VerifyChecks(execCtx TTPExecutionContext, result *ActResult) error {
	if len(s.Checks) == 0 {
		logging.L().Debugf("No checks defined for step %v", s.Name)
		return nil
	}

	var stepOutput string
	if result != nil {
		stepOutput = result.Stdout + result.Stderr
	}

	for checkIdx, check := range s.Checks {
		// Resolve the effective remote for this check
		checkRemote := check.GetRemote()
		switch checkRemote {
		case "":
			// No per-check override → inherit the step's remote
			checkRemote = s.Remote
		case "local":
			// Explicit "local" → force local execution
			checkRemote = ""
		}

		verificationCtx, err := s.buildVerificationContext(execCtx, checkRemote, stepOutput)
		if err != nil {
			return fmt.Errorf("success check %d of step %q setup failed: %w", checkIdx+1, s.Name, err)
		}

		if err := check.Verify(verificationCtx); err != nil {
			return fmt.Errorf("success check %d of step %q failed: %w", checkIdx+1, s.Name, err)
		}
		logging.L().Debugf("Success check %d (%q) of step %q PASSED", checkIdx+1, check.Msg, s.Name)
	}
	return nil
}
